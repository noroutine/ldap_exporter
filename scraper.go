package ldap_exporter

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/ldap.v2"
)

const (
	monitorBaseDN    = "cn=Monitor"
	operationsBaseDN = "cn=Operations,cn=Monitor"

	monitorCounterObject = "monitorCounterObject"
	monitorCounter       = "monitorCounter"

	monitoredObject = "monitoredObject"
	monitoredInfo   = "monitoredInfo"

	monitorOperation   = "monitorOperation"
	monitorOpCompleted = "monitorOpCompleted"

	posixAccount = "posixAccount"
	uid = "uid"

	SchemeLDAPS = "ldaps"
	SchemeLDAP  = "ldap"
	SchemeLDAPI = "ldapi"
)

type query struct {
	baseDN              string
	searchFilter        string
	searchAttr          string
	ignoreErrors        bool
	metric              *prometheus.GaugeVec
	countMetric         *prometheus.GaugeVec
	queryDurationMetric *prometheus.GaugeVec
}

type LDAPConfig struct {
	UseTLS      bool
	UseStartTLS bool
	Scheme      string
	Addr        string
	Host        string
	Port        string
	Protocol    string
	Username    string
	Password    string
	BaseDN      string
	TLSConfig   tls.Config
}

func (config *LDAPConfig) ParseAddr(addr string) error {

	var u *url.URL

	u, err := url.Parse(addr)
	if err != nil {
		// Well, so far the easy way....
		u = &url.URL{}
	}

	if u.Host == "" {

		if strings.HasPrefix(addr, SchemeLDAPI) {
			u.Scheme = SchemeLDAPI
			u.Host, _ = url.QueryUnescape(strings.Replace(addr, SchemeLDAPI+"://", "", 1))
		} else if strings.HasPrefix(addr, SchemeLDAPS) {
			u.Scheme = SchemeLDAPS
			u.Host = strings.Replace(addr, SchemeLDAPS+"://", "", 1)
		} else {
			u.Scheme = SchemeLDAP
			u.Host = strings.Replace(addr, SchemeLDAP+"://", "", 1)
		}

	}

	config.Addr = u.Host
	config.Scheme = u.Scheme
	config.Host = u.Hostname()

	if u.Scheme == SchemeLDAPS {
		config.UseTLS = true
	} else if u.Scheme == SchemeLDAP {
		config.UseTLS = false
	} else if u.Scheme == SchemeLDAPI {
		config.Protocol = "unix"
	} else {
		return errors.New(u.Scheme + " is not a scheme i understand, refusing to continue")
	}

	return nil

}

func (config *LDAPConfig) LoadCACert(cafile string) error {

	if _, err := os.Stat(cafile); os.IsNotExist(err) {
		return errors.New("CA Certificate file does not exists")
	}

	cert, err := ioutil.ReadFile(cafile)

	if err != nil {
		return errors.New("CA Certificate file is not readable")
	}

	config.TLSConfig.RootCAs = x509.NewCertPool()
	config.TLSConfig.ServerName = config.Host

	ok := config.TLSConfig.RootCAs.AppendCertsFromPEM(cert)

	if ok == false {
		return errors.New("Could not parse CA")
	}

	return nil

}

func NewLDAPConfig() LDAPConfig {

	conf := LDAPConfig{}

	conf.Scheme = SchemeLDAP
	conf.Host = "localhost"
	conf.Port = "389"
	conf.Addr = conf.Host + ":" + conf.Port
	conf.Protocol = "tcp"
	conf.UseTLS = false
	conf.UseStartTLS = false
	conf.TLSConfig = tls.Config{}

	return conf

}

var (
	monitoredObjectGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap",
			Name:      "monitored_object",
			Help:      monitorBaseDN + " " + objectClass(monitoredObject) + " " + monitoredInfo,
		},
		[]string{"dn"},
	)
	monitorCounterObjectGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap",
			Name:      "monitor_counter_object",
			Help:      monitorBaseDN + " " + objectClass(monitorCounterObject) + " " + monitorCounter,
		},
		[]string{"dn"},
	)
	monitorOperationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap",
			Name:      "monitor_operation",
			Help:      operationsBaseDN + " " + objectClass(monitorOperation) + " " + monitorOpCompleted,
		},
		[]string{"dn"},
	)
	posixAccountCountGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap",
			Name:      "posix_account_count",
			Help:      objectClass(posixAccount) + " count",
		},
		[]string{"dn"},
	)
	posixAccountQueryDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: "ldap",
			Name:      "posix_account_query_duration",
			Help:      objectClass(posixAccount) + " query duration",
		},
		[]string{"dn"},
	)
	scrapeCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "ldap",
			Name:      "scrape",
			Help:      "successful vs unsuccessful ldap scrape attempts",
		},
		[]string{"result"},
	)
)

func init() {
	prometheus.MustRegister(
		monitoredObjectGauge,
		monitorCounterObjectGauge,
		monitorOperationGauge,
		posixAccountCountGauge,
		posixAccountQueryDurationGauge,
		scrapeCounter,
	)
}

func objectClass(name string) string {
	return fmt.Sprintf("(objectClass=%v)", name)
}

func ScrapeMetrics(config *LDAPConfig) {
	if err := scrapeAll(config); err != nil {
		resultLabelValue := "fail"
		if ldapErr, ok := err.(*ldap.Error); ok {
			switch ldapErr.ResultCode {
			case ldap.LDAPResultTimeLimitExceeded: resultLabelValue = "timeout"
			}
		}
		scrapeCounter.WithLabelValues(resultLabelValue).Inc()
		log.Println("scrape failed, error is:", err)
	} else {
		scrapeCounter.WithLabelValues("ok").Inc()
	}
}

func scrapeAll(config *LDAPConfig) error {

	var l *ldap.Conn
	var err error

	var queries = []query{
		{
			baseDN:       monitorBaseDN,
			searchFilter: objectClass(monitoredObject),
			searchAttr:   monitoredInfo,
			metric:       monitoredObjectGauge,
			ignoreErrors: true,
		},
		{
			baseDN:       monitorBaseDN,
			searchFilter: objectClass(monitorCounterObject),
			searchAttr:   monitorCounter,
			metric:       monitorCounterObjectGauge,
			ignoreErrors: true,
		},
		{
			baseDN:       operationsBaseDN,
			searchFilter: objectClass(monitorOperation),
			searchAttr:   monitorOpCompleted,
			metric:       monitorOperationGauge,
			ignoreErrors: true,
		},
		{
			baseDN:              config.BaseDN,
			searchFilter:        objectClass(posixAccount),
			searchAttr:          uid,
			countMetric:         posixAccountCountGauge,
			queryDurationMetric: posixAccountQueryDurationGauge,
			ignoreErrors:        false,
		},
	}

	defer func() {
		if err != nil {
			posixAccountQueryDurationGauge.WithLabelValues(config.BaseDN).Set(0)
		}
	}()

	if config.UseTLS {
		l, err = ldap.DialTLS(config.Protocol, config.Addr, &config.TLSConfig)
	} else {
		l, err = ldap.Dial(config.Protocol, config.Addr)
		if err != nil {
			return err
		}
		if config.UseStartTLS {
			err = l.StartTLS(&config.TLSConfig)
			if err != nil {
				return err
			}
		}
	}

	if err != nil {
		return err
	}
	defer l.Close()

	if config.Username != "" && config.Password != "" {
		err = l.Bind(config.Username, config.Password)
		if err != nil {
			return err
		}
	}

	var errs error
	for _, q := range queries {
		if err := scrapeQuery(l, &q); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func scrapeQuery(l *ldap.Conn, q *query) error {
	var queryErr error

	req := ldap.NewSearchRequest(
		q.baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 1, false,
		q.searchFilter, []string{"dn", q.searchAttr}, nil,
	)

	queryStart := time.Now()
	defer func() {
		queryDuration := time.Now().Sub(queryStart)
		if q.queryDurationMetric != nil {
			q.queryDurationMetric.WithLabelValues(q.baseDN).Set(float64(queryDuration.Nanoseconds()))
		}
	}()

	sr, queryErr := l.Search(req)
	if queryErr != nil {
		if q.ignoreErrors {
			return nil
		} else {
			return queryErr
		}
	}

	if q.countMetric != nil {
		q.countMetric.WithLabelValues(q.baseDN).Set(float64(len(sr.Entries)))
	}

	if q.metric != nil {
		for _, entry := range sr.Entries {
			val := entry.GetAttributeValue(q.searchAttr)
			if val == "" {
				// not every entry will have this attribute
				continue
			}
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				// some of these attributes are not numbers
				continue
			}
			q.metric.WithLabelValues(entry.DN).Set(num)
		}
	}
	return nil
}

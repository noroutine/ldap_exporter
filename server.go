package ldap_exporter

import (
	"fmt"
	"nrtn.io/ldap_exporter/app/build"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServerConfig struct {
	Address  string
	CertFile string
	KeyFile  string
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{}
}

func StartMetricsServer(config *ServerConfig) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/version", showVersion)

	var err error

	if config.CertFile != "" && config.KeyFile != "" {
		err = http.ListenAndServeTLS(config.Address, config.CertFile, config.KeyFile, mux)
	} else {
		err = http.ListenAndServe(config.Address, mux)
	}

	if err != nil {
		log.Fatal("http listener failed, error is:", err)
	}
}

func showVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, build.ShortVersionString())
}

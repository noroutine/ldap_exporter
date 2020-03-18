package openldap_exporter

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version string

type ServerConfig struct {
	Address  string
	CertFile string
	KeyFile  string
}

func GetVersion() string {
	return version
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
	_, _ = fmt.Fprintln(w, version)
}

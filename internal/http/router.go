package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	Config struct {
		Port int `env-required:"true" json:"port" env:"HTTP_PORT"`
	}
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

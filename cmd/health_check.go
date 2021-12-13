package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// HealthEndpointHandler will return 204 for requests to '/health'
func HealthEndpointHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusNoContent)
		log.Trace("liveness probe...")
	}
}

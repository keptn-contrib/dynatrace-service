package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// HealthEndpointHandler will return 204 for requests
func HealthEndpointHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	log.Trace("alive...")
}

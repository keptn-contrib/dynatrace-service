package http_handler

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ReadinessEndpointHandler will return 204 for requests
func ReadinessEndpointHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	log.Trace("ready...")
}

package health

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// readinessEndpointHandler will return 204 for requests
func readinessEndpointHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	log.Trace("ready...")
}

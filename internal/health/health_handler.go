package health

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// healthEndpointHandler will return 204 for requests
func healthEndpointHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	log.Trace("alive...")
}

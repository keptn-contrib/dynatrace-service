package health

import (
	"net/http"
)

// HTTPGetHandler will handle '/ready' requests.
func HTTPGetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ready":
		readinessEndpointHandler(w, r)
	default:
		endpointNotFoundHandler(w, r)
	}
}

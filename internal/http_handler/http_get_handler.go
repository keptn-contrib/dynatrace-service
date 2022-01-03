package http_handler

import (
	"net/http"
)

// HTTPGetHandler will handle all requests for '/health' and '/ready'
func HTTPGetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health":
		healthEndpointHandler(w, r)
	case "/ready":
		readinessEndpointHandler(w, r)
	default:
		endpointNotFoundHandler(w, r)
	}
}

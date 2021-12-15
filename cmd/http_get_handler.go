package main

import (
	"net/http"
)

// HTTPGetHandler will handle all requests for '/health' and '/ready'
func HTTPGetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health":
		HealthEndpointHandler(w, r)
	case "/ready":
		ReadinessEndpointHandler(w, r)
	default:
		EndpointNotFoundHandler(w, r)
	}
}

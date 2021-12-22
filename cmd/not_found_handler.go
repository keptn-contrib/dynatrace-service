package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type notFoundError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// EndpointNotFoundHandler will return 404 for requests
func EndpointNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := json.Marshal(
		notFoundError{
			Status:  404,
			Message: fmt.Sprintf("'%s' not found", r.URL.Path),
		})
	if err != nil {
		log.Error("could not marshal error to JSON")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, err = w.Write(payload)
	if err != nil {
		log.Error("could not write payload to response")
	}

	log.Trace("not found...")
}

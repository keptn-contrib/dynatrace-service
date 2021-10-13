package test

import (
	"net/http"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

type errConfigForPayload struct {
	status  int
	payload []byte
}

type PayloadBasedURLHandler struct {
	exactURLs           map[string][]byte
	exactErrorURLs      map[string]errConfigForPayload
	startsWithURLs      map[string][]byte
	startsWithErrorURLs map[string]errConfigForPayload
	t                   *testing.T
}

func NewPayloadBasedURLHandler(t *testing.T) *PayloadBasedURLHandler {
	return &PayloadBasedURLHandler{
		exactURLs:           make(map[string][]byte),
		exactErrorURLs:      make(map[string]errConfigForPayload),
		startsWithURLs:      make(map[string][]byte),
		startsWithErrorURLs: make(map[string]errConfigForPayload),
		t:                   t,
	}
}

func (h *PayloadBasedURLHandler) AddExact(url string, payload []byte) {
	_, isSet := h.exactURLs[url]
	if isSet {
		log.Warningf("You are replacing the payload for exact url match '%s'!", url)
	}

	h.exactURLs[url] = payload
}

func (h *PayloadBasedURLHandler) AddExactError(url string, statusCode int, payload []byte) {
	_, isSet := h.exactURLs[url]
	if isSet {
		log.Warningf("You are replacing the payload for exact error url match '%s'!", url)
	}

	h.exactErrorURLs[url] = errConfigForPayload{status: statusCode, payload: payload}
}

func (h *PayloadBasedURLHandler) AddStartsWith(url string, payload []byte) {
	_, isSet := h.startsWithURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for starts with url match '%s'!", url)
	}

	h.startsWithURLs[url] = payload
}

func (h *PayloadBasedURLHandler) AddStartsWithError(url string, statusCode int, payload []byte) {
	_, isSet := h.startsWithErrorURLs[url]
	if isSet {
		log.Warningf("You are replacing the payload for starts with error url match '%s'!", url)
	}

	h.startsWithErrorURLs[url] = errConfigForPayload{status: statusCode, payload: payload}
}

func (h *PayloadBasedURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestedURL := r.URL.String()
	log.Debug("Mock for: " + requestedURL)

	for url, payload := range h.exactURLs {
		if url == requestedURL {
			log.Debug("Found Mock: " + url)

			writePayloadToResponseWriter(w, http.StatusOK, payload)
			return
		}
	}

	for url, payload := range h.startsWithURLs {
		if strings.Index(requestedURL, url) == 0 {
			log.Debug("Found Mock: " + url)

			writePayloadToResponseWriter(w, http.StatusOK, payload)
			return
		}
	}

	for url, config := range h.exactErrorURLs {
		if url == requestedURL {
			log.Debug("Found Mock: " + url)

			writePayloadToResponseWriter(w, config.status, config.payload)
			return
		}
	}

	for url, config := range h.startsWithErrorURLs {
		if strings.Index(requestedURL, url) == 0 {
			log.Debug("Found Mock: " + url)

			writePayloadToResponseWriter(w, config.status, config.payload)
			return
		}
	}

	h.t.Fatalf("no path defined for: %s", requestedURL)
}

func writePayloadToResponseWriter(w http.ResponseWriter, statusCode int, payload []byte) {
	w.WriteHeader(statusCode)
	_, err := w.Write(payload)
	if err != nil {
		panic("could not write to mock http handler")
	}
}

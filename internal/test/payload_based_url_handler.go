package test

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
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
}

func NewPayloadBasedURLHandler() *PayloadBasedURLHandler {
	return &PayloadBasedURLHandler{
		exactURLs:           make(map[string][]byte),
		exactErrorURLs:      make(map[string]errConfigForPayload),
		startsWithURLs:      make(map[string][]byte),
		startsWithErrorURLs: make(map[string]errConfigForPayload),
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
	log.Println("Mock for: " + r.URL.Path)

	for url, payload := range h.exactURLs {
		if url == r.URL.Path {
			log.Println("Found Mock: " + url)

			writePayloadToResponseWriter(w, http.StatusOK, payload)
			return
		}
	}

	for url, payload := range h.startsWithURLs {
		if strings.Index(r.URL.Path, url) == 0 {
			log.Println("Found Mock: " + url)

			writePayloadToResponseWriter(w, http.StatusOK, payload)
			return
		}
	}

	for url, config := range h.exactErrorURLs {
		if url == r.URL.Path {
			log.Println("Found Mock: " + url)

			writePayloadToResponseWriter(w, config.status, config.payload)
			return
		}
	}

	for url, config := range h.startsWithErrorURLs {
		if strings.Index(r.URL.Path, url) == 0 {
			log.Println("Found Mock: " + url)

			writePayloadToResponseWriter(w, config.status, config.payload)
			return
		}
	}

	panic("no path defined for: " + r.URL.Path)
}

func writePayloadToResponseWriter(w http.ResponseWriter, statusCode int, payload []byte) {
	w.WriteHeader(statusCode)
	_, err := w.Write(payload)
	if err != nil {
		panic("could not write to mock http handler")
	}
}

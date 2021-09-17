package test

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type errConfig struct {
	status   int
	fileName string
}

type URLHandler struct {
	exactURLs      map[string]string
	exactErrorURLs map[string]errConfig
	startsWithURLs map[string]string
}

func NewURLHandler() *URLHandler {
	return &URLHandler{
		exactURLs:      make(map[string]string),
		exactErrorURLs: make(map[string]errConfig),
		startsWithURLs: make(map[string]string),
	}
}

func (h *URLHandler) AddExact(url string, fileName string) {
	oldFileName, isSet := h.exactURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for exact url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.exactURLs[url] = fileName
}

func (h *URLHandler) AddExactError(url string, statusCode int, fileName string) {
	oldFileName, isSet := h.exactURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for exact error url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.exactErrorURLs[url] = errConfig{status: statusCode, fileName: fileName}
}

func (h *URLHandler) AddStartsWith(url string, fileName string) {
	oldFileName, isSet := h.startsWithURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for starts with url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.startsWithURLs[url] = fileName
}

func (h *URLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Mock for: " + r.URL.Path)

	for url, fileName := range h.exactURLs {
		if url == r.URL.Path {
			log.Println("Found Mock: " + url + " --> " + fileName)

			writeToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, fileName := range h.startsWithURLs {
		if strings.Index(r.URL.Path, url) == 0 {
			log.Println("Found Mock: " + url + " --> " + fileName)

			writeToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, config := range h.exactErrorURLs {
		if strings.Index(r.URL.Path, url) == 0 {
			log.Println("Found Mock: " + url + " --> " + config.fileName)

			writeToResponseWriter(w, config.status, config.fileName)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func writeToResponseWriter(w http.ResponseWriter, statusCode int, fileName string) {
	localFileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		panic("could not load local test file: " + fileName)
	}
	w.WriteHeader(statusCode)
	_, err = w.Write(localFileContent)
	if err != nil {
		panic("could not write to mock http handler")
	}
}
package test

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type URLHandler struct {
	exactURLs      map[string]string
	startsWithURLs map[string]string
}

func NewURLHandler() *URLHandler {
	return &URLHandler{
		exactURLs:      make(map[string]string),
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

			writeToResponseWriter(w, fileName)
			return
		}
	}

	for url, fileName := range h.startsWithURLs {
		if strings.Index(r.URL.Path, url) == 0 {
			log.Println("Found Mock: " + url + " --> " + fileName)

			writeToResponseWriter(w, fileName)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func writeToResponseWriter(w http.ResponseWriter, fileName string) {
	localFileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		panic("could not load local test file: " + fileName)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(localFileContent)
	if err != nil {
		panic("could not write to mock http handler")
	}
}

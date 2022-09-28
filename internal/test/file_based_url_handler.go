package test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

const urlFieldName = "url"

type errConfigForFile struct {
	Status   int
	FileName string
}

type FileBasedURLHandler struct {
	exactURLs           map[string]string
	exactErrorURLs      map[string]errConfigForFile
	startsWithURLs      map[string]string
	startsWithErrorURLs map[string]errConfigForFile
	t                   *testing.T
}

func NewFileBasedURLHandler(t *testing.T) *FileBasedURLHandler {
	return &FileBasedURLHandler{
		exactURLs:           make(map[string]string),
		exactErrorURLs:      make(map[string]errConfigForFile),
		startsWithURLs:      make(map[string]string),
		startsWithErrorURLs: make(map[string]errConfigForFile),
		t:                   t,
	}
}

func (h *FileBasedURLHandler) AddExact(url string, fileName string) {
	h.assertFileIsInTestDataFolder(fileName)

	oldFileName, isSet := h.exactURLs[url]
	if isSet {
		log.WithFields(log.Fields{urlFieldName: url, "oldFileName": oldFileName, "fileName": fileName}).Warn("Replacing the file for exact URL match")
	}

	h.exactURLs[url] = fileName
}

func (h *FileBasedURLHandler) AddExactError(url string, statusCode int, fileName string) {
	h.assertFileIsInTestDataFolder(fileName)

	oldEntry, isSet := h.exactErrorURLs[url]
	if isSet {
		log.WithFields(log.Fields{urlFieldName: url, "oldEntry": oldEntry, "fileName": fileName}).Warn("Replacing the file for exact error URL match")
	}

	h.exactErrorURLs[url] = errConfigForFile{Status: statusCode, FileName: fileName}
}

func (h *FileBasedURLHandler) AddStartsWith(url string, fileName string) {
	h.assertFileIsInTestDataFolder(fileName)

	oldFileName, isSet := h.startsWithURLs[url]
	if isSet {
		log.WithFields(log.Fields{urlFieldName: url, "oldFileName": oldFileName, "fileName": fileName}).Warn("Replacing the file for starts with URL match")
	}

	h.startsWithURLs[url] = fileName
}

func (h *FileBasedURLHandler) AddStartsWithError(url string, statusCode int, fileName string) {
	h.assertFileIsInTestDataFolder(fileName)

	oldEntry, isSet := h.startsWithErrorURLs[url]
	if isSet {
		log.WithFields(log.Fields{urlFieldName: url, "oldEntry": oldEntry, "fileName": fileName}).Warn("Replacing the file for starts with error URL match")
	}

	h.startsWithErrorURLs[url] = errConfigForFile{Status: statusCode, FileName: fileName}
}

func (h *FileBasedURLHandler) assertFileIsInTestDataFolder(fileName string) {
	if !strings.HasPrefix(fileName, "./testdata/") && !strings.HasPrefix(fileName, "testdata/") {
		h.t.Fatalf("the file you specified is not in the local 'testdata' folder: %s", fileName)
	}
}

func (h *FileBasedURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestedURL := r.URL.String()
	log.WithField(urlFieldName, requestedURL).Debug("Mock requested for URL")

	for url, fileName := range h.exactURLs {
		if url == requestedURL {
			log.WithFields(log.Fields{urlFieldName: url, "fileName": fileName}).Debug("Found mock for exact URL")

			writeFileToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, fileName := range h.startsWithURLs {
		if strings.Index(requestedURL, url) == 0 {
			log.WithFields(log.Fields{urlFieldName: url, "fileName": fileName}).Debug("Found mock for starts with URL")

			writeFileToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, config := range h.exactErrorURLs {
		if url == requestedURL {
			log.WithFields(log.Fields{urlFieldName: url, "fileName": config.FileName}).Debug("Found mock for exact error URL")

			writeFileToResponseWriter(w, config.Status, config.FileName)
			return
		}
	}

	for url, config := range h.startsWithErrorURLs {
		if strings.Index(requestedURL, url) == 0 {
			log.WithFields(log.Fields{urlFieldName: url, "fileName": config.FileName}).Debug("Found mock for starts with error URL")

			writeFileToResponseWriter(w, config.Status, config.FileName)
			return
		}
	}

	h.t.Fatalf("no path defined for: %s", requestedURL)
}

func writeFileToResponseWriter(w http.ResponseWriter, statusCode int, fileName string) {
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

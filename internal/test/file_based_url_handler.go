package test

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type errConfigForFile struct {
	status   int
	fileName string
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
		log.Warningf("You are replacing the file for exact url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.exactURLs[url] = fileName
}

func (h *FileBasedURLHandler) AddExactError(url string, statusCode int, fileName string) {
	h.assertFileIsInTestDataFolder(fileName)

	oldEntry, isSet := h.exactErrorURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for exact error url match '%s'! Old: %s, new: %s", url, oldEntry.fileName, fileName)
	}

	h.exactErrorURLs[url] = errConfigForFile{status: statusCode, fileName: fileName}
}

func (h *FileBasedURLHandler) AddStartsWith(url string, fileName string) {
	h.assertFileIsInTestDataFolder(fileName)

	oldFileName, isSet := h.startsWithURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for starts with url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.startsWithURLs[url] = fileName
}

func (h *FileBasedURLHandler) AddStartsWithError(url string, statusCode int, fileName string) {
	h.assertFileIsInTestDataFolder(fileName)

	oldEntry, isSet := h.startsWithErrorURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for starts with error url match '%s'! Old: %s, new: %s", url, oldEntry.fileName, fileName)
	}

	h.startsWithErrorURLs[url] = errConfigForFile{status: statusCode, fileName: fileName}
}

func (h *FileBasedURLHandler) assertFileIsInTestDataFolder(fileName string) {
	if !strings.HasPrefix(fileName, "./testdata/") {
		h.t.Fatalf("the file you specified is not in the local 'testdata' folder: %s", fileName)
	}
}

func (h *FileBasedURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestedURL := r.URL.String()
	log.Println("Mock for: " + requestedURL)

	for url, fileName := range h.exactURLs {
		if url == requestedURL {
			log.Println("Found Mock: " + url + " --> " + fileName)

			writeFileToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, fileName := range h.startsWithURLs {
		if strings.Index(requestedURL, url) == 0 {
			log.Println("Found Mock: " + url + " --> " + fileName)

			writeFileToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, config := range h.exactErrorURLs {
		if url == requestedURL {
			log.Println("Found Mock: " + url + " --> " + config.fileName)

			writeFileToResponseWriter(w, config.status, config.fileName)
			return
		}
	}

	for url, config := range h.startsWithErrorURLs {
		if strings.Index(requestedURL, url) == 0 {
			log.Println("Found Mock: " + url + " --> " + config.fileName)

			writeFileToResponseWriter(w, http.StatusOK, config.fileName)
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

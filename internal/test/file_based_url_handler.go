package test

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type errConfigForFile struct {
	status   int
	fileName string
}

type FileBasedURLHandler struct {
	exactURLs      map[string]string
	exactErrorURLs map[string]errConfigForFile
	startsWithURLs map[string]string
}

func NewFileBasedURLHandler() *FileBasedURLHandler {
	return &FileBasedURLHandler{
		exactURLs:      make(map[string]string),
		exactErrorURLs: make(map[string]errConfigForFile),
		startsWithURLs: make(map[string]string),
	}
}

func (h *FileBasedURLHandler) AddExact(url string, fileName string) {
	assertFileIsInTestDataFolder(fileName)

	oldFileName, isSet := h.exactURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for exact url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.exactURLs[url] = fileName
}

func (h *FileBasedURLHandler) AddExactError(url string, statusCode int, fileName string) {
	assertFileIsInTestDataFolder(fileName)

	oldFileName, isSet := h.exactURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for exact error url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.exactErrorURLs[url] = errConfigForFile{status: statusCode, fileName: fileName}
}

func (h *FileBasedURLHandler) AddStartsWith(url string, fileName string) {
	assertFileIsInTestDataFolder(fileName)

	oldFileName, isSet := h.startsWithURLs[url]
	if isSet {
		log.Warningf("You are replacing the file for starts with url match '%s'! Old: %s, new: %s", url, oldFileName, fileName)
	}

	h.startsWithURLs[url] = fileName
}

func assertFileIsInTestDataFolder(fileName string) {
	if !strings.HasPrefix(fileName, "./testdata/") {
		panic("the file you specified is not in the local 'testdata' folder: " + fileName)
	}
}

func (h *FileBasedURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Mock for: " + r.URL.Path)

	for url, fileName := range h.exactURLs {
		if url == r.URL.Path {
			log.Println("Found Mock: " + url + " --> " + fileName)

			writeFileToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, fileName := range h.startsWithURLs {
		if strings.Index(r.URL.Path, url) == 0 {
			log.Println("Found Mock: " + url + " --> " + fileName)

			writeFileToResponseWriter(w, http.StatusOK, fileName)
			return
		}
	}

	for url, config := range h.exactErrorURLs {
		if url == r.URL.Path {
			log.Println("Found Mock: " + url + " --> " + config.fileName)

			writeFileToResponseWriter(w, config.status, config.fileName)
			return
		}
	}

	panic("no path defined for: " + r.URL.Path)
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

package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

type FileBasedURLHandlerWithSink struct {
	*FileBasedURLHandler
	sink map[string][]byte
}

func NewFileBasedURLHandlerWithSink(t *testing.T) *FileBasedURLHandlerWithSink {
	return &FileBasedURLHandlerWithSink{
		FileBasedURLHandler: NewFileBasedURLHandler(t),
		sink:                make(map[string][]byte),
	}
}

func (h *FileBasedURLHandlerWithSink) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()

	switch r.Method {
	case http.MethodGet:
		h.FileBasedURLHandler.ServeHTTP(w, r)
	case http.MethodPost, http.MethodPut:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			h.t.Fatalf("could not read payload from POST|PUT request: %s", url)
		}
		h.sink[url] = payload
	default:
		h.t.Fatalf("unsupported HTTP method %s for URL: %s", r.Method, r.URL.String())
	}
}

func (h *FileBasedURLHandlerWithSink) GetStoredPayloadForURL(url string, container interface{}) {
	payload, found := h.sink[url]
	if !found {
		h.t.Fatalf("could not find payload for URL: %s", url)
	}

	err := json.Unmarshal(payload, container)
	if err != nil {
		h.t.Fatalf("could not unmarshall JSON payload: %s", err)
	}
}

package test

import (
	"net/http"
	"testing"
)

type CombinedURLHandler struct {
	t                 *testing.T
	useFileHandler    map[string]bool
	fileHandler       *FileBasedURLHandler
	templatingHandler *TemplatingPayloadBasedURLHandler
}

func NewCombinedURLHandler(t *testing.T, templateFile string) *CombinedURLHandler {
	return &CombinedURLHandler{
		t:                 t,
		useFileHandler:    make(map[string]bool),
		fileHandler:       NewFileBasedURLHandler(t),
		templatingHandler: NewTemplatingPayloadBasedURLHandler(t, templateFile),
	}
}

func (h *CombinedURLHandler) AddExactFile(url string, fileName string) {
	_, alreadyThere := h.useFileHandler[url]
	if alreadyThere {
		h.t.Fatalf("%s has been already stored, check your test configuration", url)
	}

	h.useFileHandler[url] = true
	h.fileHandler.AddExact(url, fileName)
}

func (h *CombinedURLHandler) AddExactTemplate(url string, templatingData interface{}) {
	_, alreadyThere := h.useFileHandler[url]
	if alreadyThere {
		h.t.Fatalf("%s has been already stored, check your test configuration", url)
	}

	h.useFileHandler[url] = false
	h.templatingHandler.AddExact(url, templatingData)
}

func (h *CombinedURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	if h.useFileHandler[url] {
		h.fileHandler.ServeHTTP(w, r)
		return
	}

	h.templatingHandler.ServeHTTP(w, r)
}

/*
/api/v2/slo/7d07efde-b714-3e6e-ad95-08490e2540c4?from=1631862000000&timeFrame=GTF&to=1631865600000

/api/v2/slo/7d07efde-b714-3e6e-ad95-08490e2540c4?from=1609459200000&timeFrame=GTF&to=1609545600000

*/

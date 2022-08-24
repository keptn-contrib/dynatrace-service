package test

import (
	"bytes"
	"net/http"
	"testing"
	"text/template"
)

// TemplatingPayloadBasedURLHandler encapsulates a payload-based URL handler and extends its with templating functionality
type TemplatingPayloadBasedURLHandler struct {
	t       *testing.T
	handler *PayloadBasedURLHandler
}

// NewTemplatingPayloadBasedURLHandler creates a new TemplatingPayloadBasedURLHandler
func NewTemplatingPayloadBasedURLHandler(t *testing.T) *TemplatingPayloadBasedURLHandler {
	return &TemplatingPayloadBasedURLHandler{
		t:       t,
		handler: NewPayloadBasedURLHandler(t),
	}
}

func (h *TemplatingPayloadBasedURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

// AddExact will add an exact match handler for a given url, using a template file and templatingData to produce a byte[] payload
func (h *TemplatingPayloadBasedURLHandler) AddExact(url string, templateFilename string, templatingData interface{}) {
	h.handler.AddExact(url, h.executeTemplate(templateFilename, templatingData))
}

// AddExactError will add an exact match error handler for a given url, including an error status code using a template file and templatingData to produce a byte[] payload
func (h *TemplatingPayloadBasedURLHandler) AddExactError(url string, statusCode int, templateFilename string, templatingData interface{}) {
	h.handler.AddExactError(url, statusCode, h.executeTemplate(templateFilename, templatingData))
}

func (h *TemplatingPayloadBasedURLHandler) executeTemplate(templateFilename string, templatingData interface{}) []byte {
	tpl, err := template.ParseFiles(templateFilename)
	if err != nil {
		h.t.Fatalf("could not create template: %s", err)
	}

	buf := bytes.Buffer{}
	err = tpl.Execute(&buf, &templatingData)
	if err != nil {
		h.t.Fatalf("could not write to buffer: %s", err)
	}

	return buf.Bytes()
}

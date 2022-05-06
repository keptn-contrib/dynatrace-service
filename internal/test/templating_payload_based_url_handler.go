package test

import (
	"bytes"
	"net/http"
	"testing"
	"text/template"
)

type TemplatingPayloadBasedURLHandler struct {
	t       *testing.T
	tpl     *template.Template // use text/template to avoid escaping
	handler *PayloadBasedURLHandler
}

func NewTemplatingPayloadBasedURLHandler(t *testing.T, templateFile string) *TemplatingPayloadBasedURLHandler {
	tpl, err := template.ParseFiles(templateFile)
	if err != nil {
		t.Fatalf("could not create template: %s", err)
	}

	return &TemplatingPayloadBasedURLHandler{
		t:       t,
		tpl:     tpl,
		handler: NewPayloadBasedURLHandler(t),
	}
}

func (h *TemplatingPayloadBasedURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *TemplatingPayloadBasedURLHandler) AddExact(url string, templatingData interface{}) {
	h.handler.AddExact(url, h.writeToBuffer(templatingData))
}

func (h *TemplatingPayloadBasedURLHandler) AddExactError(url string, statusCode int, templatingData interface{}) {
	h.handler.AddExactError(url, statusCode, h.writeToBuffer(templatingData))
}

func (h *TemplatingPayloadBasedURLHandler) writeToBuffer(templatingData interface{}) []byte {
	buf := bytes.Buffer{}
	err := h.tpl.Execute(&buf, &templatingData)
	if err != nil {
		h.t.Fatalf("could not write to buffer: %s", err)
	}

	return buf.Bytes()
}

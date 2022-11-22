package test

import (
	"net/http"
	"testing"
)

// EmptyURLHandler is an implementation of http.Handler that will not handle anything.
type EmptyURLHandler struct {
	t *testing.T
}

// NewEmptyURLHandler creates a new EmptyURLHandler instance.
func NewEmptyURLHandler(t *testing.T) *EmptyURLHandler {
	return &EmptyURLHandler{
		t: t,
	}
}

func (h *EmptyURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.t.Fatal("ServeHTTP() should not be needed in this mock!")
}

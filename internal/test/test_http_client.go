package test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
)

// CreateHTTPClient create a fake http client for integration tests
func CreateHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}

func CreateHandler(response []byte, statusCode int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, err := w.Write(response)
		if err != nil {
			panic(err)
		}
	})
}

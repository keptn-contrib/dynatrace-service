package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	log "github.com/sirupsen/logrus"
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

// CreateHTTPSClient creates a test client with a httptest server that responds to API calls
func CreateHTTPSClient(handler http.Handler) (*http.Client, string, func()) {

	server := httptest.NewTLSServer(handler)

	cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
	if err != nil {
		log.Fatal(err)
	}

	certpool := x509.NewCertPool()
	certpool.AddCert(cert)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, server.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				RootCAs: certpool,
			},
		},
	}

	return client, server.URL, server.Close
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

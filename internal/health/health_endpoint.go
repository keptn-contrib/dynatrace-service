package health

import (
	"context"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

const healthEndpointPattern = "/health"

// healthHandler will return 204 for requests.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	log.Trace("alive...")
}

// HealthEndpoint is a HTTP server that offers a health endpoint.
type HealthEndpoint struct {
	waitGroup *sync.WaitGroup
	Server    *http.Server
}

// NewHealthEndpoint creates a new HealthEndpoint.
func NewHealthEndpoint(addr string) *HealthEndpoint {
	m := http.NewServeMux()
	m.HandleFunc(healthEndpointPattern, healthHandler)
	return &HealthEndpoint{
		waitGroup: &sync.WaitGroup{},
		Server:    &http.Server{Addr: addr, Handler: m},
	}
}

// Start causes the server for the health endpoint to listen and serve in a separate goroutine.
func (h *HealthEndpoint) Start() {
	h.waitGroup.Add(1)
	go func() {
		defer h.waitGroup.Done()
		err := h.Server.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				log.WithError(err).Error("Health endpoint ListenAndServe returned an unexpected error")
			}
		}
	}()
}

// Stop shuts down the server for the health endpoint and waits for the associated goroutine to finish.
func (h *HealthEndpoint) Stop() {
	h.Server.Shutdown(context.Background())
	h.waitGroup.Wait()
}

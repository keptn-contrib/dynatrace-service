package context

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const nameFieldName = "name"

// NewTriggeredTimeoutContext creates a new Context that is done:
// a. after a timeout which is triggered after waitCtx is done, or
// b. if parentCtx is done.
// It returns the new context and a function that cancels the context.
func NewTriggeredTimeoutContext(parentCtx context.Context, waitCtx context.Context, timeout time.Duration, name string) (context.Context, func()) {
	log.WithField(nameFieldName, name).WithField("timeout", timeout).Debug("Creating ThenWithTimeoutContext")
	triggeredTimeoutCtx, innerCancel := context.WithCancel(parentCtx)

	// start a monitoring goroutine to ultimately set triggeredTimeoutContext done state
	// the goroutine will finish immediately if triggeredTimeoutContext is already done
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		defer innerCancel()

		// wait for waitCtx or triggeredTimeoutContext
		select {
		case <-waitCtx.Done():

		case <-triggeredTimeoutCtx.Done():
			return
		}

		log.WithField(nameFieldName, name).Debug("TriggeredTimeoutContext triggered")

		// wait for the timeout or triggeredTimeoutContext
		select {
		case <-time.After(timeout):

		case <-triggeredTimeoutCtx.Done():
			return
		}

		log.WithField(nameFieldName, name).Debug("TriggeredTimeoutContext timeout")

	}()

	cancel := func() {
		log.WithField(nameFieldName, name).Debug("Cancelling TriggeredTimeoutContext")
		innerCancel()
		waitGroup.Wait()
	}

	return triggeredTimeoutCtx, cancel
}

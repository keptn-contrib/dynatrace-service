package context

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const allowedDeltaSeconds = 1.0

// TestNewTriggeredTimeoutContext_Simple tests that a TriggeredTimeoutContext is done timeoutSeconds after being the waitCtx is done.
func TestNewTriggeredTimeoutContext_Simple(t *testing.T) {
	const timeoutSeconds = 8

	initialCtx, trigger := context.WithCancel(context.Background())

	triggeredTimeoutCtx, _ := NewTriggeredTimeoutContext(context.Background(), initialCtx, timeoutSeconds*time.Second, "triggeredTimeoutCtx1")

	startTime := time.Now()
	trigger()
	<-triggeredTimeoutCtx.Done()
	duration := time.Now().Sub(startTime)

	assert.InDelta(t, timeoutSeconds, duration/time.Second, allowedDeltaSeconds)
}

// TestNewTriggeredTimeoutContext_Chained tests that two TriggeredTimeoutContexts can be chained together with the expected total timeout.
func TestNewTriggeredTimeoutContext_Chained(t *testing.T) {
	const timeoutSeconds1 = 8
	const timeoutSeconds2 = 4

	initialCtx, trigger := context.WithCancel(context.Background())

	triggeredTimeoutCtx1, _ := NewTriggeredTimeoutContext(context.Background(), initialCtx, timeoutSeconds1*time.Second, "triggeredTimeoutCtx1")
	triggeredTimeoutCtx2, _ := NewTriggeredTimeoutContext(context.Background(), triggeredTimeoutCtx1, timeoutSeconds2*time.Second, "triggeredTimeoutCtx2")

	startTime := time.Now()
	trigger()
	<-triggeredTimeoutCtx2.Done()
	duration := time.Now().Sub(startTime)

	assert.InDelta(t, timeoutSeconds1+timeoutSeconds2, duration/time.Second, allowedDeltaSeconds)
}

// TestNewTriggeredTimeoutContext_Early tests cancelling the first of two chained TriggeredTimeoutContexts causes its timeout to be skipped.
func TestNewTriggeredTimeoutContext_Early(t *testing.T) {
	const timeoutSeconds1 = 6
	const timeoutSeconds2 = 2

	initialCtx, trigger := context.WithCancel(context.Background())

	triggeredTimeoutCtx1, cancel1 := NewTriggeredTimeoutContext(context.Background(), initialCtx, timeoutSeconds1*time.Second, "triggeredTimeoutCtx1")
	triggeredTimeoutCtx2, _ := NewTriggeredTimeoutContext(context.Background(), triggeredTimeoutCtx1, timeoutSeconds2*time.Second, "triggeredTimeoutCtx2")

	startTime := time.Now()
	trigger()
	cancel1()
	<-triggeredTimeoutCtx2.Done()
	duration := time.Now().Sub(startTime)

	assert.InDelta(t, timeoutSeconds2, duration/time.Second, allowedDeltaSeconds)
}

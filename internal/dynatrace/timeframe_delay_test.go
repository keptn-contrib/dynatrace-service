package dynatrace

import (
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/stretchr/testify/assert"
)

// TestTimeframeDelay_NoDelayRequired tests when timeframe is sufficiently in the past such that no delay is required.
func TestTimeframeDelay_NoDelayRequired(t *testing.T) {

	// timeframe that starts 180 seconds ago and ends 120 seconds ago
	timeframe, err := common.NewTimeframe(time.Now().Add(-180*time.Second), time.Now().Add(-120*time.Second))
	assert.NoError(t, err)

	// required delay of 1 minute to now, but not exceeding 4 minutes
	d := NewTimeframeDelay(*timeframe, 60*time.Second, 240*time.Second)
	waitDuration, err := d.calculateWaitDuration()

	assert.NoError(t, err)
	assert.EqualValues(t, 0, waitDuration)
}

// TestTimeframeDelay_NearFutureExceedsMaximumAllowed tests when the timeframe is too far in the future and not allowed.
func TestTimeframeDelay_NearFutureExceedsMaximumAllowed(t *testing.T) {

	// timeframe that starts 1 minute into the future and ends 3 minutes in the future
	timeframe, err := common.NewTimeframe(time.Now().Add(60*time.Second), time.Now().Add(180*time.Second))
	assert.NoError(t, err)

	// required delay of 2 minutes to now, but not exceeding 4 minutes
	d := NewTimeframeDelay(*timeframe, 120*time.Second, 240*time.Second)
	waitDuration, err := d.calculateWaitDuration()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
	assert.EqualValues(t, 0, waitDuration)
}

// TestTimeframeDelay_NearFutureButAllowed tests when the timeframe is in the future but within the maximum wait time.
func TestTimeframeDelay_NearFutureButAllowed(t *testing.T) {

	// timeframe that starts 30 seconds into the future and ends 90 seconds in the future
	timeframe, err := common.NewTimeframe(time.Now().Add(30*time.Second), time.Now().Add(90*time.Second))
	assert.NoError(t, err)

	// required delay of 2 minutes to now, but not exceeding 4 minutes
	d := NewTimeframeDelay(*timeframe, 120*time.Second, 240*time.Second)
	waitDuration, err := d.calculateWaitDuration()

	assert.NoError(t, err)
	assert.InDelta(t, 210, waitDuration.Seconds(), 0.1)
}

// TestTimeframeDelay_NearPastButAllowed tests when the timeframe is in the near past.
func TestTimeframeDelay_NearPastButAllowed(t *testing.T) {

	// timeframe that starts 90 seconds ago and ends 30 seconds ago
	timeframe, err := common.NewTimeframe(time.Now().Add(-90*time.Second), time.Now().Add(-30*time.Second))
	assert.NoError(t, err)

	// required delay of 2 minutes to now, but not exceeding 4 minutes
	d := NewTimeframeDelay(*timeframe, 120*time.Second, 240*time.Second)
	waitDuration, err := d.calculateWaitDuration()

	assert.NoError(t, err)
	assert.InDelta(t, 90, waitDuration.Seconds(), 0.1)
}

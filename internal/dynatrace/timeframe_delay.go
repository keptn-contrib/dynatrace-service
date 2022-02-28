package dynatrace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

// TimeframeDelay encapsulates the calculation and execution of a delay relative to a timeframe.
type TimeframeDelay struct {
	timeframe     common.Timeframe
	requiredDelay time.Duration
	maximumWait   time.Duration
}

// NewTimeframeDelay creates a new TimeframeDelay with the specified timeframe, required delay and maximum wait.
func NewTimeframeDelay(timeframe common.Timeframe, requiredDelay time.Duration, maximumWait time.Duration) TimeframeDelay {
	return TimeframeDelay{
		timeframe:     timeframe,
		requiredDelay: requiredDelay,
		maximumWait:   maximumWait,
	}
}

// Wait calculates and executes a delay relative to a timeframe. If this exceeds the maximum wait or if the sleep is interrupted, an error is returned.
func (d TimeframeDelay) Wait(ctx context.Context) error {
	waitDuration, err := d.calculateWaitDuration()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return errors.New("Delay sleep interrupted")

	case <-time.After(waitDuration):
		return nil
	}
}

// calculateWaitDuration calculated the wait time such that the required delay is satisfied. If this exceeds the maximum wait, an error is returned.
func (d TimeframeDelay) calculateWaitDuration() (time.Duration, error) {
	durationSinceTimeframeEnd := time.Since(d.timeframe.End())
	waitDuration := d.requiredDelay - durationSinceTimeframeEnd

	// return error if the wait is too long, i.e. the timeframe end is too far in the future
	if waitDuration > d.maximumWait {
		return 0, fmt.Errorf("Required delay of %.2f seconds exceeds maximum of %.2f seconds", waitDuration.Seconds(), d.maximumWait.Seconds())
	}

	// if sufficient time has passed, don't wait at all
	if waitDuration < 0 {
		return 0, nil
	}

	return waitDuration, nil
}

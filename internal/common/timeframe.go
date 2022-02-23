package common

import (
	"errors"
	"time"
)

// Timeframe represents a timeframe with a start and end time.
type Timeframe struct {
	start time.Time
	end   time.Time
}

// NewTimeframe creates a new timeframe from start and end times.
func NewTimeframe(start time.Time, end time.Time) (*Timeframe, error) {
	// ensure start time is before end time
	if end.Sub(start).Seconds() < 0 {
		return nil, errors.New("error validating timeframe: start needs to be before end")
	}

	return &Timeframe{
		start: start,
		end:   end,
	}, nil
}

// Start gets the start of the timeframe.
func (t Timeframe) Start() time.Time {
	return t.start
}

// End gets the end of the timeframe.
func (t Timeframe) End() time.Time {
	return t.end
}

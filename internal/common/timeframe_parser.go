package common

import (
	"fmt"

	"github.com/keptn/go-utils/pkg/common/timeutils"
)

// Timeframe represents a timeframe with a start and end time.
type TimeframeParser struct {
	start string
	end   string
}

// NewTimeframe creates a new timeframe from start and end strings.
func NewTimeframeParser(start string, end string) TimeframeParser {
	return TimeframeParser{
		start: start,
		end:   end,
	}
}

func (p TimeframeParser) Parse() (*Timeframe, error) {
	start, err := timeutils.ParseTimestamp(p.start)
	if err != nil {
		return nil, fmt.Errorf("error parsing timeframe start: %w", err)
	}

	end, err := timeutils.ParseTimestamp(p.end)
	if err != nil {
		return nil, fmt.Errorf("error parsing timeframe end: %w", err)
	}

	return NewTimeframe(*start, *end)
}

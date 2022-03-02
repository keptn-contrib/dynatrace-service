package common

import (
	"fmt"

	"github.com/keptn/go-utils/pkg/common/timeutils"
)

// TimeframeParser represents a timeframe ready to be parsed.
type TimeframeParser struct {
	start string
	end   string
}

// NewTimeframeParser creates a new TimeframeParser ready to parse the specified start and end strings.
func NewTimeframeParser(start string, end string) TimeframeParser {
	return TimeframeParser{
		start: start,
		end:   end,
	}
}

// Parse parses the start and end strings to create a Timeframe.
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

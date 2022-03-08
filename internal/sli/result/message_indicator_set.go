package result

import (
	"strings"
)

// messageIndicatorSet groups indicators for a particular message. Ordering is also stored to allow the messageIndicatorSets to be sorted later
type messageIndicatorSet struct {
	message    string
	indicators []string
	ordering   int
}

// newMessageIndicatorSet creates a new messageIndicatorSet with the specified message and ordering.
func newMessageIndicatorSet(message string, ordering int) messageIndicatorSet {
	return messageIndicatorSet{
		message:  message,
		ordering: ordering,
	}
}

// addIndicator adds an indicator to the messageIndicatorSet.
func (m *messageIndicatorSet) addIndicator(indicator string) {
	m.indicators = append(m.indicators, indicator)
}

// summaryMessage gets the summary message for the messageIndicatorSet. It has the form: "indicator_a, indicator_b: message".
func (m *messageIndicatorSet) summaryMessage() string {
	return strings.Join(m.indicators, ", ") + ": " + m.message
}

package result

import (
	"sort"
	"strings"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// SLIResultSummarizer determines an overall result and summary message for a slice of SLI results.
type SLIResultSummarizer struct {
	indicatorValues []SLIResult
}

// NewSLIResultSummarizer creates a new SLIResultSummarizer with the specified indicator values.
func NewSLIResultSummarizer(indicatorValues []SLIResult) SLIResultSummarizer {
	return SLIResultSummarizer{indicatorValues: indicatorValues}
}

// SummaryMessage gets a summarized message for all indicators in the form "indicator_A, indicator_B: error_1; indicator_C: error_2..."
func (s SLIResultSummarizer) SummaryMessage() string {
	return strings.Join(getSummaryMessages(sortMessageIndicators(groupIndicatorMessages(s.indicatorValues))), "; ")
}

// groupIndicatorMessages groups the indicators by their messages.
func groupIndicatorMessages(indicatorValues []SLIResult) map[string]messageIndicatorSet {
	messageSetMap := make(map[string]messageIndicatorSet)
	for ordering, indicator := range indicatorValues {
		if indicator.Success == false {
			ms, ok := messageSetMap[indicator.Message]
			if !ok {
				ms = newMessageIndicatorSet(indicator.Message, ordering)
			}
			ms.addIndicator(indicator.Metric)
			messageSetMap[indicator.Message] = ms
		}
	}
	return messageSetMap
}

// sortMessageIndicators sorts the map values by ordering.
func sortMessageIndicators(messageSetMap map[string]messageIndicatorSet) []messageIndicatorSet {
	messageSets := make([]messageIndicatorSet, 0, len(messageSetMap))
	for _, value := range messageSetMap {
		messageSets = append(messageSets, value)
	}

	sort.Slice(messageSets, func(i, j int) bool { return messageSets[i].ordering < messageSets[j].ordering })
	return messageSets
}

// getSummaryMessages extracts the summary messages from the messageIndicatorSets.
func getSummaryMessages(messageIndicatorSets []messageIndicatorSet) []string {
	var messagePieces []string
	for _, messageIndicatorSet := range messageIndicatorSets {
		messagePieces = append(messagePieces, messageIndicatorSet.summaryMessage())
	}
	return messagePieces
}

// Result gets the overall result for the indicator values.
func (s SLIResultSummarizer) Result() keptnv2.ResultType {

	seenWarning := false
	for _, indicator := range s.indicatorValues {
		switch indicator.IndicatorResult {
		case IndicatorResultSuccessful:
			// this is fine, do nothing
		case IndicatorResultWarning:
			seenWarning = true
		case IndicatorResultFailed:
			// if one indicator fails, the overall result is failed immediately
			return keptnv2.ResultFailed
		default:
			// an unexpected result fails the overall result immediately
			return keptnv2.ResultFailed
		}
	}

	if seenWarning {
		return keptnv2.ResultWarning
	}

	// remaining case is pass, i.e. no failure or warning occurred
	return keptnv2.ResultPass
}

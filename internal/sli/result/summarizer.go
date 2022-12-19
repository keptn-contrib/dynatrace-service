package result

import (
	"sort"
	"strings"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// Summarizer determines an overall result and summary message for a slice of SLI results.
type Summarizer struct {
	results []SLIWithSLO
}

// NewSummarizer creates a new Summarizer with the specified indicator values.
func NewSummarizer(indicatorValues []SLIWithSLO) Summarizer {
	return Summarizer{results: indicatorValues}
}

// SummaryMessage gets a summarized message for all indicators in the form "indicator_A, indicator_B: error_1; indicator_C: error_2..."
func (s Summarizer) SummaryMessage() string {
	return strings.Join(getSummaryMessages(sortMessageIndicators(groupIndicatorMessages(s.results))), "; ")
}

// groupIndicatorMessages groups the indicators by their messages.
func groupIndicatorMessages(results []SLIWithSLO) map[string]messageIndicatorSet {
	messageSetMap := make(map[string]messageIndicatorSet)
	for ordering, r := range results {
		sliResult := r.SLIResult()
		if sliResult.Success == false {
			ms, ok := messageSetMap[sliResult.Message]
			if !ok {
				ms = newMessageIndicatorSet(sliResult.Message, ordering)
			}
			ms.addIndicator(sliResult.Metric)
			messageSetMap[sliResult.Message] = ms
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

// OverallResult gets the overall result for the indicator values.
func (s Summarizer) OverallResult() keptnv2.ResultType {

	seenNonInformationalWarning := false
	for _, r := range s.results {
		sliResult := r.SLIResult()
		switch sliResult.IndicatorResult {
		case IndicatorResultSuccessful:
			// this is fine, do nothing
		case IndicatorResultWarning:
			if isSLONotInformational(r.SLODefinition()) {
				seenNonInformationalWarning = true
			}
		case IndicatorResultFailed:
			// if one indicator fails, the overall result is failed immediately
			return keptnv2.ResultFailed
		default:
			// an unexpected result fails the overall result immediately
			return keptnv2.ResultFailed
		}
	}

	if seenNonInformationalWarning {
		return keptnv2.ResultWarning
	}

	// remaining case is pass, i.e. no failure or warning occurred
	return keptnv2.ResultPass
}

func isSLONotInformational(slo keptnapi.SLO) bool {
	return hasActualSLOCriteria(slo.Pass) || hasActualSLOCriteria(slo.Warning)
}

func hasActualSLOCriteria(sloCriteria []*keptnapi.SLOCriteria) bool {
	for _, c := range sloCriteria {
		if c == nil {
			continue
		}

		if len(c.Criteria) > 0 {
			return true
		}
	}
	return false
}

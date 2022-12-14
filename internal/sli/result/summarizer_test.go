package result

import (
	"fmt"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

const (
	noDataPointsMessage     = "no data points"
	tooManyPointsMessage    = "too many points"
	errorQueryingAPIMessage = "error querying API"

	metricAName = "metric_a"
	metricBName = "metric_b"
	metricCName = "metric_c"
	metricDName = "metric_d"
	metricEName = "metric_e"
	metricFName = "metric_f"
	metricGName = "metric_g"
)

var (
	metricASLO = CreateInformationalSLODefinition(metricAName)
	metricBSLO = CreateInformationalSLODefinition(metricBName)
	metricCSLO = CreateInformationalSLODefinition(metricCName)
	metricDSLO = CreateInformationalSLODefinition(metricDName)
	metricESLO = CreateInformationalSLODefinition(metricEName)
	metricFSLO = CreateInformationalSLODefinition(metricFName)
	metricGSLO = CreateInformationalSLODefinition(metricGName)
)

// TestSummarizer_SummaryMessage tests the construction of summary messages for SLI results.
// i.e. successful indicators do not appear, indicators are grouped by message, groups are ordered by the message's first appearance.
func TestSummarizer_SummaryMessage(t *testing.T) {
	var results = []SLIWithSLO{
		NewWarningSLIWithSLO(metricASLO, noDataPointsMessage),
		NewWarningSLIWithSLO(metricBSLO, tooManyPointsMessage),
		NewWarningSLIWithSLO(metricCSLO, tooManyPointsMessage),
		NewSuccessfulSLIWithSLO(metricDSLO, 100),
		NewWarningSLIWithSLO(metricESLO, tooManyPointsMessage),
		NewWarningSLIWithSLO(metricFSLO, noDataPointsMessage),
		NewFailedSLIWithSLO(metricGSLO, errorQueryingAPIMessage),
	}

	message := NewSummarizer(results).SummaryMessage()

	assert.EqualValues(t,
		fmt.Sprintf(
			"%s, %s: %s; %s, %s, %s: %s; %s: %s",
			metricAName, metricFName, noDataPointsMessage,
			metricBName, metricCName, metricEName, tooManyPointsMessage,
			metricGName, errorQueryingAPIMessage),
		message)
}

// TestSummarizer_OverallResult tests that the overall result is extracted correctly.
// i.e. failed has precedence over warning, warning has precedence over pass.
func TestSummarizer_OverallResult(t *testing.T) {

	tests := []struct {
		name                  string
		results               []SLIWithSLO
		expectedOverallResult keptnv2.ResultType
	}{
		{
			name: "pass",
			results: []SLIWithSLO{
				NewSuccessfulSLIWithSLO(metricASLO, 100),
				NewSuccessfulSLIWithSLO(metricBSLO, 100),
				NewSuccessfulSLIWithSLO(metricCSLO, 100),
			},
			expectedOverallResult: keptnv2.ResultPass,
		},
		{
			name: "warning has precedence over pass",
			results: []SLIWithSLO{
				NewSuccessfulSLIWithSLO(metricASLO, 100),
				NewWarningSLIWithSLO(metricBSLO, noDataPointsMessage),
				NewSuccessfulSLIWithSLO(metricCSLO, 100),
			},
			expectedOverallResult: keptnv2.ResultWarning,
		},
		{
			name: "failed has precedence",
			results: []SLIWithSLO{
				NewSuccessfulSLIWithSLO(metricASLO, 100),
				NewWarningSLIWithSLO(metricBSLO, noDataPointsMessage),
				NewFailedSLIWithSLO(metricCSLO, errorQueryingAPIMessage),
			},
			expectedOverallResult: keptnv2.ResultFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSummarizer(tt.results).OverallResult()
			assert.EqualValues(t, tt.expectedOverallResult, result)
		})
	}
}

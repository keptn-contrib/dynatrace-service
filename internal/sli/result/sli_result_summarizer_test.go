package result

import (
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

// TestSLIResultSummarizer_SummaryMessage tests the construction of summary messages for SLI results.
// i.e. successful indicators do not appear, indicators are grouped by message, groups are ordered by the message's first appearance.
func TestSLIResultSummarizer_SummaryMessage(t *testing.T) {
	var results = []SLIResult{
		NewWarningSLIResult("metric_a", "no data points"),
		NewWarningSLIResult("metric_b", "too many points"),
		NewWarningSLIResult("metric_c", "too many points"),
		NewSuccessfulSLIResult("metric_d", 100),
		NewWarningSLIResult("metric_e", "too many points"),
		NewWarningSLIResult("metric_f", "no data points"),
		NewFailedSLIResult("metric_g", "error querying API"),
	}

	message := NewSLIResultSummarizer(results).SummaryMessage()

	assert.EqualValues(t, "metric_a, metric_f: no data points; metric_b, metric_c, metric_e: too many points; metric_g: error querying API", message)
}

// TestSLIResultSummarizer_Result tests that the overall result is extracted correctly.
// i.e. failed has precedence over warning, warning has precedence over pass.
func TestSLIResultSummarizer_Result(t *testing.T) {

	tests := []struct {
		name            string
		indicatorValues []SLIResult
		expectedResult  keptnv2.ResultType
	}{
		{
			name: "pass",
			indicatorValues: []SLIResult{
				NewSuccessfulSLIResult("metric_a", 100),
				NewSuccessfulSLIResult("metric_b", 100),
				NewSuccessfulSLIResult("metric_c", 100),
			},
			expectedResult: keptnv2.ResultPass,
		},
		{
			name: "warning has precedence over pass",
			indicatorValues: []SLIResult{
				NewSuccessfulSLIResult("metric_a", 100),
				NewWarningSLIResult("metric_b", "no data points"),
				NewSuccessfulSLIResult("metric_c", 100),
			},
			expectedResult: keptnv2.ResultWarning,
		},
		{
			name: "failed has precedence",
			indicatorValues: []SLIResult{
				NewSuccessfulSLIResult("metric_a", 100),
				NewWarningSLIResult("metric_b", "no data points"),
				NewFailedSLIResult("metric_c", "error querying API"),
			},
			expectedResult: keptnv2.ResultFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSLIResultSummarizer(tt.indicatorValues).Result()
			assert.EqualValues(t, tt.expectedResult, result)
		})
	}
}

package result

import (
	"fmt"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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
	metricASLO = createTestSLO(metricAName)
	metricBSLO = createTestSLO(metricBName)
	metricCSLO = createTestSLO(metricCName)
	metricDSLO = createTestSLO(metricDName)
	metricESLO = createTestSLO(metricEName)
	metricFSLO = createTestSLO(metricFName)
	metricGSLO = createTestSLO(metricGName)
)

// TestSummarizer_SummaryMessage tests the construction of summary messages for SLI results.
// i.e. successful indicators do not appear, indicators are grouped by message, groups are ordered by the message's first appearance.
func TestSummarizer_SummaryMessage(t *testing.T) {
	var results = []SLIWithSLO{
		newWarningSLIWithSLO(metricASLO, noDataPointsMessage),
		newWarningSLIWithSLO(metricBSLO, tooManyPointsMessage),
		newWarningSLIWithSLO(metricCSLO, tooManyPointsMessage),
		newSuccessfulSLIWithSLO(metricDSLO, 100),
		newWarningSLIWithSLO(metricESLO, tooManyPointsMessage),
		newWarningSLIWithSLO(metricFSLO, noDataPointsMessage),
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
				newSuccessfulSLIWithSLO(metricASLO, 100),
				newSuccessfulSLIWithSLO(metricBSLO, 100),
				newSuccessfulSLIWithSLO(metricCSLO, 100),
			},
			expectedOverallResult: keptnv2.ResultPass,
		},
		{
			name: "warning has precedence over pass",
			results: []SLIWithSLO{
				newSuccessfulSLIWithSLO(metricASLO, 100),
				newWarningSLIWithSLO(metricBSLO, noDataPointsMessage),
				newSuccessfulSLIWithSLO(metricCSLO, 100),
			},
			expectedOverallResult: keptnv2.ResultWarning,
		},
		{
			name: "failed has precedence",
			results: []SLIWithSLO{
				newSuccessfulSLIWithSLO(metricASLO, 100),
				newWarningSLIWithSLO(metricBSLO, noDataPointsMessage),
				NewFailedSLIWithSLO(metricCSLO, errorQueryingAPIMessage),
			},
			expectedOverallResult: keptnv2.ResultFailed,
		},
		{
			name: "informational warning ignored",
			results: []SLIWithSLO{
				newSuccessfulSLIWithSLO(metricASLO, 100),
				newWarningSLIWithSLO(CreateInformationalSLO(metricBName), noDataPointsMessage),
			},
			expectedOverallResult: keptnv2.ResultPass,
		},
		{
			name: "informational failure has precedence",
			results: []SLIWithSLO{
				newSuccessfulSLIWithSLO(metricASLO, 100),
				NewFailedSLIWithSLO(CreateInformationalSLO(metricBName), errorQueryingAPIMessage),
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

func createTestSLO(name string) SLO {
	return SLO{
		SLI:    name,
		Pass:   SLOCriteriaList{{Criteria: []string{"<=100"}}},
		Weight: 1,
	}
}

// Test_isSLONotInformational tests that SLOs can be categorized as informational or not, including some YAML corner cases.
func Test_isSLONotInformational(t *testing.T) {
	tests := []struct {
		name                     string
		sloYAML                  string
		expectedNotInformational bool
	}{
		{
			name: "not informational with both pass and warning",
			sloYAML: `
sli: "response_time_p95"
key_sli: false
pass:
  - criteria:
    - "<600"    
warning:
  - criteria:
    - "<=800"
`,
			expectedNotInformational: true,
		},
		{
			name: "not informational with just pass",
			sloYAML: `
sli: "response_time_p95"
key_sli: false
pass:
  - criteria:
    - "<600"    
`,
			expectedNotInformational: true,
		},
		{
			name: "not informational with just warning",
			sloYAML: `
sli: "response_time_p95"
key_sli: false
warning:
  - criteria:
    - "<=800"
`,
			expectedNotInformational: true,
		},
		{
			name: "not informational with actual empty pass criterion",
			sloYAML: `
sli: "response_time_p95"
key_sli: false
pass:
  - criteria:
    - ""    
`,
			expectedNotInformational: true,
		},
		{
			name: "informational with no pass or warning",
			sloYAML: `
sli: "response_time_p95"
key_sli: false
`,
			expectedNotInformational: false,
		},
		{
			name: "informational with empty pass",
			sloYAML: `
sli: "response_time_p95"
key_sli: false
pass:
`,
			expectedNotInformational: false,
		},

		{
			name: "informational with pass with no actual criteria",
			sloYAML: `
sli: "response_time_p95"
key_sli: false
pass:
  - criteria:
`,
			expectedNotInformational: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			slo := SLO{}
			err := yaml.Unmarshal([]byte(tt.sloYAML), &slo)
			assert.NoError(t, err)
			assert.EqualValues(t, tt.expectedNotInformational, slo.IsNotInformational())

		})
	}
}

func newSuccessfulSLIWithSLO(sloDefinition SLO, value float64) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     newSuccessfulSLIResult(sloDefinition.SLI, value),
		sloDefinition: sloDefinition,
	}
}

func newSuccessfulSLIResult(metric string, value float64) SLIResult {
	return NewSuccessfulSLIResultWithQuery(metric, value, "")
}

func newWarningSLIWithSLO(sloDefinition SLO, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     newWarningSLIResult(sloDefinition.SLI, message),
		sloDefinition: sloDefinition,
	}
}

func newWarningSLIResult(metric string, message string) SLIResult {
	return NewWarningSLIResultWithQuery(metric, message, "")
}

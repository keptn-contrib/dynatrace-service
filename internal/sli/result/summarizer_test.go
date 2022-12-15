package result

import (
	"fmt"
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
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
		{
			name: "informational warning ignored",
			results: []SLIWithSLO{
				NewSuccessfulSLIWithSLO(metricASLO, 100),
				NewWarningSLIWithSLO(CreateInformationalSLODefinition(metricBName), noDataPointsMessage),
			},
			expectedOverallResult: keptnv2.ResultPass,
		},
		{
			name: "informational failure has precedence",
			results: []SLIWithSLO{
				NewSuccessfulSLIWithSLO(metricASLO, 100),
				NewFailedSLIWithSLO(CreateInformationalSLODefinition(metricBName), errorQueryingAPIMessage),
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

func createTestSLO(name string) keptnapi.SLO {
	return keptnapi.SLO{
		SLI:    name,
		Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=100"}}},
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

			slo := keptnapi.SLO{}
			err := yaml.Unmarshal([]byte(tt.sloYAML), &slo)
			assert.NoError(t, err)
			assert.EqualValues(t, tt.expectedNotInformational, isSLONotInformational(slo))

		})
	}
}

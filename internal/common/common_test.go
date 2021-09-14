package common

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseMarkdownConfigurationParams(t *testing.T) {
	testConfigs := []struct {
		input          string
		expectedResult *keptnapi.ServiceLevelObjectives
	}{
		// single result
		{
			"KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			createSLO("90%", "70%", "single_result", "pass", 1, "avg"),
		},
		// several results, p50
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p50",
			createSLO("50%", "40%", "several_results", "pass", 3, "p50"),
		},
		// several results, p90
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p90",
			createSLO("50%", "40%", "several_results", "pass", 3, "p90"),
		},
		// several results, p95
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "pass", 3, "p95"),
		},
		// several results, p95, all
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=all;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "all", 3, "p95"),
		},
		// several results, p95, pass_or_warn
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass_or_warn;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "pass_or_warn", 3, "p95"),
		},

		// several results, p95, fallback to pass if compare function is unknown
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=warn;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "pass", 3, "p95"),
		},
		// several results, fallback if function is unknown e.g. p97
		{
			"KQG.Total.Pass=51%;KQG.Total.Warning=41%;KQG.Compare.WithScore=pass;KQG.Compare.Results=4;KQG.Compare.Function=p97",
			createSLO("51%", "41%", "several_results", "pass", 4, "avg"),
		},
	}
	for _, config := range testConfigs {
		actualSLO := &keptnapi.ServiceLevelObjectives{
			Objectives: []*keptnapi.SLO{},
			TotalScore: &keptnapi.SLOScore{
				Pass:    "",
				Warning: "",
			},
			Comparison: &keptnapi.SLOComparison{
				CompareWith:               "",
				IncludeResultWithScore:    "",
				NumberOfComparisonResults: 0,
				AggregateFunction:         "",
			},
		}
		ParseMarkdownConfiguration(config.input, actualSLO)

		assert.EqualValues(t, actualSLO, config.expectedResult)
	}
}

func createSLO(pass string, warning string, compareWith string, include string, numberOfResults int, aggregateFunc string) *keptnapi.ServiceLevelObjectives {
	return &keptnapi.ServiceLevelObjectives{
		Objectives: []*keptnapi.SLO{},
		TotalScore: &keptnapi.SLOScore{
			Pass:    pass,
			Warning: warning,
		},
		Comparison: &keptnapi.SLOComparison{
			CompareWith:               compareWith,
			IncludeResultWithScore:    include,
			NumberOfComparisonResults: numberOfResults,
			AggregateFunction:         aggregateFunc,
		},
	}
}

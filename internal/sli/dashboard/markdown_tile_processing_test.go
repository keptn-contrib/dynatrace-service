package dashboard

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseMarkdownConfigurationParams(t *testing.T) {
	testConfigs := []struct {
		input              string
		expectedScore      *keptnapi.SLOScore
		expectedComparison *keptnapi.SLOComparison
	}{
		// single result
		{
			"KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			createSLOScore("90%", "70%"),
			createSLOComparison("single_result", "pass", 1, "avg"),
		},
		// several results, p50
		{
			"KQG.Total.Pass=91%;KQG.Total.Warning=71%;KQG.Compare.WithScore=pass;KQG.Compare.Results=2;KQG.Compare.Function=p50",
			createSLOScore("91%", "71%"),
			createSLOComparison("several_results", "pass", 2, "p50"),
		},
		// several results, p90
		{
			"KQG.Total.Pass=92%;KQG.Total.Warning=72%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p90",
			createSLOScore("92%", "72%"),
			createSLOComparison("several_results", "pass", 3, "p90"),
		},
		// several results, p95
		{
			"KQG.Total.Pass=93%;KQG.Total.Warning=73%;KQG.Compare.WithScore=pass;KQG.Compare.Results=4;KQG.Compare.Function=p95",
			createSLOScore("93%", "73%"),
			createSLOComparison("several_results", "pass", 4, "p95"),
		},
		// several results, p95, all
		{
			"KQG.Total.Pass=94%;KQG.Total.Warning=74%;KQG.Compare.WithScore=all;KQG.Compare.Results=5;KQG.Compare.Function=p95",
			createSLOScore("94%", "74%"),
			createSLOComparison("several_results", "all", 5, "p95"),
		},
		// several results, p95, pass_or_warn
		{
			"KQG.Total.Pass=95%;KQG.Total.Warning=75%;KQG.Compare.WithScore=pass_or_warn;KQG.Compare.Results=6;KQG.Compare.Function=p95",
			createSLOScore("95%", "75%"),
			createSLOComparison("several_results", "pass_or_warn", 6, "p95"),
		},
		// several results, p95, fallback to pass if compare function is unknown
		{
			"KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=warn;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			createSLOScore("96%", "76%"),
			createSLOComparison("several_results", "pass", 7, "p95"),
		},
		// several results, fallback if function is unknown e.g. p97
		{
			"KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=8;KQG.Compare.Function=p97",
			createSLOScore("97%", "77%"),
			createSLOComparison("several_results", "pass", 8, "avg"),
		},
		// several results, fallback if function is unknown e.g. p97, ignore dashboard query behaviour
		{
			"KQG.Total.Pass=98%;KQG.Total.Warning=78%;KQG.Compare.WithScore=pass;KQG.Compare.Results=9;KQG.Compare.Function=p97;KQG.QueryBehavior=ParseOnChange",
			createSLOScore("98%", "78%"),
			createSLOComparison("several_results", "pass", 9, "avg"),
		},
	}
	for _, config := range testConfigs {
		actualScore, actualComparison := parseMarkdownConfiguration(config.input, createDefaultSLOScore(), createDefaultSLOComparison())

		assert.EqualValues(t, config.expectedScore, actualScore)
		assert.EqualValues(t, config.expectedComparison, actualComparison)
	}
}

func createSLOScore(pass string, warning string) *keptnapi.SLOScore {
	return &keptnapi.SLOScore{
		Pass:    pass,
		Warning: warning,
	}
}
func createSLOComparison(compareWith string, include string, numberOfResults int, aggregateFunc string) *keptnapi.SLOComparison {
	return &keptnapi.SLOComparison{
		CompareWith:               compareWith,
		IncludeResultWithScore:    include,
		NumberOfComparisonResults: numberOfResults,
		AggregateFunction:         aggregateFunc,
	}
}

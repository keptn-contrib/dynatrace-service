package dashboard

import (
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
)

func TestParseMarkdownConfigurationParams_SuccessCases(t *testing.T) {
	testConfigs := []struct {
		name               string
		input              string
		expectedScore      keptnapi.SLOScore
		expectedComparison keptnapi.SLOComparison
	}{
		{
			name:               "single result, without percent sign",
			input:              "KQG.Total.Pass=90;KQG.Total.Warning=70;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedScore:      createSLOScore("90", "70"),
			expectedComparison: createSLOComparison(CompareResultsSingle, CompareWithScorePass, 1, CompareFunctionAvg),
		},
		{
			name:               "single result, without percent sign, with decimals",
			input:              "KQG.Total.Pass=90.84;KQG.Total.Warning=70.22;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedScore:      createSLOScore("90.84", "70.22"),
			expectedComparison: createSLOComparison(CompareResultsSingle, CompareWithScorePass, 1, CompareFunctionAvg),
		},
		{
			name:               "single result, with percent sign, with decimals",
			input:              "KQG.Total.Pass=90.84%;KQG.Total.Warning=70.22%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedScore:      createSLOScore("90.84%", "70.22%"),
			expectedComparison: createSLOComparison(CompareResultsSingle, CompareWithScorePass, 1, CompareFunctionAvg),
		},
		{
			name:               "single result",
			input:              "KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedScore:      createSLOScore("90%", "70%"),
			expectedComparison: createSLOComparison(CompareResultsSingle, CompareWithScorePass, 1, CompareFunctionAvg),
		},
		{
			name:               "several results p50",
			input:              "KQG.Total.Pass=91%;KQG.Total.Warning=71%;KQG.Compare.WithScore=pass;KQG.Compare.Results=2;KQG.Compare.Function=p50",
			expectedScore:      createSLOScore("91%", "71%"),
			expectedComparison: createSLOComparison(CompareResultsMultiple, CompareWithScorePass, 2, CompareFunctionP50),
		},
		{
			name:               "several results p90",
			input:              "KQG.Total.Pass=92%;KQG.Total.Warning=72%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p90",
			expectedScore:      createSLOScore("92%", "72%"),
			expectedComparison: createSLOComparison(CompareResultsMultiple, CompareWithScorePass, 3, CompareFunctionP90),
		},
		{
			name:               "several results p95",
			input:              "KQG.Total.Pass=93%;KQG.Total.Warning=73%;KQG.Compare.WithScore=pass;KQG.Compare.Results=4;KQG.Compare.Function=p95",
			expectedScore:      createSLOScore("93%", "73%"),
			expectedComparison: createSLOComparison(CompareResultsMultiple, CompareWithScorePass, 4, CompareFunctionP95),
		},
		{
			name:               "several results p95 all",
			input:              "KQG.Total.Pass=94%;KQG.Total.Warning=74%;KQG.Compare.WithScore=all;KQG.Compare.Results=5;KQG.Compare.Function=p95",
			expectedScore:      createSLOScore("94%", "74%"),
			expectedComparison: createSLOComparison(CompareResultsMultiple, CompareWithScoreAll, 5, CompareFunctionP95),
		},
		{
			name:               "several results p95 pass_or_warn",
			input:              "KQG.Total.Pass=95%;KQG.Total.Warning=75%;KQG.Compare.WithScore=pass_or_warn;KQG.Compare.Results=6;KQG.Compare.Function=p95",
			expectedScore:      createSLOScore("95%", "75%"),
			expectedComparison: createSLOComparison(CompareResultsMultiple, CompareWithScorePassOrWarn, 6, CompareFunctionP95),
		},
	}
	for _, config := range testConfigs {
		t.Run(config.name, func(t *testing.T) {
			result, err := parseMarkdownConfiguration(config.input, createDefaultSLOScore(), createDefaultSLOComparison())
			if assert.NoError(t, err) {
				assert.EqualValues(t, config.expectedScore, result.totalScore)
				assert.EqualValues(t, config.expectedComparison, result.comparison)
			}
		})
	}
}

func TestParseMarkdownConfigurationParams_ErrorCases(t *testing.T) {
	const duplicationError = "duplicate key"

	testConfigs := []struct {
		name             string
		input            string
		expectedMessages []string
	}{
		{
			name:             "unknown compare with score function",
			input:            "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=warn;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			expectedMessages: []string{CompareWithScore, "warn"},
		},
		{
			name:             "unknown compare function, p97",
			input:            "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=8;KQG.Compare.Function=p97",
			expectedMessages: []string{CompareFunction, "p97"},
		},
		{
			name:             "wrong number of results, 0",
			input:            "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=0;KQG.Compare.Function=p95",
			expectedMessages: []string{CompareResults, "0"},
		},
		{
			name:             "wrong number of results, decimal",
			input:            "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7.5;KQG.Compare.Function=p95",
			expectedMessages: []string{CompareResults, "7.5"},
		},
		{
			name:             "wrong number of results, string",
			input:            "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=three;KQG.Compare.Function=p95",
			expectedMessages: []string{CompareResults, "three"},
		},
		{
			name:             "duplicate total pass",
			input:            "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Total.Pass=96%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			expectedMessages: []string{TotalPass, duplicationError},
		},
		{
			name:             "duplicate total warning",
			input:            "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Total.Warning=96%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			expectedMessages: []string{TotalWarning, duplicationError},
		},
		{
			name:             "duplicate total compare with score",
			input:            "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.WithScore=all",
			expectedMessages: []string{CompareWithScore, duplicationError},
		},
		{
			name:             "duplicate compare results",
			input:            "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.Results=1",
			expectedMessages: []string{CompareResults, duplicationError},
		},
		{
			name:             "duplicate total compare function",
			input:            "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.Function=p90",
			expectedMessages: []string{CompareFunction, duplicationError},
		},
		{
			name:             "invalid value for total pass",
			input:            "KQG.Total.Pass=96Pct",
			expectedMessages: []string{TotalPass, "96Pct"},
		},
		{
			name:             "invalid value for total warning",
			input:            "KQG.Total.Warning=OneHundred",
			expectedMessages: []string{TotalWarning, "OneHundred"},
		},
		{
			name:  "multiple problems - one for each",
			input: "KQG.Total.Pass=96Pct;KQG.Total.Warning=OneHundred;KQG.Compare.WithScore=passing;KQG.Compare.Results=7.5;KQG.Compare.Function=p97;",
			expectedMessages: []string{
				TotalPass, "96Pct",
				TotalWarning, "OneHundred",
				CompareWithScore, "passing",
				CompareFunction, "p97",
				CompareResults, "7.5",
			},
		},
	}
	for _, config := range testConfigs {
		t.Run(config.name, func(t *testing.T) {
			result, err := parseMarkdownConfiguration(config.input, createDefaultSLOScore(), createDefaultSLOComparison())
			if assert.Error(t, err) {
				assert.Nil(t, result)
				for _, msg := range config.expectedMessages {
					assert.Contains(t, err.Error(), msg)
				}
			}
		})
	}
}

func createSLOScore(pass string, warning string) keptnapi.SLOScore {
	return keptnapi.SLOScore{
		Pass:    pass,
		Warning: warning,
	}
}
func createSLOComparison(compareWith string, include string, numberOfResults int, aggregateFunc string) keptnapi.SLOComparison {
	return keptnapi.SLOComparison{
		CompareWith:               compareWith,
		IncludeResultWithScore:    include,
		NumberOfComparisonResults: numberOfResults,
		AggregateFunction:         aggregateFunc,
	}
}

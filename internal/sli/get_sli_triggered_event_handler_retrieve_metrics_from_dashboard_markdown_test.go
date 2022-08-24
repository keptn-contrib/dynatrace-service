package sli

import (
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/dashboard"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

type data struct {
	Markdown string // needs to match template file variable
}

// TestRetrieveMetricsFromDashboard_MarkdownParsingWorks test lots of markdown tiles without errors in the SLO definition.
// there is one SLO tile as well to have a fully working example where SLOs would be stored as well
func TestRetrieveMetricsFromDashboard_MarkdownParsingWorks(t *testing.T) {
	const templateFile = "./testdata/dashboards/markdown/markdown-tile-parsing-single-sli-template.json"
	const sliName = "static_slo_-_pass"

	expectedSLORequest := buildSLORequest("7d07efde-b714-3e6e-ad95-08490e2540c4")

	assertionFunc := createSuccessfulSLIResultAssertionsFunc(sliName, 95, expectedSLORequest)

	expectedSLO := &keptnapi.SLO{
		SLI:     sliName,
		Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{">=90.000000"}}},
		Warning: []*keptnapi.SLOCriteria{{Criteria: []string{">=75.000000"}}},
		Weight:  1,
		KeySLI:  false,
	}

	tests := []struct {
		name        string
		markdown    string
		expectedSLO *keptnapi.ServiceLevelObjectives
	}{
		{
			name:        "only defaults",
			markdown:    "just some information here, that does not add any configuration",
			expectedSLO: createSLO("90%", "75%", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "total pass, rest is defaults",
			markdown:    "KQG.Total.Pass=91%;",
			expectedSLO: createSLO("91%", "75%", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "total warning, rest is defaults",
			markdown:    "KQG.Total.Warning=76%;",
			expectedSLO: createSLO("90%", "76%", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "include with, rest is defaults",
			markdown:    "KQG.Compare.WithScore=all;",
			expectedSLO: createSLO("90%", "75%", dashboard.CompareResultsSingle, dashboard.CompareWithScoreAll, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "number of results, rest is defaults",
			markdown:    "KQG.Compare.Results=2;",
			expectedSLO: createSLO("90%", "75%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePass, 2, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "aggregate func, rest is defaults",
			markdown:    "KQG.Compare.Function=p95;",
			expectedSLO: createSLO("90%", "75%", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionP95, expectedSLO),
		},
		{
			name:        "single result, without percent sign",
			markdown:    "KQG.Total.Pass=90;KQG.Total.Warning=70;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90", "70", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "single result, without percent sign, with decimals",
			markdown:    "KQG.Total.Pass=90.84;KQG.Total.Warning=70.22;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90.84", "70.22", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "single result, with percent sign, with decimals",
			markdown:    "KQG.Total.Pass=90.84%;KQG.Total.Warning=70.22%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90.84%", "70.22%", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "single result",
			markdown:    "KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90%", "70%", dashboard.CompareResultsSingle, dashboard.CompareWithScorePass, 1, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "several results p50",
			markdown:    "KQG.Total.Pass=91%;KQG.Total.Warning=71%;KQG.Compare.WithScore=pass;KQG.Compare.Results=2;KQG.Compare.Function=p50",
			expectedSLO: createSLO("91%", "71%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePass, 2, dashboard.CompareFunctionP50, expectedSLO),
		},
		{
			name:        "several results p90",
			markdown:    "KQG.Total.Pass=92%;KQG.Total.Warning=72%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p90",
			expectedSLO: createSLO("92%", "72%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePass, 3, dashboard.CompareFunctionP90, expectedSLO),
		},
		{
			name:        "several results p95",
			markdown:    "KQG.Total.Pass=93%;KQG.Total.Warning=73%;KQG.Compare.WithScore=pass;KQG.Compare.Results=4;KQG.Compare.Function=p95",
			expectedSLO: createSLO("93%", "73%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePass, 4, dashboard.CompareFunctionP95, expectedSLO),
		},
		{
			name:        "several results p95 all",
			markdown:    "KQG.Total.Pass=94%;KQG.Total.Warning=74%;KQG.Compare.WithScore=all;KQG.Compare.Results=5;KQG.Compare.Function=p95",
			expectedSLO: createSLO("94%", "74%", dashboard.CompareResultsMultiple, dashboard.CompareWithScoreAll, 5, dashboard.CompareFunctionP95, expectedSLO),
		},
		{
			name:        "several results p95 pass_or_warn",
			markdown:    "KQG.Total.Pass=95%;KQG.Total.Warning=75%;KQG.Compare.WithScore=pass_or_warn;KQG.Compare.Results=6;KQG.Compare.Function=p95",
			expectedSLO: createSLO("95%", "75%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePassOrWarn, 6, dashboard.CompareFunctionP95, expectedSLO),
		},
		{
			name:        "newline at the end",
			markdown:    "KQG.Total.Pass=85%;KQG.Total.Warning=80%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=avg\\n",
			expectedSLO: createSLO("85%", "80%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePass, 3, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "newline after every key value pair",
			markdown:    "KQG.Total.Pass=85%;\\nKQG.Total.Warning=80%;\\nKQG.Compare.WithScore=pass;\\nKQG.Compare.Results=3;\\nKQG.Compare.Function=avg\\n",
			expectedSLO: createSLO("85%", "80%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePass, 3, dashboard.CompareFunctionAvg, expectedSLO),
		},
		{
			name:        "whitespace around each key-value pair",
			markdown:    "KQG.Total.Pass = 85%;\\nKQG.Total.Warning = 80%;\\nKQG.Compare.WithScore = pass;\\nKQG.Compare.Results = 3;\\nKQG.Compare.Function = avg\\n",
			expectedSLO: createSLO("85%", "80%", dashboard.CompareResultsMultiple, dashboard.CompareWithScorePass, 3, dashboard.CompareFunctionAvg, expectedSLO),
		},
	}
	for _, markdownTest := range tests {
		t.Run(markdownTest.name, func(t *testing.T) {
			handler := test.NewCombinedURLHandler(t)
			handler.AddExactTemplate(
				dynatrace.DashboardsPath+"/"+testDashboardID,
				templateFile,
				&data{
					Markdown: markdownTest.markdown,
				},
			)
			handler.AddExactFile(expectedSLORequest, "./testdata/dashboards/slo_tiles/passing_slo/slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json")

			rClient := &uploadErrorResourceClientMock{t: t}
			runAndAssertThatDashboardTestIsCorrect(t, testGetSLIEventData, handler, rClient, getSLIFinishedEventSuccessAssertionsFunc, assertionFunc)

			assert.EqualValues(t, rClient.uploadedSLOs, markdownTest.expectedSLO)
		})
	}
}

func createSLO(totalPass string, totalWarning string, compareWith string, includeWithScore string, numberOfResults int, aggregateFunc string, slo *keptnapi.SLO) *keptnapi.ServiceLevelObjectives {
	return &keptnapi.ServiceLevelObjectives{
		Comparison: &keptnapi.SLOComparison{
			CompareWith:               compareWith,
			IncludeResultWithScore:    includeWithScore,
			NumberOfComparisonResults: numberOfResults,
			AggregateFunction:         aggregateFunc,
		},
		TotalScore: &keptnapi.SLOScore{
			Pass:    totalPass,
			Warning: totalWarning,
		},
		Objectives: []*keptnapi.SLO{slo},
	}
}

// TestRetrieveMetricsFromDashboard_MarkdownParsingErrors test lots of markdown tiles with errors in the SLO definition.
// This will result in a failing SLIResult, as this is not allowed.
func TestRetrieveMetricsFromDashboard_MarkdownParsingErrors(t *testing.T) {
	const templateFile = "./testdata/dashboards/markdown/markdown-tile-parsing-errors-template.json"

	const indicator = "no metric"
	const duplicationError = "duplicate key"
	const invalidValueError = "invalid value"

	tests := []struct {
		name           string
		markdown       string
		assertionsFunc func(*testing.T, sliResult)
	}{
		{
			name:           "unknown compare with score function",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=warn;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareWithScore, "warn"),
		},
		{
			name:           "unknown compare function, p97",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=8;KQG.Compare.Function=p97",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareFunction, "p97"),
		},
		{
			name:           "wrong number of results, 0",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=0;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareResults, "0"),
		},
		{
			name:           "wrong number of results, decimal",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7.5;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareResults, "7.5"),
		},
		{
			name:           "wrong number of results, string",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=three;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareResults, "three"),
		},
		{
			name:           "duplicate total pass",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Total.Pass=96%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.TotalPass, duplicationError),
		},
		{
			name:           "duplicate total warning",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Total.Warning=96%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.TotalWarning, duplicationError),
		},
		{
			name:           "duplicate total compare with score",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.WithScore=all",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareWithScore, duplicationError),
		},
		{
			name:           "duplicate compare results",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.Results=1",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareResults, duplicationError),
		},
		{
			name:           "duplicate total compare function",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.Function=p90",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.CompareFunction, duplicationError),
		},
		{
			name:           "invalid value for total pass",
			markdown:       "KQG.Total.Pass=96Pct",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.TotalPass, "96Pct"),
		},
		{
			name:           "invalid value for total warning",
			markdown:       "KQG.Total.Warning=OneHundred",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, dashboard.TotalWarning, "OneHundred"),
		},
		{
			name:     "multiple problems - one for each",
			markdown: "KQG.Total.Pass=96Pct;KQG.Total.Warning=OneHundred;KQG.Compare.WithScore=passing;KQG.Compare.Results=7.5;KQG.Compare.Function=p97;",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator,
				dashboard.TotalPass, "96Pct",
				dashboard.TotalWarning, "OneHundred",
				dashboard.CompareWithScore, "passing",
				dashboard.CompareFunction, "p97",
				dashboard.CompareResults, "7.5"),
		},
		{
			name:           "extra content on new line causes invalid value",
			markdown:       "KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=all;KQG.Compare.Results=4;KQG.Compare.Function=avg\\n\\n## View results in the [Keptn Bridge] (https://cloudautomation.live.dynatrace.com/bridge/project/sbs)",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, invalidValueError, dashboard.CompareFunction),
		},
	}
	for _, markdownTest := range tests {
		t.Run(markdownTest.name, func(t *testing.T) {
			handler := test.NewTemplatingPayloadBasedURLHandler(t)
			handler.AddExact(
				dynatrace.DashboardsPath+"/"+testDashboardID,
				templateFile,
				&data{
					Markdown: markdownTest.markdown,
				},
			)

			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, markdownTest.assertionsFunc)
		})
	}
}

// TestRetrieveMetricsFromDashboard_MarkdownMultipleTilesErrors tests multiple markdown tiles per dashboard without errors in the SLO definition.
// This will result in a failing SLIResult, as it is not allowed to have multiple markdown tiles for configuration.
func TestRetrieveMetricsFromDashboard_MarkdownMultipleTilesErrors(t *testing.T) {
	type data struct {
		MarkdownTileOne string // needs to match template file variable
		MarkdownTileTwo string // needs to match template file variable
	}

	const templateFile = "./testdata/dashboards/markdown/markdown-tile-parsing-errors-multiple-tiles-template.json"

	const indicator = "no metric"
	const multipleTilesErrorMsg = "only one markdown tile allowed"

	tests := []struct {
		name           string
		firstMarkdown  string
		secondMarkdown string
		assertionsFunc func(*testing.T, sliResult)
	}{
		{
			name:           "union does not overlap",
			firstMarkdown:  "KQG.Total.Pass=90%",
			secondMarkdown: "KQG.Total.Warning=70%",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, multipleTilesErrorMsg),
		},
		{
			name:           "two full configurations",
			firstMarkdown:  "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;",
			secondMarkdown: "KQG.Total.Pass=94%;KQG.Total.Warning=74%;KQG.Compare.WithScore=all;KQG.Compare.Results=5;KQG.Compare.Function=p90;",
			assertionsFunc: createFailedSLIResultAssertionsFunc(indicator, multipleTilesErrorMsg),
		},
	}
	for _, markdownTest := range tests {
		t.Run(markdownTest.name, func(t *testing.T) {
			handler := test.NewTemplatingPayloadBasedURLHandler(t)
			handler.AddExact(
				dynatrace.DashboardsPath+"/"+testDashboardID,
				templateFile,
				&data{
					MarkdownTileOne: markdownTest.firstMarkdown,
					MarkdownTileTwo: markdownTest.secondMarkdown,
				},
			)

			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, markdownTest.assertionsFunc)
		})
	}
}

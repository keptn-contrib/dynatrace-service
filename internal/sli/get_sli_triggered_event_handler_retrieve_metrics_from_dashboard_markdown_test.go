package sli

import (
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

type data struct {
	Markdown string // needs to match template file variable
}

// TestRetrieveMetricsFromDashboard_MarkdownParsingWorks test lots of markdown tiles without errors in the SLO definition.
// there is one SLO tile as well to have a fully working example where SLOs would be stored as well
func TestRetrieveMetricsFromDashboard_MarkdownParsingWorks(t *testing.T) {
	const templateFile = "./testdata/dashboards/markdown/markdown-tile-parsing-single-sli-template.json"
	assertionFunc := createSuccessfulSLIResultAssertionsFunc("Static_SLO_-_Pass", 95)

	expectedSLO := &keptnapi.SLO{
		SLI:     "Static_SLO_-_Pass",
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
			expectedSLO: createSLO("90%", "75%", "single_result", "pass", 1, "avg", expectedSLO),
		},
		{
			name:        "total pass, rest is defaults",
			markdown:    "KQG.Total.Pass=91%;",
			expectedSLO: createSLO("91%", "75%", "single_result", "pass", 1, "avg", expectedSLO),
		},
		{
			name:        "total warning, rest is defaults",
			markdown:    "KQG.Total.Warning=76%;",
			expectedSLO: createSLO("90%", "76%", "single_result", "pass", 1, "avg", expectedSLO),
		},
		{
			name:        "include with, rest is defaults",
			markdown:    "KQG.Compare.WithScore=all;",
			expectedSLO: createSLO("90%", "75%", "single_result", "all", 1, "avg", expectedSLO),
		},
		{
			name:        "number of results, rest is defaults",
			markdown:    "KQG.Compare.Results=2;",
			expectedSLO: createSLO("90%", "75%", "several_results", "pass", 2, "avg", expectedSLO),
		},
		{
			name:        "aggregate func, rest is defaults",
			markdown:    "KQG.Compare.Function=p95;",
			expectedSLO: createSLO("90%", "75%", "single_result", "pass", 1, "p95", expectedSLO),
		},
		{
			name:        "single result, without percent sign",
			markdown:    "KQG.Total.Pass=90;KQG.Total.Warning=70;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90", "70", "single_result", "pass", 1, "avg", expectedSLO),
		},
		{
			name:        "single result, without percent sign, with decimals",
			markdown:    "KQG.Total.Pass=90.84;KQG.Total.Warning=70.22;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90.84", "70.22", "single_result", "pass", 1, "avg", expectedSLO),
		},
		{
			name:        "single result, with percent sign, with decimals",
			markdown:    "KQG.Total.Pass=90.84%;KQG.Total.Warning=70.22%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90.84%", "70.22%", "single_result", "pass", 1, "avg", expectedSLO),
		},
		{
			name:        "single result",
			markdown:    "KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			expectedSLO: createSLO("90%", "70%", "single_result", "pass", 1, "avg", expectedSLO),
		},
		{
			name:        "several results p50",
			markdown:    "KQG.Total.Pass=91%;KQG.Total.Warning=71%;KQG.Compare.WithScore=pass;KQG.Compare.Results=2;KQG.Compare.Function=p50",
			expectedSLO: createSLO("91%", "71%", "several_results", "pass", 2, "p50", expectedSLO),
		},
		{
			name:        "several results p90",
			markdown:    "KQG.Total.Pass=92%;KQG.Total.Warning=72%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p90",
			expectedSLO: createSLO("92%", "72%", "several_results", "pass", 3, "p90", expectedSLO),
		},
		{
			name:        "several results p95",
			markdown:    "KQG.Total.Pass=93%;KQG.Total.Warning=73%;KQG.Compare.WithScore=pass;KQG.Compare.Results=4;KQG.Compare.Function=p95",
			expectedSLO: createSLO("93%", "73%", "several_results", "pass", 4, "p95", expectedSLO),
		},
		{
			name:        "several results p95 all",
			markdown:    "KQG.Total.Pass=94%;KQG.Total.Warning=74%;KQG.Compare.WithScore=all;KQG.Compare.Results=5;KQG.Compare.Function=p95",
			expectedSLO: createSLO("94%", "74%", "several_results", "all", 5, "p95", expectedSLO),
		},
		{
			name:        "several results p95 pass_or_warn",
			markdown:    "KQG.Total.Pass=95%;KQG.Total.Warning=75%;KQG.Compare.WithScore=pass_or_warn;KQG.Compare.Results=6;KQG.Compare.Function=p95",
			expectedSLO: createSLO("95%", "75%", "several_results", "pass_or_warn", 6, "p95", expectedSLO),
		},
	}
	for _, markdownTest := range tests {
		t.Run(markdownTest.name, func(t *testing.T) {
			handler := test.NewCombinedURLHandler(t, templateFile)
			handler.AddExactTemplate(
				dynatrace.DashboardsPath+"/"+testDashboardID,
				&data{
					Markdown: markdownTest.markdown,
				},
			)
			handler.AddExactFile(dynatrace.SLOPath+"/7d07efde-b714-3e6e-ad95-08490e2540c4?from=1609459200000&timeFrame=GTF&to=1609545600000", "./testdata/dashboards/slo_tiles/passing_slo/slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json")

			rClient := &uploadErrorResourceClientMock{t: t}
			runAndAssertThatDashboardTestIsCorrect(t, testDataExplorerGetSLIEventData, handler, rClient, getSLIFinishedEventSuccessAssertionsFunc, assertionFunc)

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

	tests := []struct {
		name           string
		markdown       string
		assertionsFunc func(*testing.T, *keptnv2.SLIResult)
	}{
		{
			name:           "unknown compare with score function",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=warn;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.withscore", "warn"),
		},
		{
			name:           "unknown compare function, p97",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=8;KQG.Compare.Function=p97",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.function", "p97"),
		},
		{
			name:           "wrong number of results, 0",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=0;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.results", "0"),
		},
		{
			name:           "wrong number of results, decimal",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7.5;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.results", "7.5"),
		},
		{
			name:           "wrong number of results, string",
			markdown:       "KQG.Total.Pass=97%;KQG.Total.Warning=77%;KQG.Compare.WithScore=pass;KQG.Compare.Results=three;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.results", "three"),
		},
		{
			name:           "duplicate total pass",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Total.Pass=96%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.total.pass", "duplicate key"),
		},
		{
			name:           "duplicate total warning",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Total.Warning=96%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.total.warning", "duplicate key"),
		},
		{
			name:           "duplicate total compare with score",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.WithScore=all",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.withscore", "duplicate key"),
		},
		{
			name:           "duplicate compare results",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.Results=1",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.results", "duplicate key"),
		},
		{
			name:           "duplicate total compare function",
			markdown:       "KQG.Total.Pass=96%;KQG.Total.Warning=76%;KQG.Compare.WithScore=pass;KQG.Compare.Results=7;KQG.Compare.Function=p95;KQG.Compare.Function=p90",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.compare.function", "duplicate key"),
		},
		{
			name:           "invalid value for total pass",
			markdown:       "KQG.Total.Pass=96Pct",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.total.pass", "96Pct"),
		},
		{
			name:           "invalid value for total warning",
			markdown:       "KQG.Total.Warning=OneHundred",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric", "kqg.total.warning", "OneHundred"),
		},
		{
			name:     "multiple problems - one for each",
			markdown: "KQG.Total.Pass=96Pct;KQG.Total.Warning=OneHundred;KQG.Compare.WithScore=passing;KQG.Compare.Results=7.5;KQG.Compare.Function=p97;",
			assertionsFunc: createFailedSLIResultAssertionsFunc("no metric",
				"kqg.total.pass", "96Pct",
				"kqg.total.warning", "OneHundred",
				"kqg.compare.withscore", "passing",
				"kqg.compare.function", "p97",
				"kqg.compare.results", "7.5"),
		},
	}
	for _, markdownTest := range tests {
		t.Run(markdownTest.name, func(t *testing.T) {
			handler := test.NewTemplatingPayloadBasedURLHandler(t, templateFile)
			handler.AddExact(
				dynatrace.DashboardsPath+"/"+testDashboardID,
				&data{
					Markdown: markdownTest.markdown,
				},
			)

			rClient := &uploadErrorResourceClientMock{t: t}
			runAndAssertThatDashboardTestIsCorrect(t, testDataExplorerGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, markdownTest.assertionsFunc)
		})
	}
}

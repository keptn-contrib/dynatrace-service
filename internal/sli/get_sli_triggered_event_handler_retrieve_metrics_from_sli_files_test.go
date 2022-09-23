package sli

import (
	"fmt"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI could not be found because of a misspelled indicator name - e.g. 'response_time_p59' instead of 'response_time_p95'
//   - this would have lead to a fallback to default SLIs, but should return an error now.
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButIndicatorCannotBeMatched(t *testing.T) {
	// no need to have something here, because we should not send an API request
	handler := test.NewFileBasedURLHandler(t)

	// error here in the misspelled indicator:
	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		"response_time_p59": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "SLI definition", "not found"))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but Dynatrace cannot process the query correctly and returns a 400 error
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryIsNotValid(t *testing.T) {
	// error here: metric(s)Selector=
	handler := test.NewFileBasedURLHandler(t)

	// error here as well: metric(s)Selector=
	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricsSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "error parsing Metrics v2 query"))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI has errors, so parsing the YAML file would not be possible
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreInvalidYAML(t *testing.T) {
	// make sure we would not be able to query any metric due to a parsing error
	handler := test.NewFileBasedURLHandler(t)

	const errorMessage = "invalid YAML file - some parsing issue"
	rClient := newResourceClientMockWithGetSLIsError(t, fmt.Errorf(errorMessage))

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, errorMessage))
}

// TestRetrieveMetricsFromFile_SecurityProblemsV2 tests the success case for file-based SecurityProblemsV2 SLIs.
func TestRetrieveMetricsFromFile_SecurityProblemsV2(t *testing.T) {
	const (
		securityProblemsRequest           = "/api/v2/securityProblems?from=1609459200000&securityProblemSelector=status%28%22open%22%29&to=1609545600000"
		testDataFolder                    = "./testdata/sli_files/secpv2_success/"
		testIndicatorSecurityProblemCount = "security_problem_count"
	)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(securityProblemsRequest, testDataFolder+"security_problems_status_open.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorSecurityProblemCount: "SECPV2;securityProblemSelector=status(\"open\")",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorSecurityProblemCount, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSecurityProblemCount, 103, securityProblemsRequest))
}

// TestRetrieveMetricsFromFile_ProblemsV2 tests the success case for file-based ProblemsV2 SLIs.
func TestRetrieveMetricsFromFile_ProblemsV2(t *testing.T) {
	const (
		testDataFolder            = "./testdata/sli_files/pv2_success/"
		testIndicatorProblemCount = "problem_count"
	)

	expectedProblemsRequest := buildProblemsV2Request("status%28%22open%22%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedProblemsRequest, testDataFolder+"problems_status_open.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorProblemCount: "PV2;problemSelector=status(\"open\")",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorProblemCount, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorProblemCount, 0, expectedProblemsRequest))
}

// TestRetrieveMetricsFromFile_SLO tests the success case for file-based SLO SLIs.
func TestRetrieveMetricsFromFile_SLO(t *testing.T) {
	const (
		testDataFolder        = "./testdata/sli_files/slo_success/"
		testIndicatorSLOValue = "slo_value"
	)

	expectedSLORequest := buildSLORequest("7d07efde-b714-3e6e-ad95-08490e2540c4")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedSLORequest, testDataFolder+"slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorSLOValue: "SLO;7d07efde-b714-3e6e-ad95-08490e2540c4",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorSLOValue, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSLOValue, 95, expectedSLORequest))
}

// TestErrorMessageWhenNoSLIsAreRequested tests that the correct error message is generated when no SLIs are requested.
func TestErrorMessageWhenNoSLIsAreRequested(t *testing.T) {
	tests := []struct {
		name string
		slis map[string]string
	}{
		{
			name: "No SLIs requested and no SLIs defined",
		},
		{
			name: "No SLIs requested and a single SLI is defined",
			slis: map[string]string{
				"response_time_p95": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no need to have something here, because we should not send an API request
			handler := test.NewFileBasedURLHandler(t)

			rClient := newResourceClientMockWithSLIs(t, tt.slis)

			runGetSLIsFromFilesTestWithNoIndicatorsRequestedAndCheckSLIs(t, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("no metric", "no SLIs were requested"))
		})
	}
}

// TestGetSLIValueMetricsQuery_Success tests processing of Metrics API v2 results success case.
// One result, one data - want success
func TestGetSLIValueMetricsQuery_Success(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/success/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddStartsWith(expectedMetricsRequest, testDataFolder+"metrics_query_1result_1data_1value.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 287.10692602352884, expectedMetricsRequest))
}

// TestGetSLIValueMetricsQueryErrorHandling_RequestFails tests handling of failed requests.
func TestGetSLIValueMetricsQueryErrorHandling_RequestFails(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/constraints_violated/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddStartsWithError(expectedMetricsRequest, 400, testDataFolder+"metrics_query_constraints_violated.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, "error querying Metrics API v2"))
}

// TestGetSLIValueMetricsQuery_Warnings tests processing of Metrics API v2 results for warnings.
func TestGetSLIValueMetricsQuery_Warnings(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/warnings/"

	// TODO 2021-10-13: add rich error types as described in #358, including warnings
	tests := []struct {
		name                         string
		metricsQueryResponseFilename string
		expectedErrorSubString       string
	}{
		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "Zero metric series collections 1 - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_0results_fake3.json",
			expectedErrorSubString:       "Metrics API v2 returned zero metric series collections",
		},

		{
			name:                         "One metric series collection, no metric series - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_1result_0data.json",
			expectedErrorSubString:       "Metrics API v2 returned zero metric series",
		},

		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "One metric series collection, one metric sereis, no values, fake 1 - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_1result_1data_0values_fake1.json",
			expectedErrorSubString:       "Metrics API v2 returned zero values",
		},

		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "One metric series collection, one metric series, no values, fake 2 - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_1result_1data_0values_fake2.json",
			expectedErrorSubString:       "Metrics API v2 returned zero values",
		},

		{
			name:                         "One metric series collection, one metric series, null value - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_1result_1data_null_value.json",
			expectedErrorSubString:       "Metrics API v2 returned 'null' as value",
		},

		{
			name:                         "One metric series collection, one metric series, two values - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_1result_1data_2values.json",
			expectedErrorSubString:       "Metrics API v2 returned 2 values",
		},

		{
			name:                         "One metric series collection, two metric series - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_1result_2data.json",
			expectedErrorSubString:       "Metrics API v2 returned 2 metric series",
		},

		{
			name:                         "Two metric series collections, one metric series - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_2results_1data.json",
			expectedErrorSubString:       "Metrics API v2 returned 2 metric series collections",
		},

		{
			name:                         "Two metric series collections, two metric series - want failure",
			metricsQueryResponseFilename: testDataFolder + "metrics_query_2results_2data.json",
			expectedErrorSubString:       "Metrics API v2 returned 2 metric series collections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

			handler := test.NewFileBasedURLHandler(t)
			handler.AddStartsWith(dynatrace.MetricsQueryPath, tt.metricsQueryResponseFilename)

			rClient := newResourceClientMockWithSLIs(t, map[string]string{
				testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
			})

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventWarningAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, tt.expectedErrorSubString))
		})
	}
}

// tests the GETSliValue function to return the proper datapoint with the old custom query format
func TestGetSLIValueWithOldAndNewCustomQueryFormat(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/old_metrics_format/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("tag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29%2Ctag%28keptn_service%3Acarts%29%2Ctag%28keptn_deployment%3A%29%2Ctype%28SERVICE%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2850%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddStartsWith(expectedMetricsRequest, testDataFolder+"metrics_query.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 8433.40, expectedMetricsRequest))
}

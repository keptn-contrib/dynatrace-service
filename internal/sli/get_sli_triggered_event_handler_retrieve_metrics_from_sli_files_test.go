package sli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
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
	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			"response_time_p59": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		testSLOsWithResponseTimeP95,
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "SLI definition", "not found"))
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
	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricsSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		testSLOsWithResponseTimeP95,
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "error parsing Metrics v2 query"))
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
	configClient := newConfigClientMockThatErrorsGetSLIs(t, fmt.Errorf(errorMessage))

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, errorMessage))
}

// TestRetrieveMetricsFromFile_SecurityProblemsV2 tests the success case for file-based SecurityProblemsV2 SLIs.
func TestRetrieveMetricsFromFile_SecurityProblemsV2(t *testing.T) {
	const (
		testDataFolder                    = "./testdata/sli_files/secpv2_success/"
		testIndicatorSecurityProblemCount = "security_problem_count"
	)

	expectedSecurityProblemsRequest := buildSecurityProblemsRequest("status(\"open\")")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedSecurityProblemsRequest, filepath.Join(testDataFolder, "security_problems_status_open.json"))

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorSecurityProblemCount: "SECPV2;securityProblemSelector=status(\"open\")",
		},
		createTestSLOs(createTestSLOWithPassCriterion(testIndicatorSecurityProblemCount, "<=0")),
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorSecurityProblemCount, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSecurityProblemCount, 398, expectedSecurityProblemsRequest))
}

// TestRetrieveMetricsFromFile_ProblemsV2 tests the success case for file-based ProblemsV2 SLIs.
func TestRetrieveMetricsFromFile_ProblemsV2(t *testing.T) {
	const (
		testDataFolder            = "./testdata/sli_files/pv2_success/"
		testIndicatorProblemCount = "problem_count"
	)

	expectedProblemsRequest := buildProblemsV2Request("status(\"open\")")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedProblemsRequest, filepath.Join(testDataFolder, "problems_status_open.json"))

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorProblemCount: "PV2;problemSelector=status(\"open\")",
		},
		createTestSLOs(createTestSLOWithPassCriterion(testIndicatorProblemCount, "<=0")),
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorProblemCount, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorProblemCount, 30, expectedProblemsRequest))
}

// TestRetrieveMetricsFromFile_SLO tests the success case for file-based SLO SLIs.
func TestRetrieveMetricsFromFile_SLO(t *testing.T) {
	const (
		testDataFolder        = "./testdata/sli_files/slo_success/"
		testIndicatorSLOValue = "slo_value"
	)

	expectedSLORequest := buildSLORequest("7d07efde-b714-3e6e-ad95-08490e2540c4")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedSLORequest, filepath.Join(testDataFolder, "slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json"))

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorSLOValue: "SLO;7d07efde-b714-3e6e-ad95-08490e2540c4",
		},
		createTestSLOs(createTestSLOWithPassCriterion(testIndicatorSLOValue, "<=0")),
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorSLOValue, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSLOValue, 95, expectedSLORequest))
}

// TestErrorMessageWhenNoSLIsAreRequested tests that the correct error message is generated when no SLIs are requested.
func TestErrorMessageWhenNoSLIsAreRequested(t *testing.T) {
	tests := []struct {
		name string
		slis map[string]string
		slos *keptncommon.ServiceLevelObjectives
	}{
		{
			name: "No SLIs requested and no SLIs defined",
			slos: createTestSLOs(),
		},
		{
			name: "No SLIs requested and a single SLI is defined",
			slis: map[string]string{
				"response_time_p95": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
			},
			slos: createTestSLOs(createTestSLOWithPassCriterion(testIndicatorResponseTimeP95, "<=200")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no need to have something here, because we should not send an API request
			handler := test.NewFileBasedURLHandler(t)
			configClient := newConfigClientMockWithSLIsAndSLOs(t, tt.slis, tt.slos)
			runGetSLIsFromFilesTestWithNoIndicatorsRequestedAndCheckSLIs(t, handler, configClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorNoMetric, "no SLIs were requested"))
		})
	}
}

// TestGetSLIValueMetricsQuery_Success tests processing of Metrics API v2 results success case.
// One result, one data - want success
func TestGetSLIValueMetricsQuery_Success(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/success/"

	handler := test.NewCombinedURLHandler(t)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)"),
	)

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		testSLOsWithResponseTimeP95)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 31846.08512740705, expectedMetricsRequest))
}

// TestGetSLIValueMetricsQueryErrorHandling_RequestFails tests handling of failed requests.
func TestGetSLIValueMetricsQueryErrorHandling_RequestFails(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/constraints_violated/"

	expectedMetricsRequest := newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)").build()

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(expectedMetricsRequest, 400, filepath.Join(testDataFolder, "metrics_query_constraints_violated.json"))

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		testSLOsWithResponseTimeP95,
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, "error querying Metrics API v2"))
}

// TestGetSLIValueMetricsQuery_Warnings tests processing of Metrics API v2 query results that produce warnings.
// Many of these cases may never occur in reality but are included here for completeness. Variants are included for both the first and second metrics query responses.
func TestGetSLIValueMetricsQuery_Warnings(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/warnings/"

	requestBuilder := newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)")
	expectedMetricsRequest := requestBuilder.build()

	tests := []struct {
		name                   string
		expectedErrorSubString string
	}{
		{
			name:                   "zero_metric_series_collections_first",
			expectedErrorSubString: testErrorSubStringZeroMetricSeriesCollections,
		},
		{
			name:                   "zero_metric_series_collections_second",
			expectedErrorSubString: testErrorSubStringZeroMetricSeriesCollections,
		},

		{
			name:                   "one_metric_series_collection_zero_metric_series_first",
			expectedErrorSubString: testErrorSubStringZeroMetricSeries,
		},
		{
			name:                   "one_metric_series_collection_zero_metric_series_second",
			expectedErrorSubString: testErrorSubStringZeroMetricSeries,
		},

		{
			name:                   "one_metric_series_collection_one_metric_series_no_values_first",
			expectedErrorSubString: testErrorSubStringZeroValues,
		},
		{
			name:                   "one_metric_series_collection_one_metric_series_no_values_second",
			expectedErrorSubString: testErrorSubStringZeroValues,
		},

		{
			name:                   "one_metric_series_collection_one_metric_series_empty_values_first",
			expectedErrorSubString: testErrorSubStringZeroValues,
		},
		{
			name:                   "one_metric_series_collection_one_metric_series_empty_values_second",
			expectedErrorSubString: testErrorSubStringZeroValues,
		},

		{
			name:                   "one_metric_series_collection_one_metric_series_null_value_first",
			expectedErrorSubString: testErrorSubStringNullAsValue,
		},
		{
			name:                   "one_metric_series_collection_one_metric_series_null_value_second",
			expectedErrorSubString: testErrorSubStringNullAsValue,
		},

		{
			name:                   "one_metric_series_collection_one_metric_series_two_values_first_and_second",
			expectedErrorSubString: testErrorSubStringTwoValues,
		},
		{
			name:                   "one_metric_series_collection_one_metric_series_two_values_second",
			expectedErrorSubString: testErrorSubStringTwoValues,
		},

		{
			name:                   "one_metric_series_collection_two_metric_series_first",
			expectedErrorSubString: testErrorSubStringTwoMetricSeries,
		},
		{
			name:                   "one_metric_series_collection_two_metric_series_second",
			expectedErrorSubString: testErrorSubStringTwoMetricSeries,
		},

		{
			name:                   "two_metric_series_collections_one_metric_series_first",
			expectedErrorSubString: testErrorSubStringTwoMetricSeriesCollections,
		},
		{
			name:                   "two_metric_series_collections_one_metric_series_second",
			expectedErrorSubString: testErrorSubStringTwoMetricSeriesCollections,
		},

		{
			name:                   "two_metric_series_collections_two_metric_series_first",
			expectedErrorSubString: testErrorSubStringTwoMetricSeriesCollections,
		},
		{
			name:                   "two_metric_series_collections_two_metric_series_second",
			expectedErrorSubString: testErrorSubStringTwoMetricSeriesCollections,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)

			handler := test.NewCombinedURLHandler(t)
			addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler,
				testVariantDataFolder,
				requestBuilder,
			)

			configClient := newConfigClientMockWithSLIsAndSLOs(t,
				map[string]string{
					testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
				},
				testSLOsWithResponseTimeP95,
			)

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventWarningAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, tt.expectedErrorSubString))
		})
	}
}

// tests the GETSliValue function to return the proper datapoint with the old custom query format
func TestGetSLIValueWithOldCustomQueryFormat(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/old_metrics_format/"

	expectedMetricsRequest := newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)").copyWithEntitySelector("tag(keptn_project:sockshop),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:),type(SERVICE)").copyWithResolution(resolutionInf).build()

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query.json"))

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)",
		},
		testSLOsWithResponseTimeP95,
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 620.4411764705883, expectedMetricsRequest))
}

// Tests what happens when end time is too close to now. This test results in a short delay.
func TestGetSLISleep(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/sleep/"

	handler := test.NewCombinedURLHandler(t)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)"),
	)

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		testSLOsWithResponseTimeP95,
	)

	// time how long getting the SLI value takes
	timeBeforeGetSLIValue := time.Now()
	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 31846.08512740705, expectedMetricsRequest))
	getSLIExectutionTime := time.Since(timeBeforeGetSLIValue)

	assert.InDelta(t, 5, getSLIExectutionTime.Seconds(), 5)
}

// TestGetSLIValueSupportsEnvPlaceholders tests that environment variable placeholders are replaced correctly in SLI definitions.
func TestGetSLIValueSupportsEnvPlaceholders(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/env_placeholders/"

	handler := test.NewCombinedURLHandler(t)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time").copyWithEntitySelector("type(SERVICE),tag(\"env_tag:some_tag\")"),
	)

	indicator := "response_time_env"

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			indicator: "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"env_tag:$ENV.MY_ENV_TAG\")&metricSelector=builtin:service.response.time",
		},
		createTestSLOs(createTestSLOWithPassCriterion(indicator, "<=100")),
	)

	os.Setenv("MY_ENV_TAG", "some_tag")
	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, indicator, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(indicator, 1550.7739132118083, expectedMetricsRequest))
	os.Unsetenv("MY_ENV_TAG")
}

// TestGetSLIValueSupportsPlaceholders tests that placeholders are replaced correctly in SLI definitions.
func TestGetSLIValueSupportsPlaceholders(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/placeholders/"

	tests := []struct {
		name             string
		indicator        string
		query            string
		expectedRequest  string
		responseFilename string
		expectedSLIValue float64
	}{
		{
			name:             "Metrics V2 with MV2 encoding",
			indicator:        "response_time",
			query:            "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"keptn_managed\"),tag(\"keptn_project:$PROJECT\"),tag(\"keptn_stage:$STAGE\"),tag(\"keptn_service:$SERVICE\")&metricSelector=builtin:service.response.time&resolution=Inf",
			expectedRequest:  newMetricsV2QueryRequestBuilder("builtin:service.response.time").copyWithEntitySelector("type(SERVICE),tag(\"keptn_managed\"),tag(\"keptn_project:sockshop\"),tag(\"keptn_stage:staging\"),tag(\"keptn_service:carts\")").copyWithResolution(resolutionInf).build(),
			responseFilename: filepath.Join(testDataFolder, "metrics_query_result.json"),
			expectedSLIValue: 0.6458395061728395,
		},

		{
			name:             "Metrics V2",
			indicator:        "response_time2",
			query:            "entitySelector=type(SERVICE),tag(\"keptn_deployment:$DEPLOYMENT\"),tag(\"context:$CONTEXT\"),tag(\"keptn_stage:$STAGE\"),tag(\"keptn_service:$SERVICE\")&metricSelector=builtin:service.response.time&resolution=Inf",
			expectedRequest:  newMetricsV2QueryRequestBuilder("builtin:service.response.time").copyWithEntitySelector("type(SERVICE),tag(\"keptn_deployment:mydeployment\"),tag(\"context:mycontext\"),tag(\"keptn_stage:staging\"),tag(\"keptn_service:carts\")").copyWithResolution(resolutionInf).build(),
			responseFilename: filepath.Join(testDataFolder, "metrics_query_result.json"),
			expectedSLIValue: 645.8395061728395,
		},

		{
			name:             "PV2",
			indicator:        "problems",
			query:            "PV2;problemSelector=status($LABEL.problem_status)",
			expectedRequest:  buildProblemsV2Request("status(open)"),
			responseFilename: filepath.Join(testDataFolder, "problems_query_result.json"),
			expectedSLIValue: 12,
		},
		{
			name:             "SECPV2",
			indicator:        "security_problems",
			query:            "SECPV2;securityProblemSelector=status($LABEL.problem_status)",
			expectedRequest:  buildSecurityProblemsRequest("status(open)"),
			responseFilename: filepath.Join(testDataFolder, "security_problems_query_result.json"),
			expectedSLIValue: 414,
		},

		{
			name:             "SLO",
			indicator:        "RT_faster_500ms",
			query:            "SLO;$LABEL.slo_id",
			expectedRequest:  buildSLORequest("7d07efde-b714-3e6e-ad95-08490e2540c4"),
			responseFilename: filepath.Join(testDataFolder, "slo_query_result.json"),
			expectedSLIValue: 95,
		},

		{
			name:             "USQL",
			indicator:        "User_session_time",
			query:            "USQL;COLUMN_CHART;iOS 12.1.4;SELECT osVersion, AVG(duration) FROM usersession WHERE country IN('$LABEL.country') GROUP BY osVersion",
			expectedRequest:  buildUSQLRequest("SELECT osVersion, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY osVersion"),
			responseFilename: filepath.Join(testDataFolder, "usql_query_result.json"),
			expectedSLIValue: 29043,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(tt.expectedRequest, tt.responseFilename)

			keptnEvent := &getSLIEventData{
				context:      "mycontext",
				event:        "myevent",
				project:      "sockshop",
				stage:        "staging",
				service:      "carts",
				deployment:   "mydeployment",
				testStrategy: "mystrategy",
				labels: map[string]string{
					"slo_id":         "7d07efde-b714-3e6e-ad95-08490e2540c4",
					"problem_status": "open",
					"country":        "Austria",
				},

				indicators: []string{tt.indicator},
				sliStart:   testSLIStart,
				sliEnd:     testSLIEnd,
			}

			configClient := newConfigClientMockWithSLIsAndSLOs(t,
				map[string]string{
					tt.indicator: tt.query,
				},
				createTestSLOs(createTestSLOWithPassCriterion(tt.indicator, "<=100")),
			)

			runGetSLIsFromFilesTestWithEventAndCheckSLIs(t, handler, configClient, keptnEvent, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(tt.indicator, tt.expectedSLIValue, tt.expectedRequest))
		})
	}
}

// TestGetSLIValueMetricsQuery_SuccessWithFold tests processing of Metrics API v2 results success case using a fold rather than resolution Inf.
func TestGetSLIValueMetricsQuery_SuccessWithFold(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/success_with_fold/"

	const testIndicatorAvailability = "availability"

	handler := test.NewCombinedURLHandler(t)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithFold(handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:pgi.availability:splitBy()"),
	)

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorAvailability: "metricSelector=builtin:pgi.availability:splitBy()",
		},
		createTestSLOs(createTestSLOWithPassCriterion(testIndicatorAvailability, "<=100")),
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorAvailability, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorAvailability, 97.46884534911891, expectedMetricsRequest))
}

// TestGetSLIValueMetricsQuery_NoFoldPossible tests processing of Metrics API v2 results where a fold is not possible.
func TestGetSLIValueMetricsQuery_NoFoldPossible(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/no_fold_possible/"

	const testIndicatorAvailability = "availability"

	requestBuilder := newMetricsV2QueryRequestBuilder("builtin:pgi.availability:splitBy():avg")
	expectedMetricsRequest := requestBuilder.build()

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query1.json"))
	handler.AddExact(buildMetricsV2DefinitionRequestString(requestBuilder.metricSelector()), filepath.Join(testDataFolder, "metrics_get_id.json"))

	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorAvailability: "metricSelector=builtin:pgi.availability:splitBy():avg",
		},
		createTestSLOs(createTestSLOWithPassCriterion(testIndicatorAvailability, "<=100")),
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorAvailability, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorAvailability, expectedMetricsRequest, "unable to apply ':fold()'"))
}

// TestGetSLIValueMetricsQuery_SuccessWithResolutionInfProvided tests processing of Metrics API v2 results where resolution is explicitly set to Inf in different forms.
func TestGetSLIValueMetricsQuery_SuccessWithResolutionInfProvided(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/explicit_resolution_inf/"

	const testIndicatorResponseTime = "response_time"

	tests := []struct {
		name                 string
		resolutionInfVariant string
	}{
		{
			name:                 "all_lower_case",
			resolutionInfVariant: "inf",
		},
		{
			name:                 "all_upper_case",
			resolutionInfVariant: "INF",
		},
		{
			name:                 "just_capital_i",
			resolutionInfVariant: "Inf",
		},
		{
			name:                 "just_lower_case__i",
			resolutionInfVariant: "iNF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)

			requestBuilder := newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy()").copyWithResolution(tt.resolutionInfVariant)
			expectedMetricsRequest := requestBuilder.build()

			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(expectedMetricsRequest, filepath.Join(testVariantDataFolder, "metrics_get_by_query1.json"))

			configClient := newConfigClientMockWithSLIsAndSLOs(t,
				map[string]string{
					testIndicatorResponseTime: "metricSelector=builtin:service.response.time:splitBy()&resolution=" + tt.resolutionInfVariant,
				},
				createTestSLOs(createTestSLOWithPassCriterion(testIndicatorResponseTime, "<=100")),
			)

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTime, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTime, 54896.50418867919, expectedMetricsRequest))
		})
	}
}

// TestGetSLIValueMetricsQuery_SuccessWithOtherResolution tests processing of Metrics API v2 results success case using a fold due to an explicit resolution being set.
func TestGetSLIValueMetricsQuery_SuccessWithOtherResolution(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/success_with_other_resolution/"

	tests := []struct {
		name             string
		metricSelector   string
		expectedSLIValue float64
	}{
		{
			name:             "response_time",
			metricSelector:   "builtin:service.response.time:splitBy()",
			expectedSLIValue: 54896.504858650806,
		},
		{
			name:             "availability",
			metricSelector:   "builtin:pgi.availability:splitBy()",
			expectedSLIValue: 97.47250403469201,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)
			testIndicator := tt.name

			handler := test.NewCombinedURLHandler(t)
			expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithFold(handler,
				testVariantDataFolder,
				newMetricsV2QueryRequestBuilder(tt.metricSelector).copyWithResolution("30m"),
			)

			configClient := newConfigClientMockWithSLIsAndSLOs(t, map[string]string{
				testIndicator: "metricSelector=" + tt.metricSelector + "&resolution=30m",
			},
				createTestSLOs(createTestSLOWithPassCriterion(testIndicator, "<=100")),
			)

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicator, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicator, tt.expectedSLIValue, expectedMetricsRequest))
		})
	}

}

// TestGetSLIValueMetricsQuery_DuplicateNamesAreNotAllowedInOneSLIFile tests that duplicate SLI names are not allowed within one dynatrace/sli.yaml file.
func TestGetSLIValueMetricsQuery_DuplicateNamesAreNotAllowedInOneSLIFile(t *testing.T) {
	resourceClient := &slisOnServiceLevelResourceClientMock{
		t: t,
		serviceSLIFile: `spec_version: "1.0"
indicators:
  response_time_p95: metricSelector=builtin:service.response.time:merge("dt.entity.service"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)
  response_time_p95: metricSelector=builtin:service.response.time:merge("dt.entity.service"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)
`,
	}

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, test.NewEmptyURLHandler(t), keptn.NewConfigClient(resourceClient), testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "mapping key \"response_time_p95\" already defined"))
}

const expectedSLIResourceURI = "dynatrace/sli.yaml"

// slisOnServiceLevelResourceClientMock is an implementation of keptn.ResourceClient that only provides a dynatrace/sli.yaml file on the service level.
type slisOnServiceLevelResourceClientMock struct {
	t              *testing.T
	serviceSLIFile string
}

func (src *slisOnServiceLevelResourceClientMock) GetProjectResource(_ context.Context, project string, resourceURI string) (string, error) {
	assert.EqualValues(src.t, testProject, project)
	assert.EqualValues(src.t, expectedSLIResourceURI, resourceURI)
	return "", keptn.NewResourceNotFoundError(resourceURI, project, "", "")
}

func (src *slisOnServiceLevelResourceClientMock) GetStageResource(_ context.Context, project string, stage string, resourceURI string) (string, error) {
	assert.EqualValues(src.t, testProject, project)
	assert.EqualValues(src.t, testStage, stage)
	assert.EqualValues(src.t, expectedSLIResourceURI, resourceURI)
	return "", keptn.NewResourceNotFoundError(resourceURI, project, stage, "")
}

func (src *slisOnServiceLevelResourceClientMock) GetServiceResource(_ context.Context, project string, stage string, service string, resourceURI string) (string, error) {
	assert.EqualValues(src.t, testProject, project)
	assert.EqualValues(src.t, testStage, stage)
	assert.EqualValues(src.t, testService, service)
	assert.EqualValues(src.t, expectedSLIResourceURI, resourceURI)
	return src.serviceSLIFile, nil
}

func (src *slisOnServiceLevelResourceClientMock) GetResource(ctx context.Context, project string, stage string, service string, resourceURI string) (string, error) {
	src.t.Fatal("GetResource() should not be needed in this mock!")
	return "", nil
}

func (src *slisOnServiceLevelResourceClientMock) UploadResource(ctx context.Context, contentToUpload []byte, remoteResourceURI string, project string, stage string, service string) error {
	src.t.Fatal("UploadResource() should not be needed in this mock!")
	return nil
}

const expectedSLOResourceURI = "slo.yaml"

// TestCustomSLIsGivesErrorIfNoSLOFileExistsButIndicatorRequested tests that an error is produced if no slo.yaml file exists for a sli.yaml-based request if even indicators are requested (unlikely but possible).
func TestCustomSLIsGivesErrorIfNoSLOFileExistsButIndicatorRequested(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/no_slo_file"

	handler := test.NewCombinedURLHandler(t)
	_ = addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)"),
	)

	configClient := newConfigClientMockWithSLIsThatErrorsGetSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		keptn.NewResourceNotFoundError(expectedSLOResourceURI, testProject, testStage, testService),
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "could not retrieve SLO definitions"))
}

// TestGetSLIValueMetricsQuery_NoDataForInformationalSLOFromFileProducesWarning tests that informational SLOs with no data produce an overall warning result.
func TestGetSLIValueMetricsQuery_NoDataForInformationalSLOFromFileProducesWarning(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/no_data_informational_slo"

	const testIndicatorRequestCount = "request_count"
	responseTimeQueryBuilder := newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)")
	requestCountQueryBuilder := newMetricsV2QueryRequestBuilder("builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum")

	configClient := newConfigClientMockWithSLIsAndSLOs(
		t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=" + responseTimeQueryBuilder.metricSelector(),
			testIndicatorRequestCount:    " metricSelector=" + requestCountQueryBuilder.metricSelector(),
		},
		createTestSLOs(createTestSLOWithPassCriterion(testIndicatorResponseTimeP95, "<=90"), createTestInformationalSLO(testIndicatorRequestCount)))

	handler := test.NewCombinedURLHandler(t)
	expectedResponseTimeQuery := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, filepath.Join(testDataFolder, "response_time_p95"), responseTimeQueryBuilder)
	expectedRequestCountQuery := requestCountQueryBuilder.build()
	handler.AddExactFile(expectedRequestCountQuery, filepath.Join(filepath.Join(testDataFolder, "request_count"), "metrics_get_by_query1.json"))

	runGetSLIsFromFilesTestAndCheckSLIs(t, handler, configClient, []string{testIndicatorResponseTimeP95, testIndicatorRequestCount}, getSLIFinishedEventWarningAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 210597.99693868207, expectedResponseTimeQuery), createFailedSLIResultWithQueryAssertionsFunc(testIndicatorRequestCount, expectedRequestCountQuery, testErrorSubStringZeroMetricSeries))
}

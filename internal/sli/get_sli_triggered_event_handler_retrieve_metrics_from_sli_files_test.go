package sli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
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
	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		"response_time_p59": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

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
	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricsSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

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
		securityProblemsRequest           = "/api/v2/securityProblems?from=1609459200000&securityProblemSelector=status%28%22open%22%29&to=1609545600000"
		testDataFolder                    = "./testdata/sli_files/secpv2_success/"
		testIndicatorSecurityProblemCount = "security_problem_count"
	)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(securityProblemsRequest, filepath.Join(testDataFolder, "security_problems_status_open.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorSecurityProblemCount: "SECPV2;securityProblemSelector=status(\"open\")",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorSecurityProblemCount, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSecurityProblemCount, 103, securityProblemsRequest))
}

// TestRetrieveMetricsFromFile_ProblemsV2 tests the success case for file-based ProblemsV2 SLIs.
func TestRetrieveMetricsFromFile_ProblemsV2(t *testing.T) {
	const (
		testDataFolder            = "./testdata/sli_files/pv2_success/"
		testIndicatorProblemCount = "problem_count"
	)

	expectedProblemsRequest := buildProblemsV2Request("status%28%22open%22%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedProblemsRequest, filepath.Join(testDataFolder, "problems_status_open.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorProblemCount: "PV2;problemSelector=status(\"open\")",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorProblemCount, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorProblemCount, 0, expectedProblemsRequest))
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

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorSLOValue: "SLO;7d07efde-b714-3e6e-ad95-08490e2540c4",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorSLOValue, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSLOValue, 95, expectedSLORequest))
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

			configClient := newConfigClientMockWithSLIs(t, tt.slis)

			runGetSLIsFromFilesTestWithNoIndicatorsRequestedAndCheckSLIs(t, handler, configClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("no metric", "no SLIs were requested"))
		})
	}
}

// TestGetSLIValueMetricsQuery_Success tests processing of Metrics API v2 results success case.
// One result, one data - want success
func TestGetSLIValueMetricsQuery_Success(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/success/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_1result_1data_1value.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 287.10692602352884, expectedMetricsRequest))
}

// TestGetSLIValueMetricsQueryErrorHandling_RequestFails tests handling of failed requests.
func TestGetSLIValueMetricsQueryErrorHandling_RequestFails(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/constraints_violated/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(expectedMetricsRequest, 400, filepath.Join(testDataFolder, "metrics_query_constraints_violated.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, "error querying Metrics API v2"))
}

// TestGetSLIValueMetricsQuery_Warnings tests processing of Metrics API v2 results for warnings.
func TestGetSLIValueMetricsQuery_Warnings(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/warnings/"

	// TODO 2021-10-13: add rich error types as described in #358, including warnings
	tests := []struct {
		name                         string
		metricsQueryResponseFilename string
		expectedErrorSubString       string
	}{
		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "Zero metric series collections 1 - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_0results_fake3.json"),
			expectedErrorSubString:       "Metrics API v2 returned zero metric series collections",
		},

		{
			name:                         "One metric series collection, no metric series - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_1result_0data.json"),
			expectedErrorSubString:       "Metrics API v2 returned zero metric series",
		},

		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "One metric series collection, one metric sereis, no values, fake 1 - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_1result_1data_0values_fake1.json"),
			expectedErrorSubString:       "Metrics API v2 returned zero values",
		},

		// this case may not occur in reality, but check it here for completeness
		{
			name:                         "One metric series collection, one metric series, no values, fake 2 - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_1result_1data_0values_fake2.json"),
			expectedErrorSubString:       "Metrics API v2 returned zero values",
		},

		{
			name:                         "One metric series collection, one metric series, null value - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_1result_1data_null_value.json"),
			expectedErrorSubString:       "Metrics API v2 returned 'null' as value",
		},

		{
			name:                         "One metric series collection, one metric series, two values - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_1result_1data_2values.json"),
			expectedErrorSubString:       "Metrics API v2 returned 2 values",
		},

		{
			name:                         "One metric series collection, two metric series - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_1result_2data.json"),
			expectedErrorSubString:       "Metrics API v2 returned 2 metric series",
		},

		{
			name:                         "Two metric series collections, one metric series - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_2results_1data.json"),
			expectedErrorSubString:       "Metrics API v2 returned 2 metric series collections",
		},

		{
			name:                         "Two metric series collections, two metric series - want failure",
			metricsQueryResponseFilename: filepath.Join(testDataFolder, "metrics_query_2results_2data.json"),
			expectedErrorSubString:       "Metrics API v2 returned 2 metric series collections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(expectedMetricsRequest, tt.metricsQueryResponseFilename)

			configClient := newConfigClientMockWithSLIs(t, map[string]string{
				testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
			})

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventWarningAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, tt.expectedErrorSubString))
		})
	}
}

// tests the GETSliValue function to return the proper datapoint with the old custom query format
func TestGetSLIValueWithOldAndNewCustomQueryFormat(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/old_metrics_format/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("tag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29%2Ctag%28keptn_service%3Acarts%29%2Ctag%28keptn_deployment%3A%29%2Ctype%28SERVICE%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2850%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 8433.40, expectedMetricsRequest))
}

// Tests what happens when end time is too close to now. This test results in a short delay.
func TestGetSLISleep(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/sleep/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("tag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29%2Ctag%28keptn_service%3Acarts%29%2Ctag%28keptn_deployment%3A%29%2Ctype%28SERVICE%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2850%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)",
	})

	// time how long getting the SLI value takes
	timeBeforeGetSLIValue := time.Now()
	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 8433.40, expectedMetricsRequest))
	getSLIExectutionTime := time.Since(timeBeforeGetSLIValue)

	assert.InDelta(t, 5, getSLIExectutionTime.Seconds(), 5)
}

// TestGetSLIValueSupportsEnvPlaceholders tests that environment variable placeholders are replaced correctly in SLI definitions.
func TestGetSLIValueSupportsEnvPlaceholders(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/env_placeholders/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22env_tag%3Asome_tag%22%29", "builtin%3Aservice.response.time")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_query_result.json"))

	indicator := "response_time_env"

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		indicator: "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"env_tag:$ENV.MY_ENV_TAG\")&metricSelector=builtin:service.response.time",
	})

	os.Setenv("MY_ENV_TAG", "some_tag")
	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, indicator, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(indicator, 0.29, expectedMetricsRequest))
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
			query:            "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"keptn_managed\"),tag(\"keptn_project:$PROJECT\"),tag(\"keptn_stage:$STAGE\"),tag(\"keptn_service:$SERVICE\")&metricSelector=builtin:service.response.time",
			expectedRequest:  buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22keptn_managed%22%29%2Ctag%28%22keptn_project%3Amyproject%22%29%2Ctag%28%22keptn_stage%3Amystage%22%29%2Ctag%28%22keptn_service%3Amyservice%22%29", "builtin%3Aservice.response.time"),
			responseFilename: filepath.Join(testDataFolder, "metrics_query_result.json"),
			expectedSLIValue: 0.29,
		},

		{
			name:             "Metrics V2",
			indicator:        "response_time2",
			query:            "entitySelector=type(SERVICE),tag(\"keptn_deployment:$DEPLOYMENT\"),tag(\"context:$CONTEXT\"),tag(\"keptn_stage:$STAGE\"),tag(\"keptn_service:$SERVICE\")&metricSelector=builtin:service.response.time",
			expectedRequest:  buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22keptn_deployment%3Amydeployment%22%29%2Ctag%28%22context%3Amycontext%22%29%2Ctag%28%22keptn_stage%3Amystage%22%29%2Ctag%28%22keptn_service%3Amyservice%22%29", "builtin%3Aservice.response.time"),
			responseFilename: filepath.Join(testDataFolder, "metrics_query_result.json"),
			expectedSLIValue: 290,
		},

		{
			name:             "PV2",
			indicator:        "problems",
			query:            "PV2;problemSelector=status($LABEL.problem_status)",
			expectedRequest:  buildProblemsV2Request("status%28open%29"),
			responseFilename: filepath.Join(testDataFolder, "problems_query_result.json"),
			expectedSLIValue: 1,
		},
		{
			name:             "SECPV2",
			indicator:        "security_problems",
			query:            "SECPV2;securityProblemSelector=status($LABEL.problem_status)",
			expectedRequest:  buildSecurityProblemsRequest("status%28open%29"),
			responseFilename: filepath.Join(testDataFolder, "security_problems_query_result.json"),
			expectedSLIValue: 4,
		},

		{
			name:             "SLO",
			indicator:        "RT_faster_500ms",
			query:            "SLO;$LABEL.slo_id",
			expectedRequest:  buildSLORequest("524ca177-849b-3e8c-8175-42b93fbc33c5"),
			responseFilename: filepath.Join(testDataFolder, "slo_query_result.json"),
			expectedSLIValue: 96,
		},

		{
			name:             "USQL",
			indicator:        "User_session_time",
			query:            "USQL;COLUMN_CHART;iOS 12.1.4;SELECT osVersion, AVG(duration) FROM usersession WHERE country IN('$LABEL.country') GROUP BY osVersion",
			expectedRequest:  buildUSQLRequest("SELECT+osVersion%2C+AVG%28duration%29+FROM+usersession+WHERE+country+IN%28%27Austria%27%29+GROUP+BY+osVersion"),
			responseFilename: filepath.Join(testDataFolder, "usql_query_result.json"),
			expectedSLIValue: 21478,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(tt.expectedRequest, tt.responseFilename)

			keptnEvent := &getSLIEventData{
				context:      "mycontext",
				event:        "myevent",
				project:      "myproject",
				stage:        "mystage",
				service:      "myservice",
				deployment:   "mydeployment",
				testStrategy: "mystrategy",
				labels: map[string]string{
					"slo_id":         "524ca177-849b-3e8c-8175-42b93fbc33c5",
					"problem_status": "open",
					"country":        "Austria",
				},

				indicators: []string{tt.indicator},
				sliStart:   testSLIStart,
				sliEnd:     testSLIEnd,
			}

			configClient := newConfigClientMockWithSLIs(t, map[string]string{
				tt.indicator: tt.query,
			})

			runGetSLIsFromFilesTestWithEventAndCheckSLIs(t, handler, configClient, keptnEvent, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(tt.indicator, tt.expectedSLIValue, tt.expectedRequest))
		})
	}
}

package sli

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 0 results and a warning
//   - e.g. misspelled dimension key in merge transformation
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsNoResultsAndWarning(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/no_results_due_to_entity_type"

	expectedMetricsRequest := // error here: merge(dt.entity.services)
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.services\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)").build()

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_0_result_warning_entity-type.json"))

	// error here as well: merge("dt.entity.services")
	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.services\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		testSLOsWithResponseTimeP95,
	)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventWarningAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, testErrorSubStringZeroMetricSeries, "Warning"))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 0 results and no warning
//   - e.g. misspelled tag name
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsNoResults(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/no_results_due_to_tag"

	expectedMetricsRequest := newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:stagin)").build()

	// error here: tag(keptn_project:stagin)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_0_result_wrong-tag.json"))

	// error here as well: tag(keptn_project:stagin)
	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:stagin)",
		},
		testSLOsWithResponseTimeP95,
	)

	sliResultAssertionsFunc := func(t *testing.T, actual sliResult) {
		createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, testErrorSubStringZeroMetricSeries)(t, actual)
		assert.NotContains(t, actual.Message, "Warning")
	}

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventWarningAssertionsFunc, sliResultAssertionsFunc)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 3 results instead of 1 and no warning
//   - e.g. missing merge('dimension_key') transformation
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsMultipleResults(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/multiple_results"

	expectedMetricsRequest := newMetricsV2QueryRequestBuilder("builtin:service.response.time:percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)").build()

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_3_results.json"))

	// error here as well: missing merge("dt.entity.service) transformation
	configClient := newConfigClientMockWithSLIsAndSLOs(t,
		map[string]string{
			testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
		testSLOsWithResponseTimeP95,
	)

	sliResultAssertionsFunc := func(t *testing.T, actual sliResult) {
		createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, "3 metric series")(t, actual)
		assert.NotContains(t, actual.Message, "Warning")
	}

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventWarningAssertionsFunc, sliResultAssertionsFunc)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but the MV2 prefix is used incorrectly, so we return an error for that
//   - e.g. MV2;MicroSeconds;<query>
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryIsUsingWrongMetricUnit(t *testing.T) {
	testConfigs := []struct {
		name      string
		mv2Prefix string
	}{
		{
			name:      "unit Percent fails",
			mv2Prefix: "MV2;Percent;",
		},
		{
			name:      "unit MicroSeconds fails",
			mv2Prefix: "MV2;MicroSeconds;",
		},
		{
			name:      "unit Bytes fails",
			mv2Prefix: "MV2;Bytes;",
		},
		{
			name:      "missing unit fails",
			mv2Prefix: "MV2;",
		},
		{
			name:      "missing unit fails 2",
			mv2Prefix: "MV2;;",
		},
	}
	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {

			// no handler needed
			handler := test.NewFileBasedURLHandler(t)

			// error here: in value of tc.mv2Prefix
			configClient := newConfigClientMockWithSLIsAndSLOs(t,
				map[string]string{
					testIndicatorResponseTimeP95: tc.mv2Prefix + "metricSelector=builtin:service.response.time:percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
				},
				testSLOsWithResponseTimeP95,
			)

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "error parsing MV2 query"))
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists but there are no SLIs defined OR
// * there is no 'dynatrace/sli.yaml' file
//   - currently this would lead to a fallback for default SLI definitions
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreDefinedButEmpty(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/fallback_to_defaults"

	// fallback: mind the default SLI definitions in the URL below
	handler := test.NewCombinedURLHandler(t)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)").copyWithEntitySelector("type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:)"),
	)

	// no custom queries defined here
	// currently this could have 2 reasons: EITHER no sli.yaml file available OR no indicators defined in such a file)
	// TODO 2021-09-29: we should be able to differentiate between 'not there' and 'no SLIs defined' - the latter could be intentional
	configClient := &getSLIsAndGetSLOsConfigClientMock{
		t:    t,
		slis: nil, // no SLIs are defined
		slos: createTestSLOs(createTestSLOWithPassCriterion(testIndicatorResponseTimeP95, "<600")),
	}

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 857.6499999999999, expectedMetricsRequest))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomSLIsAreUsedWhenSpecified(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/used_if_defined"

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

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 31846.08512740705, expectedMetricsRequest))
}

// TestCustomSLIMetricsV1Parsing tests that Metrics queries with '=' in either entity or metric selectors are parsed as expected.
func TestCustomSLIMetricsV1Parsing(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/query_parsing/"

	tests := []struct {
		name           string
		metricSelector string
		entitySelector string
	}{
		{
			name:           "equals_in_metric_selector",
			metricSelector: "builtin:service.response.time:filter(and(or(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"specialTag:key=value~\")\"))))):splitBy():percentile(95.0)",
			entitySelector: "type(service)",
		},
		{
			name:           "equals_in_entity_selector",
			metricSelector: "builtin:service.response.time:splitBy():percentile(95.0)",
			entitySelector: "type(service),tag(\"specialTag:key=value\")",
		},
		{
			name:           "equals_in_metric_and_entity_selectors",
			metricSelector: "builtin:service.response.time:filter(and(or(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"specialTag:key=value~\")\"))))):splitBy():percentile(95.0)",
			entitySelector: "type(service),tag(\"specialTag:key=value\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testVariantDataFolder := path.Join(testDataFolder, tt.name)

			configClient := newConfigClientMockWithSLIsAndSLOs(t,
				map[string]string{
					testIndicatorResponseTimeP95: fmt.Sprintf("metricSelector=%s&entitySelector=%s", tt.metricSelector, tt.entitySelector),
				},
				testSLOsWithResponseTimeP95,
			)

			handler := test.NewCombinedURLHandler(t)
			expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler,
				testVariantDataFolder,
				newMetricsV2QueryRequestBuilder(tt.metricSelector).copyWithEntitySelector(tt.entitySelector),
			)

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 12910.279946732833, expectedMetricsRequest))
		})
	}
}

// TestGetSLIValueFromCustomQueriesWithLegacyQueryFormatWorkAsExpected tests that valid queries using the legacy query format work as expected.
func TestGetSLIValueFromCustomQueriesWithLegacyQueryFormatWorkAsExpected(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/legacy_format_success/"

	tests := []struct {
		name                   string
		query                  string
		expectedMetricsRequest string
		expectedSLIResultValue float64
	}{
		{
			name:                   "with_scope_key",
			query:                  "builtin:service.requestCount.total:splitBy():percentile(95.0)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)",
			expectedMetricsRequest: newMetricsV2QueryRequestBuilder("builtin:service.requestCount.total:splitBy():percentile(95.0)").copyWithEntitySelector("tag(keptn_project:sockshop),tag(keptn_stage:staging),tag(keptn_service:carts),tag(keptn_deployment:),type(SERVICE)").copyWithResolution(resolutionInf).build(),
			expectedSLIResultValue: 324,
		},
		{
			name:                   "finishing_with_question_mark",
			query:                  "builtin:service.response.time:splitBy():percentile(95.0)?",
			expectedMetricsRequest: newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():percentile(95.0)").copyWithResolution(resolutionInf).build(),
			expectedSLIResultValue: 210597.99593297063,
		},
		{
			name:                   "just_metric_selector",
			query:                  "builtin:service.response.time:splitBy():percentile(95.0)",
			expectedMetricsRequest: newMetricsV2QueryRequestBuilder("builtin:service.response.time:splitBy():percentile(95.0)").copyWithResolution(resolutionInf).build(),
			expectedSLIResultValue: 210597.99593297063,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testVariantDataFolder := path.Join(testDataFolder, tt.name)

			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(tt.expectedMetricsRequest, path.Join(testVariantDataFolder, "metrics_get_by_query1.json"))

			configClient := newConfigClientMockWithSLIsAndSLOs(t,
				map[string]string{
					testIndicatorResponseTimeP95: tt.query,
				},
				testSLOsWithResponseTimeP95,
			)

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, tt.expectedSLIResultValue, tt.expectedMetricsRequest))
		})
	}
}

// TestGetSLIValueFromInvalidCustomQueriesWithLegacyQueryFormatFailAsExpected tests that invalid legacy queries fail as expected.
func TestGetSLIValueFromInvalidCustomQueriesWithLegacyQueryFormatFailAsExpected(t *testing.T) {
	const metricV2ParsingErrorSubstring = "error parsing Metrics v2 query"
	const keyValuePairErrorSubstring = "could not parse 'key=value' pair"
	const unknownKeyErrorSubstring = "unknown key"

	tests := []struct {
		name                    string
		query                   string
		expectedErrorSubstrings []string
	}{
		{
			name:                    "missing_scope_value",
			query:                   "builtin:service.requestCount.total:splitBy():percentile(95.0)?scope=",
			expectedErrorSubstrings: []string{metricV2ParsingErrorSubstring, keyValuePairErrorSubstring},
		},
		{
			name:                    "missing_scope_value_and_equals",
			query:                   "builtin:service.requestCount.total:splitBy():percentile(95.0)?scope",
			expectedErrorSubstrings: []string{metricV2ParsingErrorSubstring, keyValuePairErrorSubstring},
		},

		// this will fail as it is considered ambiguous: it technically has a key-value pair, but it is unknown
		{
			name:                    "equals_in_metric",
			query:                   "builtin:service.response.time:filter(and(or(in(\"dt.entity.service\",entitySelector(\"type(service),tag(~\"specialTag:key=value~\")\"))))):splitBy():percentile(95.0)",
			expectedErrorSubstrings: []string{metricV2ParsingErrorSubstring, unknownKeyErrorSubstring},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			configClient := newConfigClientMockWithSLIsAndSLOs(t,
				map[string]string{
					testIndicatorResponseTimeP95: tt.query,
				},
				testSLOsWithResponseTimeP95,
			)

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, test.NewEmptyURLHandler(t), configClient, testIndicatorResponseTimeP95, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, tt.expectedErrorSubstrings...))
		})
	}
}

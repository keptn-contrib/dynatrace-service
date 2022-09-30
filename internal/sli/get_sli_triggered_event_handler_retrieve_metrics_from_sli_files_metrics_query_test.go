package sli

import (
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
	const testDataFolder = "./testdata/sli_files/metrics/no_results_due_to_entity_selector"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.services%22%29%3Apercentile%2895%29")

	// error here: merge(dt.entity.services)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_0_result_warning_entity-selector.json"))

	// error here as well: merge("dt.entity.services")
	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.services\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventWarningAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, "zero metric series", "Warning"))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 0 results and no warning
//   - e.g. misspelled tag name
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsNoResults(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/metrics/no_results_due_to_tag"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astagin%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

	// error here: tag(keptn_project:stagin)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_0_result_wrong-tag.json"))

	// error here as well: tag(keptn_project:stagin)
	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:stagin)",
	})

	sliResultAssertionsFunc := func(t *testing.T, actual sliResult) {
		createFailedSLIResultWithQueryAssertionsFunc(testIndicatorResponseTimeP95, expectedMetricsRequest, "zero metric series")(t, actual)
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

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Apercentile%2895%29")

	// error here: missing merge("dt.entity.service) transformation
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_3_results.json"))

	// error here as well: missing merge("dt.entity.service) transformation
	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

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
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			// no handler needed
			handler := test.NewFileBasedURLHandler(t)

			// error here: in value of tc.mv2Prefix
			configClient := newConfigClientMockWithSLIs(t, map[string]string{
				testIndicatorResponseTimeP95: tc.mv2Prefix + "metricSelector=builtin:service.response.time:percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
			})

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

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29%2Ctag%28keptn_service%3Acarts%29%2Ctag%28keptn_deployment%3A%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

	// fallback: mind the default SLI definitions in the URL below
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_1_result_defaults.json"))

	// no custom queries defined here
	// currently this could have 2 reasons: EITHER no sli.yaml file available OR no indicators defined in such a file)
	// TODO 2021-09-29: we should be able to differentiate between 'not there' and 'no SLIs defined' - the latter could be intentional
	configClient := newConfigClientMockWithNoSLIsOrError(t)

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 857.6499999999999, expectedMetricsRequest))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomSLIsAreUsedWhenSpecified(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/basic/used_if_defined"

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29", "builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedMetricsRequest, filepath.Join(testDataFolder, "response_time_p95_200_1_result.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorResponseTimeP95, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorResponseTimeP95, 12439.619479902443, expectedMetricsRequest))
}

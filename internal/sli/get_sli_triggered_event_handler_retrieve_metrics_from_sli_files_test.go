package sli

import (
	"fmt"
	"net/http"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const indicator = "response_time_p95"

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI could not be found because of a misspelled indicator name - e.g. 'response_time_p59' instead of 'response_time_p95'
//   - this would have lead to a fallback to default SLIs, but should return an error now.
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButIndicatorCannotBeMatched(t *testing.T) {
	// no need to have something here, because we should not send an API request
	handler := test.NewFileBasedURLHandler(t)

	// error here in the misspelled indicator:
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			"response_time_p59": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "response_time_p95")
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, true)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but Dynatrace cannot process the query correctly and returns a 400 error
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryIsNotValid(t *testing.T) {
	// error here: metric(s)Selector=
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExactError(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricsSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		400,
		"./testdata/response_time_p95_400_constraints-violated.json")

	// error here as well: metric(s)Selector=
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricsSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "metricSelector to be specified")
		assert.Contains(t, actual.Message, "400")
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, true)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI has errors, so parsing the YAML file would not be possible
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreInvalidYAML(t *testing.T) {
	// make sure we would not be able to query any metric due to a parsing error
	handler := test.NewFileBasedURLHandler(t)

	const errorMessage = "invalid YAML file - some parsing issue"
	kClient := &keptnClientMock{
		customQueriesError: fmt.Errorf(errorMessage),
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, errorMessage)
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, true)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 0 results and a warning
//   - e.g. misspelled dimension key in merge transformation
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsNoResultsAndWarning(t *testing.T) {
	// error here: merge(dt.entity.services)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.services%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_0_result_warning_entity-selector.json")

	// error here as well: merge("dt.entity.services")
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:merge(\"dt.entity.services\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "no result values")
		assert.Contains(t, actual.Message, "Warning")
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, true)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 0 results and no warning
//	 - e.g. misspelled tag name
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsNoResults(t *testing.T) {
	// error here: tag(keptn_project:stagin)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astagin%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_0_result_wrong-tag.json")

	// error here as well: tag(keptn_project:stagin)
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:stagin)",
		},
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "no result values")
		assert.NotContains(t, actual.Message, "Warning")
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, true)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 3 results instead of 1 and no warning
//	 - e.g. missing merge('dimension_key') transformation
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsMultipleResults(t *testing.T) {
	// error here: missing merge("dt.entity.service) transformation
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_3_results.json")

	// error here as well: missing merge("dt.entity.service) transformation
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "returned 3 result values")
		assert.NotContains(t, actual.Message, "Warning")
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, true)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but the MV2 prefix is used incorrectly, so we return an error for that
//	 - e.g. MV2;MicroSeconds;<query>
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
			kClient := &keptnClientMock{
				customQueries: map[string]string{
					indicator: tc.mv2Prefix + "metricSelector=builtin:service.response.time:percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
				},
			}

			assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
				assert.EqualValues(t, indicator, actual.Metric)
				assert.EqualValues(t, 0, actual.Value)
				assert.EqualValues(t, false, actual.Success)
				assert.Contains(t, actual.Message, "MV2;")
				assert.Contains(t, actual.Message, "SLI definition format")
			}

			assertThatTestIsCorrect(t, handler, kClient, assertionsFunc, true)
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists but there are no SLIs defined OR
// * there is no 'dynatrace/sli.yaml' file
//   - currently this would lead to a fallback for default SLI definitions
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreDefinedButEmpty(t *testing.T) {
	// fallback: mind the default SLI definitions in the URL below
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29%2Ctag%28keptn_service%3Acarts%29%2Ctag%28keptn_deployment%3A%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_1_result_defaults.json")

	// no custom queries defined here
	// currently this could have 2 reasons: EITHER no sli.yaml file available OR no indicators defined in such a file)
	// TODO 2021-09-29: we should be able to differentiate between 'not there' and 'no SLIs defined' - the latter could be intentional
	kClient := &keptnClientMock{}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 12.439619479902443, actual.Value) // div by 1000 from dynatrace API result!
		assert.EqualValues(t, true, actual.Success)
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, false)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomSLIsAreUsedWhenSpecified(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_1_result.json")

	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 12.439619479902443, actual.Value) // div by 1000 from dynatrace API result!
		assert.EqualValues(t, true, actual.Success)
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, assertionsFunc, false)
}

func assertThatCustomSLITestIsCorrect(t *testing.T, handler http.Handler, kClient *keptnClientMock, assertionsFunc func(t *testing.T, actual *keptnv2.SLIResult), shouldFail bool) {
	// we use the special mock for the resource client
	// we do not want to query a dashboard, so we leave it empty (and have no dashboard stored)
	setupTestAndAssertNoError(t, handler, kClient, &resourceClientMock{t: t}, "")

	assertThatEventHasExpectedPayloadWithMatchingFunc(t, assertionsFunc, kClient.eventSink, shouldFail)
}

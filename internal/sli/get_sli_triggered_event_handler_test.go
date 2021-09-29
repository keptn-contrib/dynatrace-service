package sli

import (
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
	"testing"
)

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI has errors, so parsing the YAML file would not be possible
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreInvalidYAML(t *testing.T) {
	const indicator = "response_time_p95"

	ev := &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
	}

	// make sure we would not be able to query any metric due to a parsing error
	handler := test.NewFileBasedURLHandler(t)

	const errorMessage = "invalid YAML file - some parsing issue"
	kClient := &keptnClientMock{
		customQueriesError: fmt.Errorf(errorMessage),
	}

	eh, _, teardown := createGetSLIEventHandler(ev, handler, kClient)
	defer teardown()

	err := eh.retrieveMetrics()

	expectedResult := &keptnv2.SLIResult{
		Metric:  indicator,
		Value:   0,
		Success: false,
		Message: errorMessage,
	}

	assert.NoError(t, err)
	assertThatEventHasExpectedPayload(t, expectedResult, kClient.eventSink)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 0 results and a warning
//   - e.g. misspelled dimension key in merge transformation
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsNoResultsAndWarning(t *testing.T) {
	const indicator = "response_time_p95"

	ev := &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
	}

	// error here: merge(dt.entity.services)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		"/api/v2/metrics/query?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.services%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_0_result_warning_entity-selector.json")

	// error here as well: merge("dt.entity.services")
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:merge(\"dt.entity.services\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	eh, _, teardown := createGetSLIEventHandler(ev, handler, kClient)
	defer teardown()

	err := eh.retrieveMetrics()

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "no result values")
		assert.Contains(t, actual.Message, "Warning")
	}

	assert.NoError(t, err)
	assertThatEventHasExpectedPayloadWithMatchingFunc(t, assertionsFunc, kClient.eventSink)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, Dynatrace can process the query correctly (200), but returns 0 results and no warning
//	 - e.g. misspelled tag name
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryReturnsNoResults(t *testing.T) {
	const indicator = "response_time_p95"

	ev := &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
	}

	// error here: tag(keptn_project:stagin)
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		"/api/v2/metrics/query?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astagin%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_0_result_wrong-tag.json")

	// error here as well: tag(keptn_project:stagin)
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:stagin)",
		},
	}

	eh, _, teardown := createGetSLIEventHandler(ev, handler, kClient)
	defer teardown()

	err := eh.retrieveMetrics()

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "no result values")
		assert.NotContains(t, actual.Message, "Warning")
	}

	assert.NoError(t, err)
	assertThatEventHasExpectedPayloadWithMatchingFunc(t, assertionsFunc, kClient.eventSink)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * no (previous) dashboard is stored in Keptn
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomSLIsAreUsedWhenSpecified(t *testing.T) {

	const indicator = "response_time_p95"

	ev := &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
	}

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		"/api/v2/metrics/query?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000",
		"./testdata/response_time_p95_200_1_result.json")

	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	eh, _, teardown := createGetSLIEventHandler(ev, handler, kClient)
	defer teardown()

	err := eh.retrieveMetrics()

	expectedResult := &keptnv2.SLIResult{
		Metric:  indicator,
		Value:   12.439619479902443, // div by 1000 from dynatrace API result!
		Success: true,
	}

	assert.NoError(t, err)
	assertThatEventHasExpectedPayload(t, expectedResult, kClient.eventSink)
}

func assertThatEventHasExpectedPayload(t *testing.T, expectedResult *keptnv2.SLIResult, events []*cloudevents.Event) {
	data := assertThatEventsAreThere(t, events)

	assert.EqualValues(t, 1, len(data.GetSLI.IndicatorValues))
	assert.EqualValues(t, expectedResult, data.GetSLI.IndicatorValues[0])
}

func assertThatEventHasExpectedPayloadWithMatchingFunc(t *testing.T, assertionsFunc func(*testing.T, *keptnv2.SLIResult), events []*cloudevents.Event) {
	data := assertThatEventsAreThere(t, events)

	assert.EqualValues(t, 1, len(data.GetSLI.IndicatorValues))
	assertionsFunc(t, data.GetSLI.IndicatorValues[0])
}

func assertThatEventsAreThere(t *testing.T, events []*cloudevents.Event) *keptnv2.GetSLIFinishedEventData {
	assert.EqualValues(t, 2, len(events))

	assert.EqualValues(t, keptnv2.GetStartedEventType(keptnv2.GetSLITaskName), events[0].Type())
	assert.EqualValues(t, keptnv2.GetFinishedEventType(keptnv2.GetSLITaskName), events[1].Type())

	var data keptnv2.GetSLIFinishedEventData
	err := json.Unmarshal(events[1].Data(), &data)
	if err != nil {
		t.Fatalf("could not parse event payload correctly: %s", err)
	}

	assert.EqualValues(t, keptnv2.ResultPass, data.Result)
	assert.EqualValues(t, keptnv2.StatusSucceeded, data.Status)

	return &data
}

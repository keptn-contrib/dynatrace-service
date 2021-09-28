package sli

import (
	"encoding/json"
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
func TestCustomSLIsAreUsedWhenSpecified(t *testing.T) {

	const indicator = "response_time_p95"

	ev := &getSLIEventData{
		project:    "sockshop",
		stage:      "staging",
		service:    "carts",
		indicators: []string{indicator}, // we need this to check later on in the custom queries
	}

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/metrics/query?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1632835299000", "./testdata/response_time_p95_200_1_result.json")

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
	assertThatEventsHaveSuccessPayload(t, expectedResult, kClient.eventSink)
}

func assertThatEventsHaveSuccessPayload(t *testing.T, expectedResult *keptnv2.SLIResult, events []*cloudevents.Event) {
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

	assert.EqualValues(t, 1, len(data.GetSLI.IndicatorValues))
	assert.EqualValues(t, expectedResult, data.GetSLI.IndicatorValues[0])
}

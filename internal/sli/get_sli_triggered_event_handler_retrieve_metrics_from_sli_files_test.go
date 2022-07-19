package sli

import (
	"fmt"
	"net/http"
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

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

	eventSenderClient := &eventSenderClientMock{}

	// error here in the misspelled indicator:
	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		"response_time_p59": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	sliResultAssertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "response_time_p95")
	}

	assertThatCustomSLITestIsCorrect(t, handler, eventSenderClient, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultAssertionsFunc)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but Dynatrace cannot process the query correctly and returns a 400 error
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryIsNotValid(t *testing.T) {
	// error here: metric(s)Selector=
	handler := test.NewFileBasedURLHandler(t)

	eventSenderClient := &eventSenderClientMock{}

	// error here as well: metric(s)Selector=
	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		indicator: "metricsSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	sliResultAssertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "error parsing Metrics v2 query")
	}

	assertThatCustomSLITestIsCorrect(t, handler, eventSenderClient, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultAssertionsFunc)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI has errors, so parsing the YAML file would not be possible
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreInvalidYAML(t *testing.T) {
	// make sure we would not be able to query any metric due to a parsing error
	handler := test.NewFileBasedURLHandler(t)

	eventSenderClient := &eventSenderClientMock{}

	const errorMessage = "invalid YAML file - some parsing issue"
	rClient := newResourceClientMockWithGetSLIsError(t, fmt.Errorf(errorMessage))

	sliResultAssertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, errorMessage)
	}

	assertThatCustomSLITestIsCorrect(t, handler, eventSenderClient, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultAssertionsFunc)
}

func assertThatCustomSLITestIsCorrect(t *testing.T, handler http.Handler, eventSenderClient *eventSenderClientMock, rClient *resourceClientMock, getSLIFinishedEventAssertionsFunc func(t *testing.T, data *keptnv2.GetSLIFinishedEventData), sliResultAssertionsFunc func(t *testing.T, actual *keptnv2.SLIResult)) {
	// we use the special mock for the resource client
	// we do not want to query a dashboard, so we leave it empty
	runTestAndAssertNoError(t, testGetSLIEventDataWithDefaultStartAndEnd, handler, eventSenderClient, rClient, "")

	assertCorrectGetSLIEvents(t, eventSenderClient.eventSink, getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
}

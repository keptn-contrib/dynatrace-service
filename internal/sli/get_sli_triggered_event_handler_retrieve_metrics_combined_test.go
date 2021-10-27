package sli

import (
	"io/ioutil"
	"net/http"
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

type uploadWillFailResourceClientMock struct {
	t                *testing.T
	dashboardContent string
}

func (m *uploadWillFailResourceClientMock) GetSLOs(project string, stage string, service string) (*keptnapi.ServiceLevelObjectives, error) {
	m.t.Fatalf("GetSLOs() should not be needed in this mock!")

	return nil, nil
}

func (m *uploadWillFailResourceClientMock) UploadSLI(project string, stage string, service string, sli *dynatrace.SLI) error {
	m.t.Fatalf("UploadSLI() should not be needed in this mock!")

	return nil
}

func (m *uploadWillFailResourceClientMock) UploadSLOs(project string, stage string, service string, dashboardSLOs *keptnapi.ServiceLevelObjectives) error {
	m.t.Fatalf("UploadSLO() should not be needed in this mock!")

	return nil
}

func (m *uploadWillFailResourceClientMock) GetDashboard(project string, stage string, service string) (string, error) {
	return m.dashboardContent, nil
}

// Retrieving a dashboard by ID  SLI from a dashboard works, but Upload of dashboard, SLO or SLI file could fail
//
// prerequisites:
//   * we have a dashboard stored in Keptn
//   * we use a valid dashboard ID and it is returned by Dynatrace API
//   * The dashboard did not change and it has 'KQG.QueryBehavior=ParseOnChange' set to only reparse the dashboard if it
//     changed
//   * we will fallback to processing the stored SLIs and that works
func TestThatFallbackToSLIsFromDashboardIfDashboardDidNotChangeWorks(t *testing.T) {

	// no need for metrics definition, because we will be retrieving metrics from SLI files
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/12345678-1111-4444-8888-123456789012", "./testdata/combined_test/dashboard_custom_charting_single_sli_parse_only_on_change.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Asockshop%29%2Ctag%28keptn_stage%3Astaging%29&from=1632834999000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895.000000%29%3Anames&resolution=Inf&to=1632835299000",
		"./testdata/combined_test/response_time_p95_200_1_result.json")

	// we need custom queries, as we are not using the dashboard
	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95.000000):names&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
		},
	}

	// this is the dashboard that we want to compare against - it is slightly different that the one coming from
	// Dynatrace, because we would already have processed it beforehand. ALSO: NO NEW LINE AT END OF FILE!
	dashboardContent, err := ioutil.ReadFile("./testdata/combined_test/dashboard_stored_in_keptn.json")
	if err != nil {
		t.Fatalf("could not read dashboard file")
	}

	// if an upload of sli/slo/dashboard would be triggered then we fail, because we should not be doing that
	rClient := &uploadWillFailResourceClientMock{
		t:                t,
		dashboardContent: string(dashboardContent),
	}

	assertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 12.439619479902443, actual.Value) // div by 1000 from dynatrace API result!
		assert.EqualValues(t, true, actual.Success)
	}

	eventAssertionsFunc := func(data *keptnv2.GetSLIFinishedEventData) {
		assert.EqualValues(t, keptnv2.ResultPass, data.Result)
		assert.Empty(t, data.Message)
	}

	assertThatCombinedTestIsCorrect(t, handler, kClient, rClient, assertionsFunc, eventAssertionsFunc)
}

func assertThatCombinedTestIsCorrect(t *testing.T, handler http.Handler, kClient *keptnClientMock, rClient keptn.ResourceClientInterface, assertionsFunc func(t *testing.T, actual *keptnv2.SLIResult), eventAssertionsFunc func(data *keptnv2.GetSLIFinishedEventData)) {
	setupTestAndAssertNoError(t, handler, kClient, rClient, "12345678-1111-4444-8888-123456789012")

	assertThatEventHasExpectedPayloadWithMatchingFunc(t, assertionsFunc, kClient.eventSink, eventAssertionsFunc)
}

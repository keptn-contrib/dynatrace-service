package query

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/stretchr/testify/assert"
)

const testDynatraceAPIToken = "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"

// TestGetSLIValueSupportsEnvPlaceholders tests that environment variable placeholders are replaced correctly in SLI definitions.
func TestGetSLIValueSupportsEnvPlaceholders(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/metrics/query?entitySelector=type%28SERVICE%29%2Ctag%28%22env_tag%3Asome_tag%22%29&from=1571649084000&metricSelector=builtin%3Aservice.response.time&resolution=Inf&to=1571649085000", "./testdata/get_sli_value_env_placeholders_test/metrics_query_result.json")

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := &test.EventData{}
	timeframe := createTestTimeframe(t)

	indicator := "response_time_env"

	os.Setenv("MY_ENV_TAG", "some_tag")

	customQueries := make(map[string]string)
	customQueries[indicator] = "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"env_tag:$ENV.MY_ENV_TAG\")&metricSelector=builtin:service.response.time"

	ret := createCustomQueryProcessing(t, keptnEvent, httpClient, NewCustomQueries(customQueries), timeframe)
	sliResult := ret.GetSLIResultFromIndicator(context.TODO(), indicator)

	assert.True(t, sliResult.Success)
	assert.EqualValues(t, 0.29, sliResult.Value)

	os.Unsetenv("MY_ENV_TAG")
}

// TestGetSLIValueSupportsPlaceholders tests that placeholders are replaced correctly in SLI definitions.
func TestGetSLIValueSupportsPlaceholders(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/metrics/query?entitySelector=type%28SERVICE%29%2Ctag%28%22keptn_managed%22%29%2Ctag%28%22keptn_project%3Amyproject%22%29%2Ctag%28%22keptn_stage%3Amystage%22%29%2Ctag%28%22keptn_service%3Amyservice%22%29&from=1571649084000&metricSelector=builtin%3Aservice.response.time&resolution=Inf&to=1571649085000", "./testdata/get_sli_value_placeholders_test/metrics_query_result.json")
	handler.AddExact("/api/v2/metrics/query?entitySelector=type%28SERVICE%29%2Ctag%28%22keptn_deployment%3Amydeployment%22%29%2Ctag%28%22context%3Amycontext%22%29%2Ctag%28%22keptn_stage%3Amystage%22%29%2Ctag%28%22keptn_service%3Amyservice%22%29&from=1571649084000&metricSelector=builtin%3Aservice.response.time&resolution=Inf&to=1571649085000", "./testdata/get_sli_value_placeholders_test/metrics_query_result.json")
	handler.AddExact("/api/v2/problems?from=1571649084000&problemSelector=status%28open%29&to=1571649085000", "./testdata/get_sli_value_placeholders_test/problems_query_result.json")
	handler.AddExact("/api/v2/securityProblems?from=1571649084000&securityProblemSelector=status%28open%29&to=1571649085000", "./testdata/get_sli_value_placeholders_test/security_problems_query_result.json")
	handler.AddExact("/api/v2/slo/$LABELS.slo_id?from=1571649084000&timeFrame=GTF&to=1571649085000", "./testdata/get_sli_value_placeholders_test/slo_query_result.json")
	handler.AddExact("/api/v1/userSessionQueryLanguage/table?addDeepLinkFields=false&endTimestamp=1571649085000&explain=false&query=SELECT+osVersion%2C+AVG%28duration%29+FROM+usersession+WHERE+country+IN%28%27Austria%27%29+GROUP+BY+osVersion&startTimestamp=1571649084000", "./testdata/get_sli_value_placeholders_test/usql_query_results.json")

	httpClient, teardown := test.CreateHTTPClient(handler)
	defer teardown()

	keptnEvent := &test.EventData{
		Context:      "mycontext",
		Event:        "myevent",
		Project:      "myproject",
		Stage:        "mystage",
		Service:      "myservice",
		Deployment:   "mydeployment",
		TestStrategy: "mystrategy",
		Labels: map[string]string{
			"slo_id":         "524ca177-849b-3e8c-8175-42b93fbc33c5",
			"problem_status": "open",
			"country":        "Austria",
		},
	}

	timeframe := createTestTimeframe(t)

	testConfigs := []struct {
		indicator        string
		query            string
		expectedSLIValue float64
	}{
		{
			indicator:        "response_time",
			query:            "MV2;MicroSecond;entitySelector=type(SERVICE),tag(\"keptn_managed\"),tag(\"keptn_project:$PROJECT\"),tag(\"keptn_stage:$STAGE\"),tag(\"keptn_service:$SERVICE\")&metricSelector=builtin:service.response.time",
			expectedSLIValue: 0.29,
		},
		{
			indicator:        "response_time2",
			query:            "entitySelector=type(SERVICE),tag(\"keptn_deployment:$DEPLOYMENT\"),tag(\"context:$CONTEXT\"),tag(\"keptn_stage:$STAGE\"),tag(\"keptn_service:$SERVICE\")&metricSelector=builtin:service.response.time",
			expectedSLIValue: 290,
		},
		{
			indicator:        "problems",
			query:            "PV2;problemSelector=status($LABEL.problem_status)",
			expectedSLIValue: 1,
		},
		{
			indicator:        "security_problems",
			query:            "SECPV2;securityProblemSelector=status($LABEL.problem_status)",
			expectedSLIValue: 4,
		},
		{
			indicator:        "RT_faster_500ms",
			query:            "SLO;$LABELS.slo_id",
			expectedSLIValue: 96,
		},
		{
			indicator:        "User_session_time",
			query:            "USQL;COLUMN_CHART;iOS 12.1.4;SELECT osVersion, AVG(duration) FROM usersession WHERE country IN('$LABEL.country') GROUP BY osVersion",
			expectedSLIValue: 21478,
		},
	}

	for _, testConfig := range testConfigs {
		customQueries := make(map[string]string)
		customQueries[testConfig.indicator] = testConfig.query

		ret := createCustomQueryProcessing(t, keptnEvent, httpClient, NewCustomQueries(customQueries), timeframe)

		sliResult := ret.GetSLIResultFromIndicator(context.TODO(), testConfig.indicator)

		assert.True(t, sliResult.Success)
		assert.EqualValues(t, testConfig.expectedSLIValue, sliResult.Value)
	}
}

func createQueryProcessing(t *testing.T, keptnEvent adapter.EventContentAdapter, httpClient *http.Client, timeframe common.Timeframe) *Processing {
	return createCustomQueryProcessing(
		t,
		keptnEvent,
		httpClient,
		NewEmptyCustomQueries(),
		timeframe)
}

func createCustomQueryProcessing(t *testing.T, keptnEvent adapter.EventContentAdapter, httpClient *http.Client, queries *CustomQueries, timeframe common.Timeframe) *Processing {
	credentials, err := credentials.NewDynatraceCredentials("http://dynatrace", testDynatraceAPIToken)
	assert.NoError(t, err)

	return NewProcessing(
		dynatrace.NewClientWithHTTP(
			credentials,
			httpClient),
		keptnEvent,
		[]*keptnv2.SLIFilter{},
		queries,
		timeframe)
}

func createDefaultTestEventData() adapter.EventContentAdapter {
	return &test.EventData{
		Project: "sockshop",
		Stage:   "dev",
		Service: "carts",
	}
}

func createTestTimeframe(t *testing.T) common.Timeframe {
	timeframe, err := common.NewTimeframeParser("2019-10-21T09:11:24Z", "2019-10-21T09:11:25Z").Parse()
	assert.NoError(t, err)
	return *timeframe
}

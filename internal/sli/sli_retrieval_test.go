package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	_ "github.com/keptn/go-utils/pkg/lib"
)

const QUALITYGATE_DASHBOARD_ID = "12345678-1111-4444-8888-123456789012"
const QUALITYGATE_PROJECT = "qualitygate"
const QUALTIYGATE_SERVICE = "evalservice"
const QUALITYGATE_STAGE = "qualitystage"

/**
 * This function will create a new HTTP Server for handling Dynatrace REST Calls.
 * It returns the Retrieval and the teardown method
 * ATTENTION: When using this method you have to call the "teardown" method that is returned in the last parameter
 */
func createRetrieval(keptnEvent GetSLITriggeredAdapterInterface) (*Retrieval, func()) {

	handler := test.NewURLHandler()
	// we handle these if the URLs are a full match
	handler.AddExact("/api/config/v1/dashboards", "./testfiles/test_get_dashboards.json")
	handler.AddExact("/api/config/v1/dashboards/12345678-1111-4444-8888-123456789012", "./testfiles/test_get_dashboards_id.json")
	handler.AddExact("/api/v2/metrics/builtin:tech.generic.processCount", "./testfiles/test_get_metrics_processcount.json")
	handler.AddExact("/api/v2/metrics/builtin:service.response.time", "./testfiles/test_get_metrics_svcresponsetime.json")
	handler.AddExact("/api/v2/metrics/builtin:tech.generic.mem.workingSetSize", "./testfiles/test_get_metrics_workingsetsize.json")
	handler.AddExact("/api/v2/metrics/builtin:tech.generic.cpu.usage", "./testfiles/test_get_metrics_cpuusage.json")
	handler.AddExact("/api/v2/metrics/builtin:service.errors.server.rate", "./testfiles/test_get_metrics_errorrate.json")
	handler.AddExact("/api/v2/metrics/builtin:service.requestCount.total", "./testfiles/test_get_metrics_requestcount.json")
	handler.AddExact("/api/v2/metrics/builtin:host.cpu.usage", "./testfiles/test_get_metrics_hostcpuusage.json")
	handler.AddExact("/api/v2/metrics/builtin:host.mem.usage", "./testfiles/test_get_metrics_hostmemusage.json")
	handler.AddExact("/api/v2/metrics/builtin:host.disk.queueLength", "./testfiles/test_get_metrics_hostdiskqueue.json")
	handler.AddExact("/api/v2/metrics/builtin:service.nonDbChildCallCount", "./testfiles/test_get_metrics_nondbcallcount.json")
	handler.AddExact("/api/v2/metrics/jmeter.usermetrics.transaction.meantime", "./testfiles/test_get_metrics_jmeter_usermetrics_transaction_meantime.json")
	// we handle these if the URL "starts with"
	handler.AddStartsWith("/api/v2/metrics/query", "./testfiles/test_get_metrics_query.json")
	handler.AddStartsWith("/api/v2/slo", "./testfiles/test_get_slo_id.json")
	handler.AddStartsWith("/api/v2/problems", "./testfiles/test_get_problems.json")
	handler.AddStartsWith("/api/v2/securityProblems", "./testfiles/test_get_securityproblems.json")

	ret, _, teardown := createRetrievalWithHandler(keptnEvent, handler)

	return ret, teardown
}

func TestFindDynatraceDashboardSuccess(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	dashboardID, err := dh.findDynatraceDashboard(keptnEvent)

	if err != nil {
		t.Error(err)
	}

	if dashboardID != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("findDynatraceDashboard not finding quality gate dashboard")
	}
}

func TestFindDynatraceDashboardNoneExistingDashboard(t *testing.T) {
	keptnEvent := createKeptnEvent("BAD PROJECT", QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	dashboardID, err := dh.findDynatraceDashboard(keptnEvent)

	if err != nil {
		t.Error(err)
	}

	if dashboardID != "" {
		t.Errorf("findDynatraceDashboard found a dashboard that should not have been found: " + dashboardID)
	}
}

func TestLoadDynatraceDashboardWithQUERY(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.loadDynatraceDashboard(keptnEvent, common.DynatraceConfigDashboardQUERY)

	if dashboardJSON == nil {
		t.Errorf("Didnt query dashboard for quality gate project even though it shoudl exist: " + dashboard)
	}

	if dashboard != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("Didnt query the dashboard that matches the project/stage/service names: " + dashboard)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoadDynatraceDashboardWithID(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.loadDynatraceDashboard(keptnEvent, QUALITYGATE_DASHBOARD_ID)

	if dashboardJSON == nil {
		t.Errorf("Didnt query dashboard for quality gate project even though it should exist by ID")
	}

	if dashboard != QUALITYGATE_DASHBOARD_ID {
		t.Errorf("loadDynatraceDashboard should return the passed in dashboard id")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoadDynatraceDashboardWithEmptyDashboard(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	// this should load the dashboard
	dashboardJSON, dashboard, err := dh.loadDynatraceDashboard(keptnEvent, "")

	if dashboardJSON != nil {
		t.Errorf("No dashboard should be loaded if no dashboard is passed")
	}

	if dashboard != "" {
		t.Errorf("dashboard should be empty as by default we dont load a dashboard")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestGetEntitySelectorFromEntityFilter(t *testing.T) {
	expected := ",entityId(\"SERVICE-086C46F600BA1DC6\"),tag(\"keptn_deployment:primary\")"

	var filtersPerEntityType = map[string]map[string][]string{
		"SERVICE": {
			"SPECIFIC_ENTITIES": {"SERVICE-086C46F600BA1DC6"},
			"AUTO_TAGS":         {"keptn_deployment:primary"},
		},
	}
	actual := getEntitySelectorFromEntityFilter(filtersPerEntityType, "SERVICE")

	assert.Equal(t, expected, actual)
}

func TestExecuteGetDynatraceSLO(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	sloID := "524ca177-849b-3e8c-8175-42b93fbc33c5"
	sloResult, err := dynatrace.NewSLOClient(dh.dtClient).Get(sloID, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if sloResult == nil {
		t.Errorf("No SLO Result returned for " + sloID)
	}

	if sloResult.EvaluatedPercentage != 95.66405076939219 {
		t.Error("Not returning expected value for SLO")
	}
}

func TestGetSLIValueWithSLOPrefix(t *testing.T) {

	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	customQueries := make(map[string]string)
	customQueries["RT_faster_500ms"] = "SLO;524ca177-849b-3e8c-8175-42b93fbc33c5"

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()

	_, err := dh.GetSLIValue("RT_faster_500ms", startTime, endTime, keptn.NewCustomQueries(customQueries))

	if err != nil {
		t.Error(err)
	}
}

func TestExecuteGetDynatraceProblems(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(open)"
	problemResult, err := dynatrace.NewProblemsV2Client(dh.dtClient).GetByQuery(problemQuery, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if problemResult == nil {
		t.Fatal("No Problem Result returned for " + problemQuery)
	}

	if problemResult.TotalCount != 1 {
		t.Error("Not returning expected value for Problem Query")
	}
}

func TestExecuteGetDynatraceSecurityProblems(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(OPEN)"

	// TODO 2021-09-02: fix dependency on sli/Retrieval below!
	problemResult, err := dynatrace.NewSecurityProblemsClient(dh.dtClient).GetByQuery(problemQuery, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if problemResult == nil {
		t.Fatal("No Problem Result returned for " + problemQuery)
	}

	if problemResult.TotalCount != 0 {
		t.Error("Not returning expected value for Problem Query")
	}
}

func TestGetSLIValueWithPV2Prefix(t *testing.T) {

	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	customQueries := make(map[string]string)
	customQueries["problems"] = "PV2;problemEntity=status(open)"

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()

	_, err := dh.GetSLIValue("problems", startTime, endTime, keptn.NewCustomQueries(customQueries))

	if err != nil {
		t.Error(err)
	}
}

func TestGetSLIValueWithSECPV2Prefix(t *testing.T) {

	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	customQueries := make(map[string]string)
	customQueries["security_problems"] = "SECPV2;problemEntity=status(open)"

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()

	_, err := dh.GetSLIValue("security_problems", startTime, endTime, keptn.NewCustomQueries(customQueries))

	if err != nil {
		t.Error(err)
	}
}

func TestScaleData(t *testing.T) {
	if scaleData("", "MicroSecond", 1000000.0) != 1000.0 {
		t.Errorf("scaleData incorrectly scales MicroSecond")
	}
	if scaleData("", "Byte", 1024.0) != 1.0 {
		t.Errorf("scaleData incorrectly scales Bytes")
	}
	if scaleData("builtin:service.response.time", "", 1000000.0) != 1000.0 {
		t.Errorf("scaleData incorrectly scales builtin:service.response.time")
	}
}

func TestIsValidUUID(t *testing.T) {
	testConfigs := []struct {
		uuid string
		want bool
	}{
		// reproduce issue with "|"
		{
			"311f4aa7-5257-41d7-|bd1-70420500e1c8",
			false,
		},
		// valid UUID v4, variant 1
		{
			"311f4aa7-5257-41d7-abd1-70420500e1c8",
			true,
		},
		// NIL UUID is not valid
		{
			"00000000-0000-0000-0000-000000000000",
			false,
		},
	}
	for _, config := range testConfigs {
		got := isValidUUID(config.uuid)
		if got != config.want {
			t.Errorf("uuid: %s, result should have been: %v, but got: %v", config.uuid, config.want, got)
		}
	}
}

func TestParseMarkdownConfigurationParams(t *testing.T) {
	testConfigs := []struct {
		input              string
		expectedScore      *keptnapi.SLOScore
		expectedComparison *keptnapi.SLOComparison
	}{
		// single result
		{
			"KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			createSLOScore("90%", "70%"),
			createSLOComparison("single_result", "pass", 1, "avg"),
		},
		// several results, p50
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p50",
			createSLOScore("50%", "40%"),
			createSLOComparison("several_results", "pass", 3, "p50"),
		},
		// several results, p90
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p90",
			createSLOScore("50%", "40%"),
			createSLOComparison("several_results", "pass", 3, "p90"),
		},
		// several results, p95
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLOScore("50%", "40%"),
			createSLOComparison("several_results", "pass", 3, "p95"),
		},
		// several results, p95, all
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=all;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLOScore("50%", "40%"),
			createSLOComparison("several_results", "all", 3, "p95"),
		},
		// several results, p95, pass_or_warn
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass_or_warn;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLOScore("50%", "40%"),
			createSLOComparison("several_results", "pass_or_warn", 3, "p95"),
		},

		// several results, p95, fallback to pass if compare function is unknown
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=warn;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLOScore("50%", "40%"),
			createSLOComparison("several_results", "pass", 3, "p95"),
		},
		// several results, fallback if function is unknown e.g. p97
		{
			"KQG.Total.Pass=51%;KQG.Total.Warning=41%;KQG.Compare.WithScore=pass;KQG.Compare.Results=4;KQG.Compare.Function=p97",
			createSLOScore("51%", "41%"),
			createSLOComparison("several_results", "pass", 4, "avg"),
		},
	}
	for _, config := range testConfigs {
		actualScore, actualComparison := parseMarkdownConfiguration(config.input)

		assert.EqualValues(t, config.expectedScore, actualScore)
		assert.EqualValues(t, config.expectedComparison, actualComparison)
	}
}

func createSLOScore(pass string, warning string) *keptnapi.SLOScore {
	return &keptnapi.SLOScore{
		Pass:    pass,
		Warning: warning,
	}
}
func createSLOComparison(compareWith string, include string, numberOfResults int, aggregateFunc string) *keptnapi.SLOComparison {
	return &keptnapi.SLOComparison{
		CompareWith:               compareWith,
		IncludeResultWithScore:    include,
		NumberOfComparisonResults: numberOfResults,
		AggregateFunction:         aggregateFunc,
	}
}

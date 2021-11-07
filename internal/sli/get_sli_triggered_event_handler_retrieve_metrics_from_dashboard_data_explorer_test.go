package sli

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

var testDataExplorerGetSLIEventData = createTestGetSLIEventDataWithStartAndEnd("2021-01-01T00:00:00.000Z", "2021-01-02T00:00:00.000Z")

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByName tests space aggregation average and filterby name.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByName(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_filterby_name/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_filterby_name.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.errors.total.count", testDataFolder+"metrics_get_by_id_builtin_service_errors_total_count.json")
	handler.AddExact(
		dynatrace.MetricsQueryPath+"?entitySelector=type%28SERVICE%29%2CentityName%28%22%2Fservices%2FConfigurationService%2F+on+haproxy%3A80+%28opaque%29%22%29&from=1609459200000&metricSelector=builtin%3Aservice.errors.total.count%3Amerge%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1609545600000",
		testDataFolder+"metrics_get_by_query_builtin_service_errors_total_count.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("any_errors", 5324),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "any_errors", "metricSelector=builtin:service.errors.total.count:merge(\"dt.entity.service\"):avg:names&entitySelector=type(SERVICE),entityName(\"/services/ConfigurationService/ on haproxy:80 (opaque)\")")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testDataExplorerGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

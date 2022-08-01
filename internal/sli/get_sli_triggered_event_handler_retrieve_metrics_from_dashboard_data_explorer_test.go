package sli

import (
	"testing"

	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByName tests space aggregation average and filterby name.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByName(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_filterby_name/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29%2CentityName%28%22%2Fservices%2FConfigurationService%2F+on+haproxy%3A80+%28opaque%29%22%29&from=1609459200000&metricSelector=builtin%3Aservice.errors.total.count%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_filterby_name.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.errors.total.count", testDataFolder+"metrics_get_by_id_builtin_service_errors_total_count.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_get_by_query_builtin_service_errors_total_count.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("any_errors", 5324, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_FilterByDimension tests filtering by dimension and splitting by dimension.
// TODO: 2021-11-11: Investigate and fix this test
func TestRetrieveMetricsFromDashboardDataExplorerTile_FilterByDimension(t *testing.T) {
	t.Skip("Skipping test, as DIMENSION filter type needs to be investigated")

	const testDataFolder = "./testdata/dashboards/data_explorer/filterby_dimension/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_filterby_dimension.json")
	handler.AddExact(dynatrace.MetricsPath+"/jmeter.usermetrics.transaction.meantime", testDataFolder+"metrics_get_by_id_jmeter_usermetrics_transaction_meantime.json")
	handler.AddExact(dynatrace.MetricsQueryPath+"?entitySelector=entityId%28SERVICE-FFD81F003E39B468%29&from=1571649084000&metricSelector=jmeter.usermetrics.transaction.meantime%3Aavg%3Anames&resolution=Inf&to=1571649085000",
		testDataFolder+"metrics_get_by_query_jmeter.usermetrics_transaction_meantime.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		// include two function because two results are expected, but the values are not checked
		func(t *testing.T, actual sliResult) {},
		func(t *testing.T, actual sliResult) {},
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIButNoQuery tests a data explorer tile with an SLI name defined, i.e. in the title, but no query.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIButNoQuery(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/sli_name_no_query/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_sli_name_no_query.json")

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("new"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIAndTwoQueries tests a data explorer tile with an SLI name defined and two series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIAndTwoQueries(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/sli_name_two_queries/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_sli_name_two_queries.json")

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("two"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgNoFilterBy tests average space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_avg", 29.192929640271974, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgCountNoFilterBy tests count space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgCountNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_count_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Acount%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_count_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_count.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_count", 1060428.829, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgMaxNoFilterBy tests max space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgMaxNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_max_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Amax%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_max_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_max.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_max", 45156.016, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgMedianNoFilterBy tests median space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgMedianNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_median_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Amedian%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_median_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_median.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_median", 1.4999996049587276, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgMinNoFilterBy tests min space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgMinNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_min_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Amin%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_min_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_min.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_min", 0, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgP10NoFilterBy tests percentile(10) space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgP10NoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_p10_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2810%29%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_p10_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_p10.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_p10", 1.0000048892760918, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgP75NoFilterBy tests percentile(75) space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgP75NoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_p75_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2875%29%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_p75_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_p75.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_p75", 3.2541557923119475, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgp90NoFilterBy tests percentile(90) space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgp90NoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_p90_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2890%29%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_p90_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_p90.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_p90", 35.00000424055808, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgSumNoFilterBy tests sum space aggregation and no filterby.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgSumNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_sum_no_filterby/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Asum%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_sum_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_sum.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_sum", 30957024193.513, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_NoSpaceAgNoFilterBy tests no space aggregation and no filterby.
// This is will result in a SLIResult with failure, as a space aggregation must be supplied.
func TestRetrieveMetricsFromDashboardDataExplorerTile_NoSpaceAgNoFilterBy(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/no_spaceag_no_filterby/"

	handler := test.NewFileBasedURLHandler(t)

	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_no_spaceag_no_filterby.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("rt"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterById tests average space aggregation and filterby entity id.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterById(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_filterby_id/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=entityId%28SERVICE-B67B3EC4C95E0FA7%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_filterby_id.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_jid", 136.52852484946527, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByTag tests average space aggregation and filterby tag.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByTag(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_filterby_tag/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29%2Ctag%28%22keptnmanager%22%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_filterby_tag.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_keptn_manager", 18.533351299277793, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByEntityAttribute tests average space aggregation and filterby entity attribute.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByEntityAttribute(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_filterby_entityattribute/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29%2CdatabaseName%28%22EasyTravelWeatherCache%22%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_filterby_entityattribute.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_svc_etw_db", 1.0706877628404, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByDimension tests average space aggregation and filterby dimension.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgFilterByDimension(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_filterby_dimension/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=calc%3Aservice.dbcalls%3Afilter%28EQ%28%22Statement%22%2C%22Reads+in+JourneyCollection%22%29%29%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_filterby_dimension.json")
	handler.AddExact(dynatrace.MetricsPath+"/calc:service.dbcalls", testDataFolder+"metrics_calc_service_dbcalls.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_calc_service_dbcalls_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("svc_db_calls", 0.005373592355230029, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgTwoFilters tests average space aggregation and two filters.
// This is will result in a SLIResult with failure, as only a single filter is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAgAvgTwoFilters(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/spaceag_avg_two_filters/"

	handler := test.NewFileBasedURLHandler(t)

	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_spaceag_avg_two_filters.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("rt_jt"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_NoFilter_NoManagementZone tests applying no filter and no management zone.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_NoFilter_NoManagementZone(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/management_zones/no_filter_no_managementzone/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt_no_filter_no_mz", 29.192929640271974, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_ServiceTag_Filter_NoManagementZone tests applying service tag filter and no management zone.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_ServiceTag_Filter_NoManagementZone(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/management_zones/servicetag_filter_no_managementzone/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29%2Ctag%28%22service_tag%22%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt_servicetag_filter_no_mz", 288.95723558253565, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_NoFilter_WithCustomManagementZone tests applying no filter and custom management zone.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_NoFilter_WithCustomManagementZone(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/management_zones/no_filter_with_custommanagementzone/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29%2CmzId%282311420533206603714%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt_no_filter_custom_mz", 7.045031103506126, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_ServiceTag_Filter_WithCustomManagementZone tests applying service tag filter and custom management zone.
// This is will result in a SLIResult with success, as this is supported.
func TestRetrieveMetricsFromDashboardDataExplorerTile_ServiceTag_Filter_WithCustomManagementZone(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/management_zones/servicetag_filter_with_custommanagementzone/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=type%28SERVICE%29%2Ctag%28%22service_tag%22%29%2CmzId%282311420533206603714%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt_servicetag_filter_custom_mz", 8.283891270010905, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZoneWithNoEntityType tests that an error is produced for data explorer tiles with a management zone and no obvious entity type.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZoneWithNoEntityType(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/no_entity_type/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:security.securityProblem.open.managementZone", testDataFolder+"metrics_builtin_security_securityProblem_open_managementZone.json")

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("vulnerabilities_high", "has no entity type"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_CustomSLO tests propagation of a customized SLO.
// This is will result in a SLIResult with success, as this is supported.
// Here also the SLO is checked, including the display name, weight and key SLI.
func TestRetrieveMetricsFromDashboardDataExplorerTile_CustomSLO(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/custom_slo/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt", 29.192929640271974, expectedMetricsRequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "srt",
			DisplayName: "Service response time",
			Pass:        []*keptnapi.SLOCriteria{{Criteria: []string{"<30"}}},
			Weight:      4,
			KeySLI:      true,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_ExcludedTile tests that an excluded tile is skipped.
func TestRetrieveMetricsFromDashboardDataExplorerTile_ExcludedTile(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/excluded_tile/"

	const expectedMetricsRequest = dynatrace.MetricsQueryPath + "?entitySelector=entityId%28SERVICE-B67B3EC4C95E0FA7%29&from=1609459200000&metricSelector=builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1609545600000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_excluded_tile.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_jid", 136.52852484946527, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

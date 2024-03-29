package sli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIButNoQuery tests a data explorer tile with an SLI name defined, i.e. in the title, but no query.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIButNoQuery(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/sli_name_no_query/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("new"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIAndTwoQueries tests a data explorer tile with an SLI name defined and two series.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_WithSLIAndTwoQueries(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/sli_name_two_queries/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("two"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_SingleValueVisualizationSingleResult tests that Data Explorer tiles with single value visualization that generates a single result works as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_SingleValueVisualizationSingleResult(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"
	testVariantDataFolder := filepath.Join(testDataFolder, "single_value_visualization_single_result")

	const metricSelector = "(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	requestBuilder := newMetricsV2QueryRequestBuilder(metricSelector)

	handler := createHandlerForWithDashboardForMetricExpressionsTest(t, testDataFolder, singleValueVisualConfigType, &[]string{resolutionIsNullKeyValuePair + metricSelector, metricSelector})
	metricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, requestBuilder)

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("srt", 54896.50447404383, metricsRequest))
}

// TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_GraphChartVisualizationSingleResult tests that Data Explorer tiles with graph chart visualization that generate a single result work as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_GraphChartVisualizationSingleResult(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"
	testVariantDataFolder := filepath.Join(testDataFolder, "graph_chart_visualization_single_result")

	const metricSelector = "(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	requestBuilder := newMetricsV2QueryRequestBuilder(metricSelector)

	handler := createHandlerForWithDashboardForMetricExpressionsTest(t, testDataFolder, graphChartVisualConfigType, &[]string{resolutionIsNullKeyValuePair + metricSelector})
	metricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, requestBuilder)

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("srt", 54896.50447404383, metricsRequest))
}

// TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_GraphChartVisualizationMultipleResults tests that Data Explorer tiles with graph chart visualization that generate multiple results work as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_GraphChartVisualizationMultipleResults(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"
	testVariantDataFolder := filepath.Join(testDataFolder, "graph_chart_visualization_multiple_results")

	const metricSelector = "(builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	requestBuilder := newMetricsV2QueryRequestBuilder(metricSelector)

	handler := createHandlerForWithDashboardForMetricExpressionsTest(t, testDataFolder, graphChartVisualConfigType, &[]string{resolutionIsNullKeyValuePair + metricSelector})
	metricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, requestBuilder)

	multipleSuccessfulSLIResultAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt_narf_dynatrace-saas:1_246_9999_90000101-000000", 117432942.66666667, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_narf_evaluation", 24685775.14285714, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_volatile_span", 9603724.393200295, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_service_port:_8080", 5012392, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_requests_executed_in_background_threads_of_lambda_1-ig-1", 4899999.862939175, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_service_port_a:_443", 4899999.862939175, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_service_port:_8810", 2675495.4583333335, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_service_port_b:_443", 2662613, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt__services_authenticationservice_authenticationservicehttpsoap12endpoint_on_dynatrace-dev-bb:8091_(opaque)", 2038121, metricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_service_port_c:_443", 1599031.3548387096, metricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, multipleSuccessfulSLIResultAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_SingleValueVisualizationMultipleResults tests that Data Explorer tiles with single value visualization that generate multiple results don't work.
func TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_SingleValueVisualizationMultipleResults(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"
	testVariantDataFolder := filepath.Join(testDataFolder, "single_value_visualization_multiple_results")

	const metricSelector = "(builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	requestBuilder := newMetricsV2QueryRequestBuilder(metricSelector)

	handler := createHandlerForWithDashboardForMetricExpressionsTest(t, testDataFolder, singleValueVisualConfigType, &[]string{resolutionIsNullKeyValuePair + metricSelector, metricSelector})
	addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, requestBuilder)

	multipleSuccessfulSLIResultAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("srt", requestBuilder.build(), "metric series but only one is supported"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, multipleSuccessfulSLIResultAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_OtherResolution tests that Data Explorer tiles with an explicit resolution work as expected via fold.
func TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_OtherResolution(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"
	testVariantDataFolder := filepath.Join(testDataFolder, "other_resolution")

	const metricSelector = "(builtin:service.response.time:splitBy():sort(value(auto,descending)):limit(100)):limit(100):names"
	requestBuilder := newMetricsV2QueryRequestBuilder(metricSelector).copyWithResolution("30m")

	handler := createHandlerForWithDashboardForMetricExpressionsTest(t, testDataFolder, graphChartVisualConfigType, &[]string{"resolution=30m&" + metricSelector})
	metricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithFold(handler, testVariantDataFolder, requestBuilder)

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("srt", 54896.50522397428, metricsRequest))
}

// TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_FoldValueDoesntWork tests that folding for Data Explorer tiles doesn't work if the default aggregation type is value.
func TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_FoldValueDoesntWork(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"
	testVariantDataFolder := filepath.Join(testDataFolder, "fold_value")

	const metricSelector = "(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	requestBuilder := newMetricsV2QueryRequestBuilder(metricSelector).copyWithResolution("30m")

	handler := createHandlerForWithDashboardForMetricExpressionsTest(t, testDataFolder, graphChartVisualConfigType, &[]string{"resolution=30m&" + metricSelector})
	addRequestsToHandlerForSuccessfulMetricsQueryWithFold(handler, testVariantDataFolder, requestBuilder)

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc("srt", requestBuilder.build(), "unable to apply ':fold'"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_Errors tests that invalid metric expressions configurations generate errors as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTileMetricExpressions_Errors(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"

	const singleResultMetricExpression = "(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"

	tests := []struct {
		name                              string
		visualConfigType                  string
		metricExpressions                 *[]string
		getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData)
		sliResultsAssertionsFuncs         []func(t *testing.T, actual sliResult)
	}{
		{
			name:                              "nil metric expressions produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 nil,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "tile has no metric expressions"),
		},
		{
			name:              "zero metric expressions produces error",
			visualConfigType:  graphChartVisualConfigType,
			metricExpressions: &[]string{},

			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "tile has no metric expressions"),
		},
		{
			name:                              "missing resolution key value pair produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{singleResultMetricExpression},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "metric expression does not contain two components"),
		},
		{
			name:                              "wrong order in metric expression produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{singleResultMetricExpression + resolutionIsNullKeyValuePair},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "unexpected prefix in key value pair"),
		},
		{
			name:                              "empty value in resolution key value pair produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{"resolution=&" + singleResultMetricExpression},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "resolution must not be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createHandlerForWithDashboardForMetricExpressionsTest(t, testDataFolder, tt.visualConfigType, tt.metricExpressions)
			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, tt.getSLIFinishedEventAssertionsFunc, tt.sliResultsAssertionsFuncs...)
		})
	}
}

func createHandlerForWithDashboardForMetricExpressionsTest(t *testing.T, testDataFolder string, visualConfigType string, metricExpressions *[]string) *test.CombinedURLHandler {
	return createHandlerWithTemplatedDashboard(t,
		filepath.Join(testDataFolder, dashboardTemplateFilename),
		struct {
			VisualConfigType        string
			MetricExpressionsString string
		}{
			VisualConfigType:        visualConfigType,
			MetricExpressionsString: convertToJSONStringOrEmptyIfNil(t, metricExpressions),
		},
	)
}

func createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings(expectedQuery string, expectedMessageSubStrings ...string) []func(t *testing.T, actual sliResult) {
	return []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("srt", expectedQuery, expectedMessageSubStrings...)}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZonesWork tests applying management zones to the dashboard and tile work as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZonesWork(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/management_zones_work/"

	dashboardFilterWithManagementZone := dynatrace.DashboardFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "2311420533206603714",
			Name: "ap_mz_1",
		},
	}

	dashboardFilterWithAllManagementZones := dynatrace.DashboardFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "all",
			Name: "All",
		},
	}

	emptyTileFilter := dynatrace.TileFilter{}

	tileFilterWithManagementZone := dynatrace.TileFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "-6219736993013608218",
			Name: "ap_mz_2",
		},
	}

	tileFilterWithAllManagementZones := dynatrace.TileFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "all",
			Name: "All",
		},
	}

	requestBuilderWithNoManagementZone := newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():sort(value(auto,descending)):limit(10)):limit(100):names")
	requestBuilderWithManagementZone1 := requestBuilderWithNoManagementZone.copyWithMZSelector("mzName(\"ap_mz_1\")")
	requestBuilderWithManagementZone2 := requestBuilderWithNoManagementZone.copyWithMZSelector("mzName(\"ap_mz_2\")")

	tests := []struct {
		name             string
		dashboardFilter  *dynatrace.DashboardFilter
		tileFilter       dynatrace.TileFilter
		requestBuilder   *metricsV2QueryRequestBuilder
		expectedSLIValue float64
	}{
		{
			name:             "no_dashboard_filter_and_empty_tile_filter",
			dashboardFilter:  nil,
			tileFilter:       emptyTileFilter,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.50455400265,
		},
		{
			name:             "dashboard_filter_with_mz_and_empty_tile_filter",
			dashboardFilter:  &dashboardFilterWithManagementZone,
			tileFilter:       emptyTileFilter,
			requestBuilder:   requestBuilderWithManagementZone1,
			expectedSLIValue: 115445.40697872869,
		},
		{
			name:             "dashboard_filter_with_all_mz_and_empty_tile_filter",
			dashboardFilter:  &dashboardFilterWithAllManagementZones,
			tileFilter:       emptyTileFilter,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.50455400265,
		},
		{
			name:             "no_dashboard_filter_and_tile_filter_with_mz",
			dashboardFilter:  nil,
			tileFilter:       tileFilterWithManagementZone,
			requestBuilder:   requestBuilderWithManagementZone2,
			expectedSLIValue: 1519500.493859082,
		},
		{
			name:             "dashboard_filter_with_mz_and_tile_filter_with_mz",
			dashboardFilter:  &dashboardFilterWithManagementZone,
			tileFilter:       tileFilterWithManagementZone,
			requestBuilder:   requestBuilderWithManagementZone2,
			expectedSLIValue: 1519500.493859082,
		},
		{
			name:             "dashboard_filter_with_all_mz_and_tile_filter_with_mz",
			dashboardFilter:  &dashboardFilterWithAllManagementZones,
			tileFilter:       tileFilterWithManagementZone,
			requestBuilder:   requestBuilderWithManagementZone2,
			expectedSLIValue: 1519500.493859082,
		},
		{
			name:             "no_dashboard_filter_and_tile_filter_with_all_mz",
			dashboardFilter:  nil,
			tileFilter:       tileFilterWithAllManagementZones,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.50455400265,
		},
		{
			name:             "dashboard_filter_with_mz_and_tile_filter_with_all_mz",
			dashboardFilter:  &dashboardFilterWithManagementZone,
			tileFilter:       tileFilterWithAllManagementZones,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.50455400265,
		},
		{
			name:             "dashboard_filter_with_all_mz_and_tile_filter_with_all_mz",
			dashboardFilter:  &dashboardFilterWithAllManagementZones,
			tileFilter:       tileFilterWithAllManagementZones,
			requestBuilder:   requestBuilderWithNoManagementZone,
			expectedSLIValue: 54896.50455400265,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				filepath.Join(testDataFolder, dashboardTemplateFilename),
				struct {
					DashboardFilterString string
					TileFilterString      string
				}{
					DashboardFilterString: convertToJSONStringOrEmptyIfNil(t, tt.dashboardFilter),
					TileFilterString:      convertToJSONString(t, tt.tileFilter),
				},
			)

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)
			metricsQueryRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, tt.requestBuilder)
			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("srt", tt.expectedSLIValue, metricsQueryRequest))
		})
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZoneWithNoEntityType tests that data explorer tiles with a management zone and no obvious entity type work.
func TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZoneWithNoEntityType(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/no_entity_type/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	expectedMetricsRequest := newMetricsV2QueryRequestBuilder("(builtin:security.securityProblem.open.managementZone:filter(and(or(eq(\"Risk Level\",HIGH)))):splitBy(\"Risk Level\"):sum:sort(value(sum,descending)):fold(avg)):limit(100):names:fold(auto)").copyWithMZSelector("mzName(\"mz-1\")").build()
	handler.AddExactFile(expectedMetricsRequest, filepath.Join(testDataFolder, "metrics_get_by_query1.json"))

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("vulnerabilities_high", 536.4052024482108, expectedMetricsRequest))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_CustomSLO tests propagation of a customized SLO.
// This is will result in a SLIResult with success, as this is supported.
// Here also the SLO is checked, including the display name, weight and key SLI.
func TestRetrieveMetricsFromDashboardDataExplorerTile_CustomSLO(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/custom_slo/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt", 54896.50455400265, expectedMetricsRequest),
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

	handler := createHandlerWithDashboard(t, testDataFolder)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("(builtin:service.response.time:filter(and(or(in(\"dt.entity.service\",entitySelector(\"type(service),entityId(~\"SERVICE-C6876D601CA5DDFD~\")\"))))):splitBy(\"dt.entity.service\"):avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_jid", 57974.262650996854, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdsWork tests that setting pass and warning criteria via thresholds on the tile works as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdsWork(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/tile_thresholds_success/"

	requestBuilder := newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names")

	tests := []struct {
		name        string
		tileName    string
		thresholds  dynatrace.VisualizationThreshold
		expectedSLO *keptnapi.SLO
	}{
		{
			name:        "Valid pass-warn-fail increasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createPassThresholdRule(0), createWarnThresholdRule(68000), createFailThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createBandSLOCriteria(0, 68000), createBandSLOCriteria(0, 69000)),
		},
		{
			name:        "Valid fail-warn-pass increasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createFailThresholdRule(0), createWarnThresholdRule(68000), createPassThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createLowerBoundSLOCriteria(69000), createLowerBoundSLOCriteria(68000)),
		},
		{
			name:        "Valid pass-none-fail increasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createPassThresholdRule(0), createWarnThresholdRuleWithPointer(nil), createFailThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createBandSLOCriteria(0, 69000), nil),
		},
		{
			name:        "Valid fail-none-pass increasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createFailThresholdRule(0), createWarnThresholdRuleWithPointer(nil), createPassThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createLowerBoundSLOCriteria(69000), nil),
		},

		{
			name:        "Valid pass-warn-fail decreasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createPassThresholdRule(69000), createWarnThresholdRule(68000), createFailThresholdRule(0)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createLowerBoundSLOCriteria(69000), createLowerBoundSLOCriteria(68000)),
		},
		{
			name:        "Valid fail-warn-pass decreasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createFailThresholdRule(69000), createWarnThresholdRule(68000), createPassThresholdRule(0)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createBandSLOCriteria(0, 68000), createBandSLOCriteria(0, 69000)),
		},
		{
			name:        "Valid pass-none-fail decreasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createPassThresholdRule(69000), createWarnThresholdRuleWithPointer(nil), createFailThresholdRule(0)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createLowerBoundSLOCriteria(69000), nil),
		},
		{
			name:        "Valid fail-none-pass decreasing thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createFailThresholdRule(69000), createWarnThresholdRuleWithPointer(nil), createPassThresholdRule(0)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createBandSLOCriteria(0, 69000), nil),
		},

		{
			name:       "Pass or warning defined in title take precedence over valid thresholds",
			tileName:   "Service Response Time; sli=srt; pass=<70000; warning=<71000",
			thresholds: createVisibleThresholds(createPassThresholdRule(0), createWarnThresholdRule(68000), createFailThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(
				[]*keptnapi.SLOCriteria{{Criteria: []string{"<70000"}}},
				[]*keptnapi.SLOCriteria{{Criteria: []string{"<71000"}}}),
		},
		{
			name:        "Visible thresholds with no values are ignored",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createPassThresholdRuleWithPointer(nil), createWarnThresholdRuleWithPointer(nil), createFailThresholdRuleWithPointer(nil)),
			expectedSLO: createExpectedServiceResponseTimeSLO(nil, nil),
		},
		{
			name:        "Not visible thresholds with valid values are ignored",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createNotVisibleThresholds(createPassThresholdRule(0), createWarnThresholdRule(68000), createFailThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(nil, nil),
		},
		{
			name:        "Not visible thresholds with invalid values are ignored",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createNotVisibleThresholds(createPassThresholdRule(0), createWarnThresholdRule(68000), createFailThresholdRule(68000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(nil, nil),
		},
	}

	for _, thresholdTest := range tests {
		t.Run(thresholdTest.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				filepath.Join(testDataFolder, dashboardTemplateFilename),
				struct {
					TileName         string
					ThresholdsString string
				}{
					TileName:         thresholdTest.tileName,
					ThresholdsString: convertToJSONString(t, thresholdTest.thresholds),
				})

			metricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testDataFolder, requestBuilder)

			successfulSLIResultAssertionsFunc := createSuccessfulSLIResultAssertionsFunc("srt", 54896.50455400265, metricsRequest)

			uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
				if assert.Equal(t, 1, len(actual.Objectives)) {
					assert.EqualValues(t, thresholdTest.expectedSLO, actual.Objectives[0])
				}
			}

			runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, successfulSLIResultAssertionsFunc)
		})
	}
}

func createPassThresholdRule(value float64) dynatrace.VisualizationThresholdRule {
	return createPassThresholdRuleWithPointer(&value)
}

func createPassThresholdRuleWithPointer(value *float64) dynatrace.VisualizationThresholdRule {
	return dynatrace.VisualizationThresholdRule{Value: value, Color: "#7dc540"}
}

func createWarnThresholdRule(value float64) dynatrace.VisualizationThresholdRule {
	return createWarnThresholdRuleWithPointer(&value)
}

func createWarnThresholdRuleWithPointer(value *float64) dynatrace.VisualizationThresholdRule {
	return dynatrace.VisualizationThresholdRule{Value: value, Color: "#f5d30f"}
}

func createFailThresholdRule(value float64) dynatrace.VisualizationThresholdRule {
	return createFailThresholdRuleWithPointer(&value)
}

func createFailThresholdRuleWithPointer(value *float64) dynatrace.VisualizationThresholdRule {
	return dynatrace.VisualizationThresholdRule{Value: value, Color: "#dc172a"}
}

func createVisibleThresholds(rule1 dynatrace.VisualizationThresholdRule, rule2 dynatrace.VisualizationThresholdRule, rule3 dynatrace.VisualizationThresholdRule) dynatrace.VisualizationThreshold {
	return dynatrace.VisualizationThreshold{
		Rules:   []dynatrace.VisualizationThresholdRule{rule1, rule2, rule3},
		Visible: true,
	}
}

func createNotVisibleThresholds(rule1 dynatrace.VisualizationThresholdRule, rule2 dynatrace.VisualizationThresholdRule, rule3 dynatrace.VisualizationThresholdRule) dynatrace.VisualizationThreshold {
	return dynatrace.VisualizationThreshold{
		Rules:   []dynatrace.VisualizationThresholdRule{rule1, rule2, rule3},
		Visible: false,
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_MultipleTileConfigurationProblems tests that a Data Explorer tile with multiple configuration problems results in an error that includes all these problems.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_MultipleTileConfigurationProblems(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/multiple_tile_configuration_problems/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("srt", "error parsing SLO definition", "tile has 2 queries enabled but only one is supported", "tile has no metric expressions"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_PickCorrectVisualConfigRule tests that the visual config rule corresponding to the query is used and others are ignored.
func TestRetrieveMetricsFromDashboardDataExplorerTile_PickCorrectVisualConfigRule(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/pick_correct_visual_config_rule/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	expectedMetricsRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"),
		createToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt_milliseconds", 54.896485186544574, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_TwoMatchingVisualConfigRulesProducesError tests that two matchings visual config rules result in the expected error
func TestRetrieveMetricsFromDashboardDataExplorerTile_TwoMatchingVisualConfigRulesProducesError(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/error_two_matching_visual_config_rules/"

	handler := createHandlerWithDashboard(t, testDataFolder)
	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("srt", "expected one visualization rule for query", "found 2"))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_AllowDuplicateNames tests that unit transform works as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTile_AllowDuplicateNames(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/basic/duplicate_names/"

	requestBuilder := newMetricsV2QueryRequestBuilder("(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names")

	expectedMetricsRequest := requestBuilder.copyWithResolution(resolutionInf).build()
	const expectedSLIValue = 54896.486143787944
	const expectedDuplicateSLINameSubstring = "duplicate SLI name;"
	const expectedDuplicateDisplayNameSubstring = "duplicate display name;"
	const expectedDuplicateSLIAndDisplayNameSubstring = "duplicate SLI and display name;"

	const (
		srtSLIName      = "srt"
		srt1SLIName     = "srt1"
		srt2SLIName     = "srt2"
		srtDisplayName  = "Service response time"
		srt1DisplayName = "Service response time 1"
		srt2DisplayName = "Service response time 2"
	)

	tests := []struct {
		name                                      string
		tileName1                                 string
		tileName2                                 string
		expectedSLIResultsAssertionsFuncs         []func(t *testing.T, actual sliResult)
		expectedUploadedSLOsAssertionsFunc        func(t *testing.T, actual *keptn.ServiceLevelObjectives)
		expectedGetSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData)
	}{
		{
			name:      "no duplication of SLI or display name",
			tileName1: makeTileTitleWithSLIAndDisplayNames(srt1SLIName, srt1DisplayName),
			tileName2: makeTileTitleWithSLIAndDisplayNames(srt2SLIName, srt2DisplayName),
			expectedSLIResultsAssertionsFuncs: []func(t *testing.T, actual sliResult){
				createSuccessfulSLIResultAssertionsFunc(srt1SLIName, expectedSLIValue, expectedMetricsRequest),
				createSuccessfulSLIResultAssertionsFunc(srt2SLIName, expectedSLIValue, expectedMetricsRequest),
			},

			expectedUploadedSLOsAssertionsFunc: func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
				if !assert.EqualValues(t, 2, len(actual.Objectives)) {
					return
				}

				assert.EqualValues(t, createExpectedSLO(srt1SLIName, srt1DisplayName), actual.Objectives[0])
				assert.EqualValues(t, createExpectedSLO(srt2SLIName, srt2DisplayName), actual.Objectives[1])
			},
			expectedGetSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
		},
		{
			name:      "duplicate SLI names",
			tileName1: makeTileTitleWithSLIAndDisplayNames(srtSLIName, srt1DisplayName),
			tileName2: makeTileTitleWithSLIAndDisplayNames(srtSLIName, srt2DisplayName),
			expectedSLIResultsAssertionsFuncs: []func(t *testing.T, actual sliResult){
				createFailedSLIResultWithQueryAssertionsFunc(srtSLIName, expectedMetricsRequest, expectedDuplicateSLINameSubstring),
				createFailedSLIResultWithQueryAssertionsFunc(srtSLIName, expectedMetricsRequest, expectedDuplicateSLINameSubstring),
			},

			expectedUploadedSLOsAssertionsFunc: func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
				if !assert.EqualValues(t, 2, len(actual.Objectives)) {
					return
				}

				assert.EqualValues(t, createExpectedSLO(srtSLIName, srt1DisplayName), actual.Objectives[0])
				assert.EqualValues(t, createExpectedSLO(srtSLIName, srt2DisplayName), actual.Objectives[1])
			},
			expectedGetSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:      "duplicate display names",
			tileName1: makeTileTitleWithSLIAndDisplayNames(srt1SLIName, srtDisplayName),
			tileName2: makeTileTitleWithSLIAndDisplayNames(srt2SLIName, srtDisplayName),
			expectedSLIResultsAssertionsFuncs: []func(t *testing.T, actual sliResult){
				createFailedSLIResultWithQueryAssertionsFunc(srt1SLIName, expectedMetricsRequest, expectedDuplicateDisplayNameSubstring),
				createFailedSLIResultWithQueryAssertionsFunc(srt2SLIName, expectedMetricsRequest, expectedDuplicateDisplayNameSubstring),
			},

			expectedUploadedSLOsAssertionsFunc: func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
				if !assert.EqualValues(t, 2, len(actual.Objectives)) {
					return
				}

				assert.EqualValues(t, createExpectedSLO(srt1SLIName, srtDisplayName), actual.Objectives[0])
				assert.EqualValues(t, createExpectedSLO(srt2SLIName, srtDisplayName), actual.Objectives[1])
			},
			expectedGetSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:      "duplicate SLI and display names",
			tileName1: makeTileTitleWithSLIAndDisplayNames(srtSLIName, srtDisplayName),
			tileName2: makeTileTitleWithSLIAndDisplayNames(srtSLIName, srtDisplayName),
			expectedSLIResultsAssertionsFuncs: []func(t *testing.T, actual sliResult){
				createFailedSLIResultWithQueryAssertionsFunc(srtSLIName, expectedMetricsRequest, expectedDuplicateSLIAndDisplayNameSubstring),
				createFailedSLIResultWithQueryAssertionsFunc(srtSLIName, expectedMetricsRequest, expectedDuplicateSLIAndDisplayNameSubstring),
			},

			expectedUploadedSLOsAssertionsFunc: func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
				if !assert.EqualValues(t, 2, len(actual.Objectives)) {
					return
				}

				assert.EqualValues(t, createExpectedSLO(srtSLIName, srtDisplayName), actual.Objectives[0])
				assert.EqualValues(t, createExpectedSLO(srtSLIName, srtDisplayName), actual.Objectives[1])
			},
			expectedGetSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				filepath.Join(testDataFolder, dashboardTemplateFilename),
				struct {
					TileName1 string
					TileName2 string
				}{
					TileName1: tt.tileName1,
					TileName2: tt.tileName2,
				})

			addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testDataFolder, requestBuilder)

			runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, tt.expectedGetSLIFinishedEventAssertionsFunc, tt.expectedUploadedSLOsAssertionsFunc, tt.expectedSLIResultsAssertionsFuncs...)
		})
	}
}

func makeTileTitleWithSLIAndDisplayNames(sliName string, displayName string) string {
	return fmt.Sprintf("%s; sli=%s", displayName, sliName)
}

func createExpectedSLO(sliName string, displayName string) *keptnapi.SLO {
	return &keptnapi.SLO{
		SLI:         sliName,
		DisplayName: displayName,
		Weight:      1,
		KeySLI:      false,
	}
}

func createExpectedServiceResponseTimeSLO(passCriteria []*keptnapi.SLOCriteria, warningCriteria []*keptnapi.SLOCriteria) *keptnapi.SLO {
	return &keptnapi.SLO{
		SLI:         "srt",
		DisplayName: "Service Response Time",
		Pass:        passCriteria,
		Warning:     warningCriteria,
		Weight:      1,
		KeySLI:      false,
	}
}

func createBandSLOCriteria(lowerBoundInclusive float64, upperBoundExclusive float64) []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{{Criteria: []string{createGreaterThanOrEqualSLOCriterion(lowerBoundInclusive), createLessThanSLOCriterion(upperBoundExclusive)}}}
}

func createLowerBoundSLOCriteria(lowerBoundInclusive float64) []*keptnapi.SLOCriteria {
	return []*keptnapi.SLOCriteria{{Criteria: []string{createGreaterThanOrEqualSLOCriterion(lowerBoundInclusive)}}}
}

func createGreaterThanOrEqualSLOCriterion(v float64) string {
	return fmt.Sprintf(">=%f", v)
}

func createLessThanSLOCriterion(v float64) string {
	return fmt.Sprintf("<%f", v)
}

func convertToJSONStringOrEmptyIfNil[T any](t *testing.T, o *T) string {
	if o == nil {
		return ""
	}
	return convertToJSONString(t, *o)
}

func convertToJSONString[T any](t *testing.T, o T) string {
	bytes, err := json.Marshal(o)
	if err != nil {
		t.Fatal("could not marshal object to JSON")
	}
	return string(bytes)
}

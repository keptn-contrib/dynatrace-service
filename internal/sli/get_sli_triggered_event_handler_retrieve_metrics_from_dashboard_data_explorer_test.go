package sli

import (
	"encoding/json"
	"fmt"
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

// TestRetrieveMetricsFromDashboardDataExplorerTile_MetricExpressions tests support of metric expressions works as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTile_MetricExpressions(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/metric_expressions/"

	const singleValueVisualConfigType = "SINGLE_VALUE"
	const graphChartVisualConfigType = "GRAPH_CHART"
	const splitByService = "dt.entity.service"

	const singleResultMetricExpression = "(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	const multipleResultMetricExpression = "(builtin:service.response.time:splitBy(\"dt.entity.service\"):avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"

	const resolutionIsNullKeyValuePair = "resolution=null&"

	const singleQueryResultFilename = testDataFolder + "metrics_query_builtin_service_response_time_avg_single.json"
	const multipleQueryResultsFilename = testDataFolder + "metrics_query_builtin_service_response_time_avg_multiple.json"

	expectedSingleResultMetricsRequest := buildMetricsV2RequestString("%28builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames")
	expectedMultipleResultMetricsRequest := buildMetricsV2RequestString("%28builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames")

	singleSuccessfulSLIResultAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt", 29192.929640271974, expectedSingleResultMetricsRequest)}
	multipleSuccessfulSLIResultAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt_service_a", 31676.5399830501183, expectedMultipleResultMetricsRequest),
		createSuccessfulSLIResultAssertionsFunc("srt_service_b", 11285.1679389312976, expectedMultipleResultMetricsRequest)}

	tests := []struct {
		name                              string
		visualConfigType                  string
		splitBy                           string
		metricExpressions                 *[]string
		expectedMetricsRequest            string
		queryResultsFilename              string
		getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData)
		sliResultsAssertionsFuncs         []func(t *testing.T, actual sliResult)
	}{
		{
			name:                              "single value visualization with single result works",
			visualConfigType:                  singleValueVisualConfigType,
			metricExpressions:                 &[]string{resolutionIsNullKeyValuePair + singleResultMetricExpression, singleResultMetricExpression},
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultsAssertionsFuncs:         singleSuccessfulSLIResultAssertionsFuncs,
		},
		{
			name:                              "graph chart visualization with single result works",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{resolutionIsNullKeyValuePair + singleResultMetricExpression},
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultsAssertionsFuncs:         singleSuccessfulSLIResultAssertionsFuncs,
		},
		{
			name:                              "graph chart visualization with multiple result works",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{resolutionIsNullKeyValuePair + multipleResultMetricExpression},
			queryResultsFilename:              multipleQueryResultsFilename,
			expectedMetricsRequest:            expectedMultipleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultsAssertionsFuncs:         multipleSuccessfulSLIResultAssertionsFuncs,
		},
		{
			name:                              "single value visualization with multiple results produces error",
			visualConfigType:                  singleValueVisualConfigType,
			metricExpressions:                 &[]string{resolutionIsNullKeyValuePair + multipleResultMetricExpression, multipleResultMetricExpression},
			queryResultsFilename:              multipleQueryResultsFilename,
			expectedMetricsRequest:            expectedMultipleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings(expectedMultipleResultMetricsRequest, "tile is configured for single value"),
		},
		{
			name:                              "graph chart visualization with no metric expressions produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 nil,
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "tile has no metric expressions"),
		},
		{
			name:                              "zero metric expressions produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{},
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "tile has no metric expressions"),
		},
		{
			name:                              "missing resolution key value pair produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{singleResultMetricExpression},
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "metric expression does not contain two components"),
		},
		{
			name:                              "wrong order in metric expression produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{singleResultMetricExpression + resolutionIsNullKeyValuePair},
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "unexpected prefix in key value pair"),
		},
		{
			name:                              "empty value in resolution key value pair produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{"resolution=&" + singleResultMetricExpression},
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "resolution must not be empty"),
		},
		{
			name:                              "resolution other than null (inf) produces error",
			visualConfigType:                  graphChartVisualConfigType,
			metricExpressions:                 &[]string{"resolution=30m&" + singleResultMetricExpression},
			queryResultsFilename:              singleQueryResultFilename,
			expectedMetricsRequest:            expectedSingleResultMetricsRequest,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultsAssertionsFuncs:         createSRTFailedSLIResultsAssertionsFuncsWithErrorSubstrings("", "resolution must be set to 'Auto'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				testDataFolder+"dashboard.template.json",
				struct {
					SplitBy                 string
					VisualConfigType        string
					MetricExpressionsString string
				}{
					SplitBy:                 tt.splitBy,
					VisualConfigType:        tt.visualConfigType,
					MetricExpressionsString: convertToJSONStringOrEmptyIfNil(t, tt.metricExpressions),
				},
			)
			handler.AddExactFile(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
			handler.AddExactFile(tt.expectedMetricsRequest, tt.queryResultsFilename)

			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, tt.getSLIFinishedEventAssertionsFunc, tt.sliResultsAssertionsFuncs...)
		})
	}
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
			ID:   "-1234567890123456789",
			Name: "mz-1",
		},
	}

	emptyTileFilter := dynatrace.TileFilter{}

	tileFilterWithManagementZone := dynatrace.TileFilter{
		ManagementZone: &dynatrace.ManagementZoneEntry{
			ID:   "2311420533206603714",
			Name: "ap_mz_1",
		},
	}

	expectedEncodedMetricSelector := "%28builtin%3Aservice.response.time%3AsplitBy%28%29%3Asort%28value%28auto%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames"

	tests := []struct {
		name                   string
		dashboardFilter        *dynatrace.DashboardFilter
		tileFilter             dynatrace.TileFilter
		expectedMetricsRequest string
	}{
		{
			name:                   "no dashboard filter, empty tile filter",
			dashboardFilter:        nil,
			tileFilter:             emptyTileFilter,
			expectedMetricsRequest: buildMetricsV2RequestString(expectedEncodedMetricSelector),
		},
		{
			name:                   "dashboard filter with mz, empty tile filter",
			dashboardFilter:        &dashboardFilterWithManagementZone,
			tileFilter:             emptyTileFilter,
			expectedMetricsRequest: buildMetricsV2RequestStringWithMZSelector(expectedEncodedMetricSelector, "mzId%28-1234567890123456789%29"),
		},
		{
			name:                   "no dashboard filter, tile filter with mz",
			dashboardFilter:        nil,
			tileFilter:             tileFilterWithManagementZone,
			expectedMetricsRequest: buildMetricsV2RequestStringWithMZSelector(expectedEncodedMetricSelector, "mzId%282311420533206603714%29"),
		},
		{
			name:                   "dashboard filter with mz, tile filter with mz",
			dashboardFilter:        &dashboardFilterWithManagementZone,
			tileFilter:             tileFilterWithManagementZone,
			expectedMetricsRequest: buildMetricsV2RequestStringWithMZSelector(expectedEncodedMetricSelector, "mzId%282311420533206603714%29"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				testDataFolder+"dashboard.template.json",
				struct {
					DashboardFilterString string
					TileFilterString      string
				}{
					DashboardFilterString: convertToJSONStringOrEmptyIfNil(t, tt.dashboardFilter),
					TileFilterString:      convertToJSONString(t, tt.tileFilter),
				},
			)
			handler.AddExactFile(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
			handler.AddExactFile(tt.expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("srt", 8283.891270010905, tt.expectedMetricsRequest))
		})
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZoneWithNoEntityType tests that an error is produced for data explorer tiles with a management zone and no obvious entity type.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZoneWithNoEntityType(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/no_entity_type/"

	expectedMetricsRequest := buildMetricsV2RequestStringWithMZSelector("%28builtin%3Asecurity.securityProblem.open.managementZone%3Afilter%28and%28or%28eq%28%22Risk+Level%22%2CHIGH%29%29%29%29%3AsplitBy%28%22Risk+Level%22%29%3Asum%3Aauto%3Asort%28value%28sum%2Cdescending%29%29%3Alimit%28100%29%29%3Alimit%28100%29%3Anames", "mzId%28-1234567890123456789%29")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:security.securityProblem.open.managementZone", testDataFolder+"metrics_builtin_security_securityProblem_open_managementZone.json")
	handler.AddExactError(expectedMetricsRequest, 400, testDataFolder+"metrics_query_builtin_security_securityProblem_open_managementZone.json")

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc("vulnerabilities_high", expectedMetricsRequest))
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_CustomSLO tests propagation of a customized SLO.
// This is will result in a SLIResult with success, as this is supported.
// Here also the SLO is checked, including the display name, weight and key SLI.
func TestRetrieveMetricsFromDashboardDataExplorerTile_CustomSLO(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/custom_slo/"

	expectedMetricsRequest := buildMetricsV2RequestString("%28builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("srt", 29192.929640271974, expectedMetricsRequest),
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

	expectedMetricsRequest := buildMetricsV2RequestString("%28builtin%3Aservice.response.time%3Afilter%28and%28or%28in%28%22dt.entity.service%22%2CentitySelector%28%22type%28service%29%2CentityId%28~%22SERVICE-B67B3EC4C95E0FA7~%22%29%22%29%29%29%29%29%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard_excluded_tile.json")
	handler.AddExact(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
	handler.AddExact(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("rt_jid", 136528.52484946526, expectedMetricsRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdsWork tests that setting pass and warning criteria via thresholds on the tile works as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdsWork(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/tile_thresholds_success/"

	expectedMetricsRequest := buildMetricsV2RequestString("%28builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%3Alimit%2810%29%29%3Alimit%28100%29%3Anames")

	successfulSLIResultAssertionsFunc := createSuccessfulSLIResultAssertionsFunc("srt", 29192.929640271974, expectedMetricsRequest)

	tests := []struct {
		name        string
		tileName    string
		thresholds  dynatrace.Threshold
		expectedSLO *keptnapi.SLO
	}{
		{
			name:        "Valid pass-warn-fail thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createPassThresholdRule(0), createWarnThresholdRule(68000), createFailThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createBandSLOCriteria(0, 68000), createBandSLOCriteria(0, 69000)),
		},
		{
			name:        "Valid fail-warn-pass thresholds and no pass or warning defined in title",
			tileName:    "Service Response Time; sli=srt",
			thresholds:  createVisibleThresholds(createFailThresholdRule(0), createWarnThresholdRule(68000), createPassThresholdRule(69000)),
			expectedSLO: createExpectedServiceResponseTimeSLO(createLowerBoundSLOCriteria(69000), createLowerBoundSLOCriteria(68000)),
		},
		{
			name:       "Pass or warning defined in title take precedence over valid thresholds ",
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
				"./testdata/dashboards/data_explorer/tile_thresholds_success/dashboard.template.json",
				struct {
					TileName         string
					ThresholdsString string
				}{
					TileName:         thresholdTest.tileName,
					ThresholdsString: convertToJSONString(t, thresholdTest.thresholds),
				})
			handler.AddExactFile(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
			handler.AddExactFile(expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time_avg.json")

			uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
				if assert.Equal(t, 1, len(actual.Objectives)) {
					assert.EqualValues(t, thresholdTest.expectedSLO, actual.Objectives[0])
				}
			}

			runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, successfulSLIResultAssertionsFunc)
		})
	}
}

func createPassThresholdRule(value float64) dynatrace.ThresholdRule {
	return createPassThresholdRuleWithPointer(&value)
}

func createPassThresholdRuleWithPointer(value *float64) dynatrace.ThresholdRule {
	return dynatrace.ThresholdRule{Value: value, Color: "#7dc540"}
}

func createWarnThresholdRule(value float64) dynatrace.ThresholdRule {
	return createWarnThresholdRuleWithPointer(&value)
}

func createWarnThresholdRuleWithPointer(value *float64) dynatrace.ThresholdRule {
	return dynatrace.ThresholdRule{Value: value, Color: "#f5d30f"}
}

func createFailThresholdRule(value float64) dynatrace.ThresholdRule {
	return createFailThresholdRuleWithPointer(&value)
}

func createFailThresholdRuleWithPointer(value *float64) dynatrace.ThresholdRule {
	return dynatrace.ThresholdRule{Value: value, Color: "#dc172a"}
}

func createVisibleThresholds(rule1 dynatrace.ThresholdRule, rule2 dynatrace.ThresholdRule, rule3 dynatrace.ThresholdRule) dynatrace.Threshold {
	return dynatrace.Threshold{
		Rules:   []dynatrace.ThresholdRule{rule1, rule2, rule3},
		Visible: true,
	}
}

func createNotVisibleThresholds(rule1 dynatrace.ThresholdRule, rule2 dynatrace.ThresholdRule, rule3 dynatrace.ThresholdRule) dynatrace.Threshold {
	return dynatrace.Threshold{
		Rules:   []dynatrace.ThresholdRule{rule1, rule2, rule3},
		Visible: false,
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_UnitTransformIsNotAuto tests that unit transforms other than auto are not allowed.
// This is will result in a SLIResult with failure, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_UnitTransformIsNotAuto(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, "./testdata/dashboards/data_explorer/unit_transform_is_not_auto/dashboard.json")

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc("srt", "must be set to 'Auto'"))
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

func createHandlerWithTemplatedDashboard(t *testing.T, templateFilename string, templatingData interface{}) *test.CombinedURLHandler {
	handler := test.NewCombinedURLHandler(t)
	handler.AddExactTemplate(dynatrace.DashboardsPath+"/"+testDashboardID, templateFilename, templatingData)
	return handler
}

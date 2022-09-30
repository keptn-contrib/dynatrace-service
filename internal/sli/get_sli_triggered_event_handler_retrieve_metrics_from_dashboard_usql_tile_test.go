package sli

import (
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
)

// TestRetrieveMetricsFromDashboardUSQLTile_ColumnChart tests that extracting SLIs from a USQL tile with column chart visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_ColumnChart(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/column_chart_visualization/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+browserFamily%2C+AVG%28useraction.visuallyCompleteTime%29%2C+AVG%28useraction.domCompleteTime%29%2C+AVG%28totalErrorCount%29+FROM+usersession+GROUP+BY+browserFamily+LIMIT+3")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric_aol_explorer", 1428.282447519664, expectedUSQLRequest),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_null", 1402.0759668508288, expectedUSQLRequest),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_acoo_browser", 1427.2675735742882, expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_PieChart tests that extracting SLIs from a USQL tile with pie chart visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_PieChart(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/pie_chart_visualization/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+city%2C+AVG%28duration%29+FROM+usersession+GROUP+BY+city+LIMIT+3")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric_null", 85673.22361214698, expectedUSQLRequest),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_ashburn", 73031.44078516903, expectedUSQLRequest),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_london", 88003.95643322475, expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue tests that extracting a single SLI from a USQL tile with single value visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+AVG%28duration%29+FROM+usersession")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric", 87793.1776896271, expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table tests that extracting SLIs from a USQL tile with table visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_Table(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+continent%2C+totalErrorCount%2C+totalLicenseCreditCount%2C+userActionCount+FROM+usersession+limit+2")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric_asia", 2, expectedUSQLRequest),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_north_america", 5, expectedUSQLRequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
		if !assert.EqualValues(t, 2, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "usql_metric_north_america",
			DisplayName: "User sessions query results (North America)",
			Pass:        []*keptnapi.SLOCriteria{{Criteria: []string{"<=100"}}},
			Weight:      1,
		}, actual.Objectives[0])

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "usql_metric_europe",
			DisplayName: "User sessions query results (Europe)",
			Pass:        []*keptnapi.SLOCriteria{{Criteria: []string{"<=100"}}},
			Weight:      1,
		}, actual.Objectives[1])

	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_LineChart tests that extracting SLIs from a USQL tile with line chart visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_LineChart(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/line_chart_visualization/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+continent%2C+userActionCount+FROM+usersession+limit+2")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric_asia", 2, expectedUSQLRequest),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_north_america", 5, expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Funnel tests that extracting SLIs from a USQL tile with funnel visualization type does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_Funnel(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/funnel_visualization/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+FUNNEL%28useraction.name%3D%22AppStart+%28easyTravel%29%22+AS+%22Open+easytravel%22%2C+useraction.name+%3D+%22searchJourney%22+AS+%22Search+journey%22%2C+useraction.name+%3D+%22bookJourney%22+AS+%22Book+journey%22%29+FROM+usersession")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_NoQuery tests that extracting SLIs from a USQL tile with no query does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_NoQuery(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/tile_with_no_query/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_MissingScopes tests that extracting SLIs from a USQL tile with missing API token scopes does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_MissingScopes(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/missing_scopes/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+AVG%28duration%29+FROM+usersession")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExactError(expectedUSQLRequest, 403, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiColumns tests that extracting SLIs from a USQL tile with single value visualization type and a multi-column result produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiColumns(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization_multi_columns/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+city%2C+AVG%28duration%29+FROM+usersession+GROUP+BY+city+LIMIT+3")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiRows tests that extracting SLIs from a USQL tile with single value visualization type and a multi-row result produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiRows(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization_multi_rows/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+AVG%28duration%29+FROM+usersession+GROUP+BY+city+LIMIT+3")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_InvalidResultType tests that extracting SLIs from a USQL tile with single value visualization type with an invalid result type produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_InvalidResultType(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization_invalid_result_type/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+city+FROM+usersession+GROUP+BY+city+LIMIT+1")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_NoValues tests that extracting SLIs from a USQL tile with single value visualization type with no values produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_NoValues(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization_no_values/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+city+FROM+usersession+GROUP+BY+city+LIMIT+1")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table_NotEnoughColumns tests that extracting SLIs from a USQL tile with table visualization type with not enough columns produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_Table_NotEnoughColumns(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization_not_enough_columns/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+city+FROM+usersession+GROUP+BY+city+LIMIT+3")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionName tests that extracting SLIs from a USQL tile with table visualization type with non-string dimension names produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionName(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization_invalid_dimension_name/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+totalErrorCount%2C+totalLicenseCreditCount%2C+userActionCount+FROM+usersession+LIMIT+2")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionValue tests that extracting SLIs from a USQL tile with table visualization type with non-numerical dimension values produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionValue(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization_invalid_dimension_value/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+totalErrorCount%2C+city+FROM+usersession+LIMIT+2")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table_NoValues tests that extracting SLIs from a USQL tile with table visualization type with no values produces a warning.
func TestRetrieveMetricsFromDashboardUSQLTile_Table_NoValues(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization_no_values/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+totalErrorCount%2C+city+FROM+usersession+LIMIT+2")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("usql_metric", expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventWarningAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_CustomSLO tests that extracting an SLI and SLO from a USQL tile works as expected.
// This is will result in a SLIResult with success, as this is supported.
// Here also the SLO is checked, including the display name, weight and key SLI.
func TestRetrieveMetricsFromDashboardUSQLTile_CustomSLO(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/custom_slo/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+AVG%28duration%29+FROM+usersession")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("average_session_duration", 87793.1776896271, expectedUSQLRequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptn.ServiceLevelObjectives) {
		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:         "average_session_duration",
			DisplayName: "Average user session duration",
			Pass:        []*keptnapi.SLOCriteria{{Criteria: []string{"<=100"}}},
			Weight:      2,
			KeySLI:      true,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_ExcludedTile tests that a tile with exclude set to true is skipped.
func TestRetrieveMetricsFromDashboardUSQLTile_ExcludedTile(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/usql_tiles/excluded_tile/"

	expectedUSQLRequest := buildUSQLRequest("SELECT+AVG%28duration%29+FROM+usersession")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_result_table.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric", 87793.1776896271, expectedUSQLRequest),
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
}

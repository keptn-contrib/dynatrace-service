package sli

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

var testUSQLTileGetSLIEventData = createTestGetSLIEventDataWithStartAndEnd("2021-09-17T07:00:00.000Z", "2021-09-17T08:00:00.000Z")

// TestRetrieveMetricsFromDashboardUSQLTile_ColumnChart tests that extracting SLIs from a USQL tile with column chart visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_ColumnChart(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/column_chart_visualization/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+browserFamily%2C+AVG%28useraction.visuallyCompleteTime%29%2C+AVG%28useraction.domCompleteTime%29%2C+AVG%28totalErrorCount%29+FROM+usersession+GROUP+BY+browserFamily+LIMIT+3&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric_null", 492.6603364080304),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_AOL_Explorer", 500.2868314283638),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_Acoo_Browser", 500.5150319856381),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_null", "USQL;COLUMN_CHART;null;SELECT browserFamily, AVG(useraction.visuallyCompleteTime), AVG(useraction.domCompleteTime), AVG(totalErrorCount) FROM usersession GROUP BY browserFamily LIMIT 3")
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_AOL_Explorer", "USQL;COLUMN_CHART;AOL Explorer;SELECT browserFamily, AVG(useraction.visuallyCompleteTime), AVG(useraction.domCompleteTime), AVG(totalErrorCount) FROM usersession GROUP BY browserFamily LIMIT 3")
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_Acoo_Browser", "USQL;COLUMN_CHART;Acoo Browser;SELECT browserFamily, AVG(useraction.visuallyCompleteTime), AVG(useraction.domCompleteTime), AVG(totalErrorCount) FROM usersession GROUP BY browserFamily LIMIT 3")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testUSQLTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_PieChart tests that extracting SLIs from a USQL tile with pie chart visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_PieChart(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/pie_chart_visualization/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+city%2C+AVG%28duration%29+FROM+usersession+GROUP+BY+city+LIMIT+3&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric_null", 60154.328623114205),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_Ashburn", 53567.040172786175),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_Beijing", 65199.979558462794),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_null", "USQL;PIE_CHART;null;SELECT city, AVG(duration) FROM usersession GROUP BY city LIMIT 3")
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_Ashburn", "USQL;PIE_CHART;Ashburn;SELECT city, AVG(duration) FROM usersession GROUP BY city LIMIT 3")
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_Beijing", "USQL;PIE_CHART;Beijing;SELECT city, AVG(duration) FROM usersession GROUP BY city LIMIT 3")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testUSQLTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue tests that extracting a single SLI from a USQL tile with single value visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+AVG%28duration%29+FROM+usersession&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric", 62731.12784806213),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "usql_metric", "USQL;SINGLE_VALUE;;SELECT AVG(duration) FROM usersession")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testUSQLTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table tests that extracting SLIs from a USQL tile with table visualization type works as expected.
func TestRetrieveMetricsFromDashboardUSQLTile_Table(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+continent%2C+totalErrorCount%2C+totalLicenseCreditCount%2C+userActionCount+FROM+usersession+limit+2&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createSuccessfulSLIResultAssertionsFunc("usql_metric_North_America", 1),
		createSuccessfulSLIResultAssertionsFunc("usql_metric_Europe", 2),
	}

	uploadedSLIsAssertionsFunc := func(t *testing.T, actual *dynatrace.SLI) {
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_North_America", "USQL;TABLE;North America;SELECT continent, totalErrorCount, totalLicenseCreditCount, userActionCount FROM usersession limit 2")
		assertSLIDefinitionIsPresent(t, actual, "usql_metric_Europe", "USQL;TABLE;Europe;SELECT continent, totalErrorCount, totalLicenseCreditCount, userActionCount FROM usersession limit 2")
	}

	runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testUSQLTileGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLIsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_LineChart tests that extracting SLIs from a USQL tile with line chart visualization type does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_LineChart(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/line_chart_visualization/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+continent%2C+userActionCount+FROM+usersession+limit+2&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Funnel tests that extracting SLIs from a USQL tile with funnel visualization type does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_Funnel(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/funnel_visualization/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+FUNNEL%28useraction.name%3D%22AppStart+%28easyTravel%29%22+AS+%22Open+easytravel%22%2C+useraction.name+%3D+%22searchJourney%22+AS+%22Search+journey%22%2C+useraction.name+%3D+%22bookJourney%22+AS+%22Book+journey%22%29+FROM+usersession&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_NoQuery tests that extracting SLIs from a USQL tile with no query does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_NoQuery(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/tile_with_no_query/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_MissingScopes tests that extracting SLIs from a USQL tile with missing API token scopes does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_MissingScopes(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/missing_scopes/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExactError(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+AVG%28duration%29+FROM+usersession&startTimestamp=1631862000000", 403, testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiColumns tests that extracting SLIs from a USQL tile with single value visualization type and a multi-column result does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiColumns(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization_multi_columns/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+city%2C+AVG%28duration%29+FROM+usersession+GROUP+BY+city+LIMIT+3&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiRows tests that extracting SLIs from a USQL tile with single value visualization type and a multi-row result does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_MultiRows(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization_multi_rows/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+AVG%28duration%29+FROM+usersession+GROUP+BY+city+LIMIT+3&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_InvalidResultType tests that extracting SLIs from a USQL tile with single value visualization type with an invalid result type does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_SingleValue_InvalidResultType(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/single_value_visualization_invalid_result_type/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+city+FROM+usersession+GROUP+BY+city+LIMIT+1&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table_NotEnoughColumns tests that extracting SLIs from a USQL tile with table visualization type with not enough columns does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_Table_NotEnoughColumns(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization_not_enough_columns/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+city+FROM+usersession+GROUP+BY+city+LIMIT+3&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionName tests that extracting SLIs from a USQL tile with table visualization type with non-string dimension names does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionName(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization_invalid_dimension_name/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+totalErrorCount%2C+totalLicenseCreditCount%2C+userActionCount+FROM+usersession+LIMIT+2&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionValue tests that extracting SLIs from a USQL tile with table visualization type with non-numerical dimension values does not work.
func TestRetrieveMetricsFromDashboardUSQLTile_Table_InvalidDimensionValue(t *testing.T) {

	const testDataFolder = "./testdata/dashboards/usql_tiles/table_visualization_invalid_dimension_value/"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, testDataFolder+"dashboard.json")
	handler.AddExact(dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1631865600000&explain=false&query=SELECT+totalErrorCount%2C+city+FROM+usersession+LIMIT+2&startTimestamp=1631862000000", testDataFolder+"usql_result_table.json")

	sliResultsAssertionsFuncs := []func(t *testing.T, actual *keptnv2.SLIResult){
		createFailedSLIResultAssertionsFunc("usql_metric"),
		createFailedSLIResultAssertionsFunc("usql_metric"),
	}

	rClient := &uploadErrorResourceClientMock{t: t}
	runAndAssertThatDashboardTestIsCorrect(t, testUSQLTileGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, sliResultsAssertionsFuncs...)
}

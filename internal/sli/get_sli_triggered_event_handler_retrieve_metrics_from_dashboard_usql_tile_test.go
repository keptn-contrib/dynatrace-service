package sli

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
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

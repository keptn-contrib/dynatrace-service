package sli

import (
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// TestRetrieveMetricsFromDashboardDataExplorerTile_Errors test lots of data explorer tiles with errors in the SLO definition.
// This will result in a SLIResult with failures, as this is not allowed.
func TestRetrieveMetricsFromDashboardDataExplorerTile_Errors(t *testing.T) {
	type data struct {
		TileTitle string // needs to match template file variable
	}

	const dashboardTemplateFile = "./testdata/dashboards/data_explorer/tile_title_errors/data-explorer-tile-title-parsing-errors-template.json"

	tests := []struct {
		tileTitle      string
		assertionsFunc func(*testing.T, *keptnv2.SLIResult)
	}{
		{
			tileTitle:      "empty sli name;sli=;pass=<500;pass=<600,<+5%;warning=<800,<+10%;weight=10;key=true;",
			assertionsFunc: createFailedSLIResultAssertionsFunc("empty sli name;sli=;pass=<500;pass=<600,<+5%;warning=<800,<+10%;weight=10;key=true;", "sli name is empty"),
		},
		{
			tileTitle:      "duplicate weight;sli=duplicate weight;pass=<500;weight=10;key=true;weight=1",
			assertionsFunc: createFailedSLIResultAssertionsFunc("duplicate weight", "duplicate key", "'weight'"),
		},
		{
			tileTitle:      "duplicate key;sli=duplicate key;pass=<500;weight=10;key=true;key=false",
			assertionsFunc: createFailedSLIResultAssertionsFunc("duplicate key", "duplicate key", "'key'"),
		},
		{
			tileTitle:      "duplicate sli;sli=duplicate sli;pass=<500;sli=other name",
			assertionsFunc: createFailedSLIResultAssertionsFunc("duplicate sli", "duplicate key", "'sli'"),
		},
		{
			tileTitle: "empty sli name;sli=;pass=<500;pass=<600,<+5%;warning=<800,<+10%;weight=10;key=true;",
			// full tile name will be used here, because no SLI name could be extracted
			assertionsFunc: createFailedSLIResultAssertionsFunc("empty sli name;sli=;pass=<500;pass=<600,<+5%;warning=<800,<+10%;weight=10;key=true;", "sli name is empty"),
		},
		{
			tileTitle: "whitespace sli name;sli=   ;pass=<500;pass=<600,<+5%;warning=<800,<+10%;weight=10;key=true;",
			// full tile name will be used here, because no SLI name could be extracted
			assertionsFunc: createFailedSLIResultAssertionsFunc("whitespace sli name;sli=   ;pass=<500;pass=<600,<+5%;warning=<800,<+10%;weight=10;key=true;", "sli name is empty"),
		},
		{
			tileTitle:      "invalid pass op;sli=invalid pass op;pass=>>7",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid pass op", "pass", ">>7"),
		},
		{
			tileTitle:      "invalid pass value;sli=invalid pass value;pass=<5ms;",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid pass value", "<5ms"),
		},
		{
			tileTitle:      "invalid pass - missing op;sli=invalid pass missing op;pass=5",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid pass missing op", "5"),
		},
		{
			tileTitle:      "invalid pass - wrong type;sli=invalid pass wrong type;pass=no!",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid pass wrong type", "no!"),
		},
		{
			tileTitle:      "invalid pass value decimal;sli=invalid pass value decimal;pass=<5.",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid pass value decimal", "<5."),
		},
		{
			tileTitle:      "invalid pass value percent;sli=invalid pass value percent;pass=<5.%",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid pass value percent", "<5.%"),
		},
		{
			tileTitle:      "invalid warning op;sli=invalid warning op;warning=<<5",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid warning op", "<<5"),
		},
		{
			tileTitle:      "invalid warning value;sli=invalid warning value;warning=<5s;",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid warning value", "<5s"),
		},
		{
			tileTitle:      "invalid warning - missing op;sli=invalid warning missing op;warning=5",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid warning missing op", "5"),
		},
		{
			tileTitle:      "invalid warning - wrong type;sli=invalid warning wrong type;warning=yes",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid warning wrong type", "yes"),
		},
		{
			tileTitle:      "invalid warning value decimal;sli=invalid warning value decimal;warning=<5.",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid warning value decimal", "<5."),
		},
		{
			tileTitle:      "invalid warning value percent;sli=invalid warning value percent;warning=<5.%",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid warning value percent", "<5.%"),
		},
		{
			tileTitle:      "invalid weight value;sli=invalid weight value;pass=<7;weight=4.",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid weight value", "4."),
		},
		{
			tileTitle:      "invalid weight value percent;sli=invalid weight value percent;pass=<7;weight=4.%",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid weight value percent", "4.%"),
		},
		{
			tileTitle:      "invalid weight value decimal;sli=invalid weight value decimal;pass=<7;weight=4.5",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid weight value decimal", "4.5"),
		},
		{
			tileTitle:      "invalid weight value string;sli=invalid weight value string;pass=<7;weight=heavy",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid weight value string", "heavy"),
		},
		{
			tileTitle:      "invalid key value string;sli=invalid key value string;pass=<7;key=no",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid key value string", "not a boolean value: no"),
		},
		{
			tileTitle:      "invalid key value number;sli=invalid key value number;pass=<7;key=3",
			assertionsFunc: createFailedSLIResultAssertionsFunc("invalid key value number", "not a boolean value: 3"),
		},
	}
	for _, dataExplorerTest := range tests {
		t.Run(dataExplorerTest.tileTitle, func(t *testing.T) {
			handler := test.NewTemplatingPayloadBasedURLHandler(t, dashboardTemplateFile)
			handler.AddExact(
				dynatrace.DashboardsPath+"/"+testDashboardID,
				&data{
					TileTitle: dataExplorerTest.tileTitle,
				},
			)

			rClient := &uploadErrorResourceClientMock{t: t}
			runAndAssertThatDashboardTestIsCorrect(t, testDataExplorerGetSLIEventData, handler, rClient, getSLIFinishedEventFailureAssertionsFunc, dataExplorerTest.assertionsFunc)
		})
	}
}

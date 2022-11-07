package sli

import (
	"path/filepath"
	"testing"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// TestRetrieveMetricsFromDashboardProblemTile_Success tests the success case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
func TestRetrieveMetricsFromDashboardProblemTile_Success(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/problem_tile_success/"

	expectedProblemsRequest := buildProblemsV2Request("status(\"open\")")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExact(expectedProblemsRequest, filepath.Join(testDataFolder, "problems_status_open.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createSuccessfulSLIResultAssertionsFunc("problems", 42, expectedProblemsRequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

// TestRetrieveMetricsFromDashboardProblemTile_ManagementZonesWork tests applying management zones to the dashboard and tile work as expected.
func TestRetrieveMetricsFromDashboardProblemTile_ManagementZonesWork(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/management_zones_work/"

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

	expectedProblemsRequestWithNoManagementZone := buildProblemsV2Request("status(\"open\")")
	expectedProblemsRequestWithManagementZone1 := buildProblemsV2Request("status(\"open\"),managementZones(\"ap_mz_1\")")
	expectedProblemsRequestWithManagementZone2 := buildProblemsV2Request("status(\"open\"),managementZones(\"ap_mz_2\")")

	tests := []struct {
		name                    string
		dashboardFilter         *dynatrace.DashboardFilter
		tileFilter              dynatrace.TileFilter
		expectedProblemsRequest string
		expectedSLIValue        float64
	}{
		{
			name:                    "no_dashboard_filter_and_empty_tile_filter",
			dashboardFilter:         nil,
			tileFilter:              emptyTileFilter,
			expectedProblemsRequest: expectedProblemsRequestWithNoManagementZone,
			expectedSLIValue:        2,
		},
		{
			name:                    "dashboard_filter_with_mz_and_empty_tile_filter",
			dashboardFilter:         &dashboardFilterWithManagementZone,
			tileFilter:              emptyTileFilter,
			expectedProblemsRequest: expectedProblemsRequestWithManagementZone1,
			expectedSLIValue:        0,
		},
		{
			name:                    "dashboard_filter_with_all_mz_and_empty_tile_filter",
			dashboardFilter:         &dashboardFilterWithAllManagementZones,
			tileFilter:              emptyTileFilter,
			expectedProblemsRequest: expectedProblemsRequestWithNoManagementZone,
			expectedSLIValue:        2,
		},
		{
			name:                    "no_dashboard_filter_and_tile_filter_with_mz",
			dashboardFilter:         nil,
			tileFilter:              tileFilterWithManagementZone,
			expectedProblemsRequest: expectedProblemsRequestWithManagementZone2,
			expectedSLIValue:        0,
		},
		{
			name:                    "dashboard_filter_with_mz_and_tile_filter_with_mz",
			dashboardFilter:         &dashboardFilterWithManagementZone,
			tileFilter:              tileFilterWithManagementZone,
			expectedProblemsRequest: expectedProblemsRequestWithManagementZone2,
			expectedSLIValue:        0,
		},
		{
			name:                    "dashboard_filter_with_all_mz_and_tile_filter_with_mz",
			dashboardFilter:         &dashboardFilterWithAllManagementZones,
			tileFilter:              tileFilterWithManagementZone,
			expectedProblemsRequest: expectedProblemsRequestWithManagementZone2,
			expectedSLIValue:        0,
		},
		{
			name:                    "no_dashboard_filter_and_tile_filter_with_all_mz",
			dashboardFilter:         nil,
			tileFilter:              tileFilterWithAllManagementZones,
			expectedProblemsRequest: expectedProblemsRequestWithNoManagementZone,
			expectedSLIValue:        2,
		},
		{
			name:                    "dashboard_filter_with_mz_and_tile_filter_with_all_mz",
			dashboardFilter:         &dashboardFilterWithManagementZone,
			tileFilter:              tileFilterWithAllManagementZones,
			expectedProblemsRequest: expectedProblemsRequestWithNoManagementZone,
			expectedSLIValue:        2,
		},
		{
			name:                    "dashboard_filter_with_all_mz_and_tile_filter_with_all_mz",
			dashboardFilter:         &dashboardFilterWithAllManagementZones,
			tileFilter:              tileFilterWithAllManagementZones,
			expectedProblemsRequest: expectedProblemsRequestWithNoManagementZone,
			expectedSLIValue:        2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				filepath.Join(testDataFolder, "dashboard.template.json"),
				struct {
					DashboardFilterString string
					TileFilterString      string
				}{
					DashboardFilterString: convertToJSONStringOrEmptyIfNil(t, tt.dashboardFilter),
					TileFilterString:      convertToJSONString(t, tt.tileFilter),
				},
			)

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)
			handler.AddExactFile(tt.expectedProblemsRequest, filepath.Join(testVariantDataFolder, "problems_status_open.json"))
			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("problems", tt.expectedSLIValue, tt.expectedProblemsRequest))
		})
	}
}

// TestRetrieveMetricsFromDashboardProblemTile_MissingScopes tests the failure case for retrieving the problem and security problem count SLIs in response to a problems dashboard tile.
// Retrieving SLIs fails because the the API token is missing the required scopes.
func TestRetrieveMetricsFromDashboardProblemTile_MissingScopes(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/problem_tile/missing_scopes/"

	expectedProblemsRequest := buildProblemsV2Request("status(\"open\")")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(dynatrace.DashboardsPath+"/"+testDashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExactError(expectedProblemsRequest, 403, filepath.Join(testDataFolder, "problems_missing_scope.json"))

	sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
		createFailedSLIResultWithQueryAssertionsFunc("problems", expectedProblemsRequest),
	}

	uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
		if !assert.NotNil(t, actual) {
			return
		}

		if !assert.EqualValues(t, 1, len(actual.Objectives)) {
			return
		}

		assert.EqualValues(t, &keptnapi.SLO{
			SLI:    "problems",
			Pass:   []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
			Weight: 1,
			KeySLI: true,
		}, actual.Objectives[0])
	}

	runGetSLIsFromDashboardTestAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
}

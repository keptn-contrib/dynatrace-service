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

func TestRetrieveMetricsFromDashboardDataExplorerTile_SplitByFilterByWorks(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/splitby_filterby_works/"

	splitByNone := []string{}
	splitByService := []string{"dt.entity.service"}
	splitByStatement := []string{"Statement"}

	var filterByNone *dynatrace.DataExplorerFilter = nil
	filterByName := &dynatrace.DataExplorerFilter{
		Filter:          "dt.entity.service",
		FilterType:      "NAME",
		FilterOperator:  "OR",
		EntityAttribute: "entityName",
		Criteria: []dynatrace.DataExplorerCriterion{
			{
				Value:     "EasyTravelWeatherCache",
				Evaluator: "IN",
			},
		},
	}

	filterByID := &dynatrace.DataExplorerFilter{
		Filter:         "dt.entity.service",
		FilterType:     "ID",
		FilterOperator: "OR",
		Criteria: []dynatrace.DataExplorerCriterion{
			{
				Value:     "SERVICE-15366E052E74E22B",
				Evaluator: "IN",
			},
		},
	}

	filterByTag := &dynatrace.DataExplorerFilter{
		Filter:         "dt.entity.service",
		FilterType:     "TAG",
		FilterOperator: "OR",
		Criteria: []dynatrace.DataExplorerCriterion{
			{
				Value:     "frontend",
				Evaluator: "IN",
			},
		},
	}

	filterByEntityAttribute := &dynatrace.DataExplorerFilter{
		Filter:          "dt.entity.service",
		FilterType:      "ENTITY_ATTRIBUTE",
		FilterOperator:  "OR",
		EntityAttribute: "databaseName",
		Criteria: []dynatrace.DataExplorerCriterion{
			{
				Value:     "EasyTravelWeatherCache",
				Evaluator: "IN",
			},
		},
	}

	filterByDimension := &dynatrace.DataExplorerFilter{
		Filter:         "Statement",
		FilterType:     "DIMENSION",
		FilterOperator: "OR",
		Criteria: []dynatrace.DataExplorerCriterion{
			{
				Value:     "Writes in weathercacheentries",
				Evaluator: "EQ",
			},
		},
	}

	type sliNameAndValue struct {
		name  string
		value float64
	}

	tests := []struct {
		name                   string
		queryResultsFilename   string
		splitBy                []string
		filterby               *dynatrace.DataExplorerFilter
		expectedMetricsRequest string
		sliNamesAndValues      []sliNameAndValue
	}{
		{
			name:                   "no splitby, no filterby",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_no_splitby_no_filterby.json",
			splitBy:                splitByNone,
			filterby:               filterByNone,
			expectedMetricsRequest: buildMetricsV2RequestString("calc%3Aservice.dbcalls%3AsplitBy%28%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{{
				name:  "db_calls",
				value: 1867.049541046755,
			}},
		},
		{
			name:                   "splitby service, no filterby",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_service_no_filterby.json",
			splitBy:                splitByService,
			filterby:               filterByNone,
			expectedMetricsRequest: buildMetricsV2RequestString("calc%3Aservice.dbcalls%3AsplitBy%28%22dt.entity.service%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls_service_port:_1234",
					value: 3166.5399830501183,
				},
				{
					name:  "db_calls_service_port:_3306",
					value: 1125.1679389312976,
				},
			},
		},
		{
			name:                   "splitby statement, no filterby",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_statement_no_filterby.json",
			splitBy:                splitByStatement,
			filterby:               filterByNone,
			expectedMetricsRequest: buildMetricsV2RequestString("calc%3Aservice.dbcalls%3AsplitBy%28%22Statement%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls_aggregations_in_bookingcollection",
					value: 9.46440931624731,
				},
				{
					name:  "db_calls_aggregations_in_journeycollection",
					value: 9.532175723785853,
				},
			},
		},

		{
			name:                   "no splitby, filterby name",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_no_splitby_filterby_name.json",
			splitBy:                splitByNone,
			filterby:               filterByName,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CentityName%28%22EasyTravelWeatherCache%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{{
				name:  "db_calls",
				value: 1941.0758136433176,
			}},
		},
		{
			name:                   "splitby service, filterby name",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_service_filterby_name.json",
			splitBy:                splitByService,
			filterby:               filterByName,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CentityName%28%22EasyTravelWeatherCache%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22dt.entity.service%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls",
					value: 1941.034179538489,
				},
			},
		},
		{
			name:                   "splitby statement, filterby name",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_statement_filterby_name.json",
			splitBy:                splitByStatement,
			filterby:               filterByName,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CentityName%28%22EasyTravelWeatherCache%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22Statement%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls_writes_in_weathercacheentries",
					value: 2035.7494656150955,
				},
				{
					name:  "db_calls_reads_in_weathercacheentries",
					value: 1920.5535757845726,
				},
			},
		},

		{
			name:                   "no splitby, filterby ID",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_no_splitby_filterby_id.json",
			splitBy:                splitByNone,
			filterby:               filterByID,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("entityId%28SERVICE-15366E052E74E22B%29", "calc%3Aservice.dbcalls%3AsplitBy%28%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{{
				name:  "db_calls",
				value: 3166.6748652161255,
			}},
		},
		{
			name:                   "splitby service, filterby ID",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_service_filterby_id.json",
			splitBy:                splitByService,
			filterby:               filterByID,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("entityId%28SERVICE-15366E052E74E22B%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22dt.entity.service%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls",
					value: 3166.674865603506,
				},
			},
		},
		{
			name:                   "splitby statement, filterby ID",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_statement_filterby_id.json",
			splitBy:                splitByStatement,
			filterby:               filterByID,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("entityId%28SERVICE-15366E052E74E22B%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22Statement%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls_sql_commit",
					value: 2000,
				},
				{
					name:  "db_calls_insert_into_table1",
					value: 1500,
				},
				{
					name:  "db_calls_insert_into_table2",
					value: 1500,
				},
			},
		},

		{
			name:                   "no splitby, filterby tag",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_no_splitby_filterby_tag.json",
			splitBy:                splitByNone,
			filterby:               filterByTag,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22frontend%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{{
				name:  "db_calls",
				value: 2023.748,
			}},
		},
		{
			name:                   "splitby service, filterby tag",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_service_filterby_tag.json",
			splitBy:                splitByService,
			filterby:               filterByTag,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22frontend%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22dt.entity.service%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls",
					value: 1136.486560350667,
				},
			},
		},
		{
			name:                   "splitby statement, filterby tag",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_statement_filterby_tag.json",
			splitBy:                splitByStatement,
			filterby:               filterByTag,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22frontend%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22Statement%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls_insert_into_table1",
					value: 1200,
				},
				{
					name:  "db_calls_insert_into_table2",
					value: 1200,
				},
			},
		},

		{
			name:                   "no splitby, filterby entity attribute",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_no_splitby_filterby_entity_attribute.json",
			splitBy:                splitByNone,
			filterby:               filterByEntityAttribute,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CdatabaseName%28%22EasyTravelWeatherCache%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{{
				name:  "db_calls",
				value: 1321.804989244147,
			}},
		},
		{
			name:                   "splitby service, filterby entity attribute",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_service_filterby_entity_attribute.json",
			splitBy:                splitByService,
			filterby:               filterByEntityAttribute,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CdatabaseName%28%22EasyTravelWeatherCache%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22dt.entity.service%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls_easytravelweathercache",
					value: 1940.9776856154995,
				},
				{
					name:  "db_calls_service_port:_27017",
					value: 1312.8391313736909,
				},
			},
		},
		{
			name:                   "splitby statement, filterby entity attribute",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_statement_filterby_entity_attribute.json",
			splitBy:                splitByStatement,
			filterby:               filterByEntityAttribute,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CdatabaseName%28%22EasyTravelWeatherCache%22%29", "calc%3Aservice.dbcalls%3AsplitBy%28%22Statement%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls_writes_in_weathercacheentries",
					value: 1387.1373288989516,
				},
				{
					name:  "db_calls_writes",
					value: 1232.7777777777778,
				},
				{
					name:  "db_calls_reads_in_weathercacheentries",
					value: 1301.3870571893435,
				},
			},
		},

		{
			name:                   "no splitby, filterby dimension",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_no_splitby_filterby_dimension.json",
			splitBy:                splitByNone,
			filterby:               filterByDimension,
			expectedMetricsRequest: buildMetricsV2RequestString("calc%3Aservice.dbcalls%3Afilter%28EQ%28%22Statement%22%2C%22Writes+in+weathercacheentries%22%29%29%3AsplitBy%28%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{{
				name:  "db_calls",
				value: 2113.049892441478,
			}},
		},
		{
			name:                   "splitby service, filterby dimension",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_service_filterby_dimension.json",
			splitBy:                splitByService,
			filterby:               filterByDimension,
			expectedMetricsRequest: buildMetricsV2RequestString("calc%3Aservice.dbcalls%3Afilter%28EQ%28%22Statement%22%2C%22Writes+in+weathercacheentries%22%29%29%3AsplitBy%28%22dt.entity.service%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls",
					value: 1941.9776856154995,
				},
			},
		},
		{
			name:                   "splitby statement, filterby entity dimension",
			queryResultsFilename:   "metrics_query_calc_service_dbcalls_splitby_statement_filterby_dimension.json",
			splitBy:                splitByStatement,
			filterby:               filterByDimension,
			expectedMetricsRequest: buildMetricsV2RequestString("calc%3Aservice.dbcalls%3Afilter%28EQ%28%22Statement%22%2C%22Writes+in+weathercacheentries%22%29%29%3AsplitBy%28%22Statement%22%29%3Aauto%3Anames"),
			sliNamesAndValues: []sliNameAndValue{
				{
					name:  "db_calls",
					value: 1387.1373288989516,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				testDataFolder+"dashboard.template.json",
				struct {
					SplitByString  string
					FilterByString string
				}{
					SplitByString:  convertToJSONString(t, tt.splitBy),
					FilterByString: convertToJSONStringOrEmptyIfNil(t, tt.filterby),
				},
			)
			handler.AddExactFile(dynatrace.MetricsPath+"/calc:service.dbcalls", testDataFolder+"metrics_calc_service_dbcalls.json")
			handler.AddExactFile(tt.expectedMetricsRequest, testDataFolder+tt.queryResultsFilename)

			sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){}
			for _, s := range tt.sliNamesAndValues {
				sliResultsAssertionsFuncs = append(sliResultsAssertionsFuncs, createSuccessfulSLIResultAssertionsFunc(s.name, s.value, tt.expectedMetricsRequest))
			}

			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, sliResultsAssertionsFuncs...)
		})
	}
}

// TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAggregationWorks tests applying a space aggregation to the Data Explorer tile works as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTile_SpaceAggregationWorks(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/space_aggregation_works/"

	tests := []struct {
		name                          string
		spaceAggregation              string
		metricsQueryResultAggregation string
		expectedMetricsRequest        string
	}{
		{
			name:                          "auto",
			spaceAggregation:              "",
			metricsQueryResultAggregation: "auto",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Aauto%3Anames"),
		},
		{
			name:                          "average",
			spaceAggregation:              "AVG",
			metricsQueryResultAggregation: "avg",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                          "count",
			spaceAggregation:              "COUNT",
			metricsQueryResultAggregation: "count",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Acount%3Anames"),
		},
		{
			name:                          "maximum",
			spaceAggregation:              "MAX",
			metricsQueryResultAggregation: "max",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Amax%3Anames"),
		},
		{
			name:                          "minimum",
			spaceAggregation:              "MIN",
			metricsQueryResultAggregation: "min",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Amin%3Anames"),
		},
		{
			name:                          "sum",
			spaceAggregation:              "SUM",
			metricsQueryResultAggregation: "sum",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Asum%3Anames"),
		},
		{
			name:                          "median",
			spaceAggregation:              "MEDIAN",
			metricsQueryResultAggregation: "median",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Amedian%3Anames"),
		},
		{
			name:                          "value",
			spaceAggregation:              "VALUE",
			metricsQueryResultAggregation: "value",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Avalue%3Anames"),
		},
		{
			name:                          "percentile(10)",
			spaceAggregation:              "PERCENTILE_10",
			metricsQueryResultAggregation: "percentile(10)",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2810%29%3Anames"),
		},
		{
			name:                          "percentile(75)",
			spaceAggregation:              "PERCENTILE_75",
			metricsQueryResultAggregation: "percentile(75)",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2875%29%3Anames"),
		},
		{
			name:                          "percentile(90)",
			spaceAggregation:              "PERCENTILE_90",
			metricsQueryResultAggregation: "percentile(90)",
			expectedMetricsRequest:        buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Apercentile%2890%29%3Anames"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createHandlerWithTemplatedDashboard(t,
				testDataFolder+"dashboard.template.json",
				struct {
					SpaceAggregation string
				}{
					SpaceAggregation: tt.spaceAggregation,
				},
			)
			handler.AddExactFile(dynatrace.MetricsPath+"/builtin:service.response.time", testDataFolder+"metrics_builtin_service_response_time.json")
			handler.AddExactTemplate(tt.expectedMetricsRequest, testDataFolder+"metrics_query_builtin_service_response_time.template.json",
				struct {
					Aggregation string
				}{
					Aggregation: tt.metricsQueryResultAggregation,
				})

			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc("srt", 29192.929640271974, tt.expectedMetricsRequest))
		})
	}
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

// TestRetrieveMetricsFromDashboardDataExplorerTile_ManagementZonesWork tests applying management zones to the dashboard and tile work as expected, also when combined with a filter that appears on the entity selector.
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

	emptyQueryFilter := dynatrace.DataExplorerFilter{
		NestedFilters: []dynatrace.DataExplorerFilter{},
		Criteria:      []dynatrace.DataExplorerCriterion{},
	}

	queryFilterWithTag := dynatrace.DataExplorerFilter{
		FilterOperator: "AND",
		NestedFilters: []dynatrace.DataExplorerFilter{
			{
				Filter:         "dt.entity.service",
				FilterType:     "TAG",
				FilterOperator: "OR",
				Criteria: []dynatrace.DataExplorerCriterion{
					{
						Value:     "service_tag",
						Evaluator: "in",
					},
				},
			},
		},
	}

	tests := []struct {
		name                   string
		dashboardFilter        *dynatrace.DashboardFilter
		tileFilter             dynatrace.TileFilter
		queryFilter            dynatrace.DataExplorerFilter
		expectedMetricsRequest string
	}{
		{
			name:                   "no dashboard filter, empty tile filter, empty query filter",
			dashboardFilter:        nil,
			tileFilter:             emptyTileFilter,
			queryFilter:            emptyQueryFilter,
			expectedMetricsRequest: buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                   "dashboard filter with mz, empty tile filter, empty query filter",
			dashboardFilter:        &dashboardFilterWithManagementZone,
			tileFilter:             emptyTileFilter,
			queryFilter:            emptyQueryFilter,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CmzId%28-1234567890123456789%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                   "no dashboard filter, tile filter with mz, empty query filter",
			dashboardFilter:        nil,
			tileFilter:             tileFilterWithManagementZone,
			queryFilter:            emptyQueryFilter,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CmzId%282311420533206603714%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                   "dashboard filter with mz, tile filter with mz, empty query filter",
			dashboardFilter:        &dashboardFilterWithManagementZone,
			tileFilter:             tileFilterWithManagementZone,
			queryFilter:            emptyQueryFilter,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2CmzId%282311420533206603714%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                   "no dashboard filter, empty tile filter, query filter with tag",
			dashboardFilter:        nil,
			tileFilter:             emptyTileFilter,
			queryFilter:            queryFilterWithTag,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22service_tag%22%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                   "dashboard filter with mz, empty tile filter, query filter with tag",
			dashboardFilter:        &dashboardFilterWithManagementZone,
			tileFilter:             emptyTileFilter,
			queryFilter:            queryFilterWithTag,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22service_tag%22%29%2CmzId%28-1234567890123456789%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                   "no dashboard filter, tile filter with mz, query filter with tag",
			dashboardFilter:        nil,
			tileFilter:             tileFilterWithManagementZone,
			queryFilter:            queryFilterWithTag,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22service_tag%22%29%2CmzId%282311420533206603714%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
		{
			name:                   "dashboard filter with mz, tile filter with mz, query filter with tag",
			dashboardFilter:        &dashboardFilterWithManagementZone,
			tileFilter:             tileFilterWithManagementZone,
			queryFilter:            queryFilterWithTag,
			expectedMetricsRequest: buildMetricsV2RequestStringWithEntitySelector("type%28SERVICE%29%2Ctag%28%22service_tag%22%29%2CmzId%282311420533206603714%29", "builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				testDataFolder+"dashboard.template.json",
				struct {
					DashboardFilterString string
					TileFilterString      string
					QueryFilterString     string
				}{
					DashboardFilterString: convertToJSONStringOrEmptyIfNil(t, tt.dashboardFilter),
					TileFilterString:      convertToJSONString(t, tt.tileFilter),
					QueryFilterString:     convertToJSONString(t, tt.queryFilter),
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

	expectedMetricsRequest := buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames")

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

	expectedMetricsRequest := buildMetricsV2RequestStringWithEntitySelector("entityId%28SERVICE-B67B3EC4C95E0FA7%29", "builtin%3Aservice.response.time%3AsplitBy%28%22dt.entity.service%22%29%3Aavg%3Anames")

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
// TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdsWork tests that setting pass and warning criteria via thresholds on the tile works as expected.
func TestRetrieveMetricsFromDashboardDataExplorerTile_TileThresholdsWork(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/data_explorer/tile_thresholds_success/"

	expectedMetricsRequest := buildMetricsV2RequestString("builtin%3Aservice.response.time%3AsplitBy%28%29%3Aavg%3Anames")

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

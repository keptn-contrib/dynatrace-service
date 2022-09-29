package sli

import (
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

const testIndicatorUSQL = "usql_sli"

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but the USQL prefix is used incorrectly, so we return an error for that
func TestCustomSLIWithIncorrectUSQLQueryPrefix(t *testing.T) {

	testConfigs := []struct {
		name       string
		usqlPrefix string
	}{
		{
			name:       "missing column fails",
			usqlPrefix: "USQL;COLUMN_CHART;",
		},
		{
			name:       "3 missing fields fails",
			usqlPrefix: "USQL;",
		},
		{
			name:       "2 missing fields fails",
			usqlPrefix: "USQL;;",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			// no handler needed
			handler := test.NewFileBasedURLHandler(t)

			// error here: in value of tc.usqlPrefix
			configClient := newConfigClientMockWithSLIs(t, map[string]string{
				testIndicatorUSQL: tc.usqlPrefix + "SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			})

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorUSQL, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "incorrect prefix"))
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
//   - a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
//   - the defined SLI is valid YAML and the USQL prefix is used correctly, but the fields are used incorrectly.
//     So we return an error for that
func TestCustomSLIWithCorrectUSQLQueryPrefixMappings(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/usql/processing_errors"

	expectedUSQLRequest := buildUSQLRequest("SELECT+osVersion%2CAVG%28duration%29+FROM+usersession+GROUP+BY+osVersion")
	testConfigs := []struct {
		name                              string
		usqlPrefix                        string
		expectedErrorMessage              string
		getSLIFinishedEventAssertionsFunc func(*testing.T, *getSLIFinishedEventData)
		sliResultAssertionsFunc           func(*testing.T, sliResult)
	}{
		{
			name:                              "unknown type fails",
			usqlPrefix:                        "USQL;COLUMN_CHARTS;iOS 11.4.1;",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "unknown result type: COLUMN_CHARTS"),
		},
		{
			name:                              "unknown dimension name fails",
			usqlPrefix:                        "USQL;COLUMN_CHART;iOS 17.2.3;",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, expectedUSQLRequest, "could not find dimension name 'iOS 17.2.3'"),
		},
		{
			name:                              "missing fields fails",
			usqlPrefix:                        "USQL;;;",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "result type should not be empty"),
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			// handler with 200 result needed
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_200_multiple_results.json"))

			// errors here: in value of tc.usqlPrefix
			configClient := newConfigClientMockWithSLIs(t, map[string]string{
				testIndicatorUSQL: tc.usqlPrefix + "SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			})

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorUSQL, tc.getSLIFinishedEventAssertionsFunc, tc.sliResultAssertionsFunc)
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomUSQLQueriesReturnsMultipleResults(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/usql/multiple_result_processing"

	expectedUSQLRequest := buildUSQLRequest("SELECT+osVersion%2CAVG%28duration%29%2CMAX%28duration%29+FROM+usersession+GROUP+BY+osVersion")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_200_multiple_results.json"))

	testConfigs := []struct {
		name          string
		query         string
		expectedValue float64
	}{
		{
			name:          "column chart",
			query:         "USQL;COLUMN_CHART;Android 6.0.1;SELECT osVersion,AVG(duration),MAX(duration) FROM usersession GROUP BY osVersion",
			expectedValue: 21862.42,
		},
		{
			name:          "line chart",
			query:         "USQL;LINE_CHART;Android 7.0.1;SELECT osVersion,AVG(duration),MAX(duration) FROM usersession GROUP BY osVersion",
			expectedValue: 26304,
		},
		{
			name:          "pie chart",
			query:         "USQL;PIE_CHART;iOS 11.4.1;SELECT osVersion,AVG(duration),MAX(duration) FROM usersession GROUP BY osVersion",
			expectedValue: 23576,
		},
		{
			name:          "table",
			query:         "USQL;TABLE;iOS 12.1.4;SELECT osVersion,AVG(duration),MAX(duration) FROM usersession GROUP BY osVersion",
			expectedValue: 24824,
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			configClient := newConfigClientMockWithSLIs(t, map[string]string{
				testIndicatorUSQL: tc.query,
			})

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorUSQL, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorUSQL, tc.expectedValue, expectedUSQLRequest))
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomUSQLQueriesReturnsSingleResults(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/usql/single_result_processing"

	expectedUSQLRequest := buildUSQLRequest("SELECT+AVG%28duration%29+FROM+usersession")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_200_single_result.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorUSQL: "USQL;SINGLE_VALUE;;SELECT AVG(duration) FROM usersession",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorUSQL, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorUSQL, 62737.44360695537, expectedUSQLRequest))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomUSQLQueriesReturnsNoResults(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/usql/no_results"

	expectedUSQLRequest := buildUSQLRequest("SELECT+osVersion%2CAVG%28duration%29+FROM+usersession+GROUP+BY+osVersion")

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(expectedUSQLRequest, filepath.Join(testDataFolder, "usql_200_0_results.json"))

	configClient := newConfigClientMockWithSLIs(t, map[string]string{
		testIndicatorUSQL: "USQL;COLUMN_CHART;iOS 11.4.1;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
	})

	runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorUSQL, getSLIFinishedEventWarningAssertionsFunc, createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, expectedUSQLRequest, "could not find dimension name"))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but the fields of the USQL prefix are used incorrectly together, so we return errors for that
func TestCustomSLIWithIncorrectUSQLConfiguration(t *testing.T) {
	const testDataFolder = "./testdata/sli_files/usql/incorrect_configuration"
	usqlSingleResultRequest := buildUSQLRequest("SELECT+AVG%28duration%29+FROM+usersession")
	usqlMultipleResultRequest1 := buildUSQLRequest("SELECT+AVG%28duration%29%2CosVersion+FROM+usersession+GROUP+BY+osVersion")
	usqlMultipleResultRequest2 := buildUSQLRequest("SELECT+osVersion%2CosVersion%2CAVG%28duration%29+FROM+usersession+GROUP+BY+osVersion")
	usqlMultipleResultRequest3 := buildUSQLRequest("SELECT+osVersion%2CAVG%28duration%29%2CosVersion+FROM+usersession+GROUP+BY+osVersion")

	testConfigs := []struct {
		name                              string
		request                           string
		usqlQuery                         string
		dataReturned                      string
		expectedErrorMessage              string
		getSLIFinishedEventAssertionsFunc func(t *testing.T, data *getSLIFinishedEventData)
		sliResultAssertionsFunc           func(*testing.T, sliResult)
	}{
		{
			name:                              "dimension name is not allowed for single value result type",
			usqlQuery:                         "USQL;SINGLE_VALUE;iOS 11.4.1;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "dimension should be empty"),
		},
		{
			name:                              "dimension name should not be empty for COLUMN_CHART result types",
			usqlQuery:                         "USQL;COLUMN_CHART;;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "dimension should not be empty"),
		},
		{
			name:                              "dimension name should not be empty for PIE_CHART result types",
			usqlQuery:                         "USQL;PIE_CHART;;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "dimension should not be empty"),
		},
		{
			name:                              "dimension name should not be empty for TABLE result types",
			usqlQuery:                         "USQL;TABLE;;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "dimension should not be empty"),
		},
		{
			name:                              "dimension name should not be empty for COLUMN_CHART result types even if result only has single value",
			usqlQuery:                         "USQL;COLUMN_CHART;;SELECT AVG(duration) FROM usersession",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_single_result.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "dimension should not be empty"),
		},
		{
			name:                              "dimension name should not be empty for PIE_CHART result types even if result only has single value",
			usqlQuery:                         "USQL;PIE_CHART;;SELECT AVG(duration) FROM usersession",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_single_result.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "dimension should not be empty"),
		},
		{
			name:                              "dimension name should not be empty for TABLE result types even if result only has single value",
			usqlQuery:                         "USQL;TABLE;;SELECT AVG(duration) FROM usersession",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_single_result.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultAssertionsFunc(testIndicatorUSQL, "dimension should not be empty"),
		},
		{
			name:                              "COLUMN_CHART should have at least two columns",
			request:                           usqlSingleResultRequest,
			usqlQuery:                         "USQL;COLUMN_CHART;iOS 11.4.1;SELECT AVG(duration) FROM usersession",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_single_result.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlSingleResultRequest, "should at least have two columns"),
		},
		{
			name:                              "PIE_CHART should have at least two columns",
			request:                           usqlSingleResultRequest,
			usqlQuery:                         "USQL;PIE_CHART;iOS 11.4.1;SELECT AVG(duration) FROM usersession",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_single_result.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlSingleResultRequest, "should at least have two columns"),
		},
		{
			name:                              "TABLE should have at least two columns",
			request:                           usqlSingleResultRequest,
			usqlQuery:                         "USQL;TABLE;iOS 11.4.1;SELECT AVG(duration) FROM usersession",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_single_result.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlSingleResultRequest, "should at least have two columns"),
		},
		{
			name:                              "result has more than one column, but first column is not a string value for COLUMN_CHART",
			request:                           usqlMultipleResultRequest1,
			usqlQuery:                         "USQL;COLUMN_CHART;iOS 11.4.1;SELECT AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results_wrong_first_column_type.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlMultipleResultRequest1, "dimension name should be a string"),
		},

		{
			name:                              "result has more than one column, but first column is not a string value for PIE_CHART",
			request:                           usqlMultipleResultRequest1,
			usqlQuery:                         "USQL;PIE_CHART;iOS 11.4.1;SELECT AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results_wrong_first_column_type.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlMultipleResultRequest1, "dimension name should be a string"),
		},

		{
			name:                              "result has more than one column, but first column is not a string value for TABLE",
			request:                           usqlMultipleResultRequest1,
			usqlQuery:                         "USQL;TABLE;iOS 11.4.1;SELECT AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results_wrong_first_column_type.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlMultipleResultRequest1, "dimension name should be a string"),
		},
		{
			name:                              "result has more than one column, but second column is not a numeric value for COLUMN_CHART",
			request:                           usqlMultipleResultRequest2,
			usqlQuery:                         "USQL;COLUMN_CHART;iOS 11.4.1;SELECT osVersion,osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results_wrong_second_column_type.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlMultipleResultRequest2, "dimension value should be a number"),
		},
		{
			name:                              "result has more than one column, but second column is not a numeric value for PIE_CHART",
			request:                           usqlMultipleResultRequest2,
			usqlQuery:                         "USQL;PIE_CHART;iOS 11.4.1;SELECT osVersion,osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results_wrong_second_column_type.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlMultipleResultRequest2, "dimension value should be a number"),
		},
		{
			name:                              "result has more than one column, but last column is not a numeric value for TABLE",
			request:                           usqlMultipleResultRequest3,
			usqlQuery:                         "USQL;TABLE;iOS 11.4.1;SELECT osVersion,AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      filepath.Join(testDataFolder, "usql_200_multiple_results_wrong_last_column_type.json"),
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
			sliResultAssertionsFunc:           createFailedSLIResultWithQueryAssertionsFunc(testIndicatorUSQL, usqlMultipleResultRequest3, "dimension value should be a number"),
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			// as there is only a single SLI, matching with 'starts with' should be sufficiently 'exact'
			handler := test.NewFileBasedURLHandler(t)
			if tc.request != "" {
				handler.AddStartsWith(tc.request, tc.dataReturned)
			}

			configClient := newConfigClientMockWithSLIs(t, map[string]string{
				testIndicatorUSQL: tc.usqlQuery,
			})

			runGetSLIsFromFilesTestWithOneIndicatorRequestedAndCheckSLIs(t, handler, configClient, testIndicatorUSQL, tc.getSLIFinishedEventAssertionsFunc, tc.sliResultAssertionsFunc)
		})
	}
}

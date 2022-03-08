package sli

import (
	"testing"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

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
			kClient := &keptnClientMock{
				customQueries: map[string]string{
					indicator: tc.usqlPrefix + "SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
				},
			}

			sliResultAssertionsFunc := func(t *testing.T, sliResult *keptnv2.SLIResult) {
				assert.EqualValues(t, indicator, sliResult.Metric)
				assert.EqualValues(t, 0, sliResult.Value)
				assert.EqualValues(t, false, sliResult.Success)
				assert.Contains(t, sliResult.Message, "incorrect prefix")
			}

			assertThatCustomSLITestIsCorrect(t, handler, kClient, getSLIFinishedEventFailureAssertionsFunc, sliResultAssertionsFunc)
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML and the USQL prefix is used correctly, but the fields are used incorrectly.
//   So we return an error for that
func TestCustomSLIWithCorrectUSQLQueryPrefixMappings(t *testing.T) {

	testConfigs := []struct {
		name                              string
		usqlPrefix                        string
		expectedErrorMessage              string
		getSLIFinishedEventAssertionsFunc func(t *testing.T, data *keptnv2.GetSLIFinishedEventData)
	}{
		{
			name:                              "unknown type fails",
			usqlPrefix:                        "USQL;COLUMN_CHARTS;iOS 11.4.1;",
			expectedErrorMessage:              "unknown result type: COLUMN_CHARTS",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "unknown dimension name fails",
			usqlPrefix:                        "USQL;COLUMN_CHART;iOS 17.2.3;",
			expectedErrorMessage:              "could not find dimension name 'iOS 17.2.3'",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "missing fields fails",
			usqlPrefix:                        "USQL;;;",
			expectedErrorMessage:              "result type should not be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			// handler with 200 result needed
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(
				dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1632835299000&explain=false&query=SELECT+osVersion%2CAVG%28duration%29+FROM+usersession+GROUP+BY+osVersion&startTimestamp=1632834999000",
				"./testdata/usql_200_multiple_results.json")

			// errors here: in value of tc.usqlPrefix
			kClient := &keptnClientMock{
				customQueries: map[string]string{
					indicator: tc.usqlPrefix + "SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
				},
			}

			sliResultAssertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
				assert.EqualValues(t, indicator, actual.Metric)
				assert.EqualValues(t, 0, actual.Value)
				assert.EqualValues(t, false, actual.Success)
				assert.Contains(t, actual.Message, tc.expectedErrorMessage)
			}

			assertThatCustomSLITestIsCorrect(t, handler, kClient, tc.getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomUSQLQueriesReturnsMultipleResults(t *testing.T) {

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1632835299000&explain=false&query=SELECT+osVersion%2CAVG%28duration%29%2CMAX%28duration%29+FROM+usersession+GROUP+BY+osVersion&startTimestamp=1632834999000",
		"./testdata/usql_200_multiple_results.json")

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

			kClient := &keptnClientMock{
				customQueries: map[string]string{
					indicator: tc.query,
				},
			}

			assertThatCustomSLITestIsCorrect(t, handler, kClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(indicator, tc.expectedValue))
		})
	}
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomUSQLQueriesReturnsSingleResults(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1632835299000&explain=false&query=SELECT+AVG%28duration%29+FROM+usersession&startTimestamp=1632834999000",
		"./testdata/usql_200_single_result.json")

	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "USQL;SINGLE_VALUE;;SELECT AVG(duration) FROM usersession",
		},
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(indicator, 62737.44360695537))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
func TestCustomUSQLQueriesReturnsNoResults(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(
		dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1632835299000&explain=false&query=SELECT+osVersion%2CAVG%28duration%29+FROM+usersession+GROUP+BY+osVersion&startTimestamp=1632834999000",
		"./testdata/usql_200_0_results.json")

	kClient := &keptnClientMock{
		customQueries: map[string]string{
			indicator: "USQL;COLUMN_CHART;iOS 11.4.1;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
		},
	}

	sliResultAssertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
		assert.EqualValues(t, indicator, actual.Metric)
		assert.EqualValues(t, 0, actual.Value)
		assert.EqualValues(t, false, actual.Success)
		assert.Contains(t, actual.Message, "could not find dimension name")
	}

	assertThatCustomSLITestIsCorrect(t, handler, kClient, getSLIFinishedEventWarningAssertionsFunc, sliResultAssertionsFunc)
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but the fields of the USQL prefix are used incorrectly together, so we return errors for that
func TestCustomSLIWithIncorrectUSQLConfiguration(t *testing.T) {

	testConfigs := []struct {
		name                              string
		usqlQuery                         string
		dataReturned                      string
		expectedErrorMessage              string
		getSLIFinishedEventAssertionsFunc func(t *testing.T, data *keptnv2.GetSLIFinishedEventData)
	}{
		{
			name:                              "dimension name is not allowed for single value result type",
			usqlQuery:                         "USQL;SINGLE_VALUE;iOS 11.4.1;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results.json",
			expectedErrorMessage:              "dimension should be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "dimension name should not be empty for COLUMN_CHART result types",
			usqlQuery:                         "USQL;COLUMN_CHART;;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results.json",
			expectedErrorMessage:              "dimension should not be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "dimension name should not be empty for PIE_CHART result types",
			usqlQuery:                         "USQL;PIE_CHART;;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results.json",
			expectedErrorMessage:              "dimension should not be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "dimension name should not be empty for TABLE result types",
			usqlQuery:                         "USQL;TABLE;;SELECT osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results.json",
			expectedErrorMessage:              "dimension should not be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "dimension name should not be empty for COLUMN_CHART result types even if result only has single value",
			usqlQuery:                         "USQL;COLUMN_CHART;;SELECT AVG(duration) FROM usersession",
			dataReturned:                      "./testdata/usql_200_single_result.json",
			expectedErrorMessage:              "dimension should not be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "dimension name should not be empty for PIE_CHART result types even if result only has single value",
			usqlQuery:                         "USQL;PIE_CHART;;SELECT AVG(duration) FROM usersession",
			dataReturned:                      "./testdata/usql_200_single_result.json",
			expectedErrorMessage:              "dimension should not be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "dimension name should not be empty for TABLE result types even if result only has single value",
			usqlQuery:                         "USQL;TABLE;;SELECT AVG(duration) FROM usersession",
			dataReturned:                      "./testdata/usql_200_single_result.json",
			expectedErrorMessage:              "dimension should not be empty",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
		},
		{
			name:                              "COLUMN_CHART should have at least two columns",
			usqlQuery:                         "USQL;COLUMN_CHART;iOS 11.4.1;SELECT AVG(duration) FROM usersession",
			dataReturned:                      "./testdata/usql_200_single_result.json",
			expectedErrorMessage:              "should at least have two columns",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "PIE_CHART should have at least two columns",
			usqlQuery:                         "USQL;PIE_CHART;iOS 11.4.1;SELECT AVG(duration) FROM usersession",
			dataReturned:                      "./testdata/usql_200_single_result.json",
			expectedErrorMessage:              "should at least have two columns",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "TABLE should have at least two columns",
			usqlQuery:                         "USQL;TABLE;iOS 11.4.1;SELECT AVG(duration) FROM usersession",
			dataReturned:                      "./testdata/usql_200_single_result.json",
			expectedErrorMessage:              "should at least have two columns",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "result has more than one column, but first column is not a string value for COLUMN_CHART",
			usqlQuery:                         "USQL;COLUMN_CHART;iOS 11.4.1;SELECT AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results_wrong_first_column_type.json",
			expectedErrorMessage:              "dimension name should be a string",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "result has more than one column, but first column is not a string value for PIE_CHART",
			usqlQuery:                         "USQL;PIE_CHART;iOS 11.4.1;SELECT AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results_wrong_first_column_type.json",
			expectedErrorMessage:              "dimension name should be a string",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "result has more than one column, but first column is not a string value for TABLE",
			usqlQuery:                         "USQL;TABLE;iOS 11.4.1;SELECT AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results_wrong_first_column_type.json",
			expectedErrorMessage:              "dimension name should be a string",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "result has more than one column, but second column is not a numeric value for COLUMN_CHART",
			usqlQuery:                         "USQL;COLUMN_CHART;iOS 11.4.1;SELECT osVersion,osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results_wrong_second_column_type.json",
			expectedErrorMessage:              "dimension value should be a number",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "result has more than one column, but second column is not a numeric value for PIE_CHART",
			usqlQuery:                         "USQL;PIE_CHART;iOS 11.4.1;SELECT osVersion,osVersion,AVG(duration) FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results_wrong_second_column_type.json",
			expectedErrorMessage:              "dimension value should be a number",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
		{
			name:                              "result has more than one column, but last column is not a numeric value for TABLE",
			usqlQuery:                         "USQL;TABLE;iOS 11.4.1;SELECT osVersion,AVG(duration),osVersion FROM usersession GROUP BY osVersion",
			dataReturned:                      "./testdata/usql_200_multiple_results_wrong_last_column_type.json",
			expectedErrorMessage:              "dimension value should be a number",
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventWarningAssertionsFunc,
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {

			// as there is only a single SLI, matching with 'starts with' should be sufficiently 'exact'
			handler := test.NewFileBasedURLHandler(t)
			handler.AddStartsWith(
				dynatrace.USQLPath+"?addDeepLinkFields=false&endTimestamp=1632835299000&explain=false&query=",
				tc.dataReturned)

			kClient := &keptnClientMock{
				customQueries: map[string]string{
					indicator: tc.usqlQuery,
				},
			}

			sliResultAssertionsFunc := func(t *testing.T, actual *keptnv2.SLIResult) {
				assert.EqualValues(t, indicator, actual.Metric)
				assert.EqualValues(t, 0, actual.Value)
				assert.EqualValues(t, false, actual.Success)
				assert.Contains(t, actual.Message, tc.expectedErrorMessage)
			}

			assertThatCustomSLITestIsCorrect(t, handler, kClient, tc.getSLIFinishedEventAssertionsFunc, sliResultAssertionsFunc)
		})
	}
}

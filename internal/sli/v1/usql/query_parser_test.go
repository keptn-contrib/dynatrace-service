package usql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryParser tests the QueryParser
func TestQueryParser(t *testing.T) {
	tests := []struct {
		name                 string
		inputQuery           string
		expectedResultType   string
		expectedDimension    string
		expectedQuery        string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:               "valid",
			inputQuery:         "USQL;COLUMN_CHART;iPad mini;SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
			expectedResultType: "COLUMN_CHART",
			expectedDimension:  "iPad mini",
			expectedQuery:      "SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
		},
		{
			name:               "valid - line chart",
			inputQuery:         "USQL;LINE_CHART;iPad mini;SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
			expectedResultType: "LINE_CHART",
			expectedDimension:  "iPad mini",
			expectedQuery:      "SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
		},
		{
			name:               "valid - extra semi colon in query",
			inputQuery:         "USQL;COLUMN_CHART;iPad mini;SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria;Italy') GROUP BY device",
			expectedResultType: "COLUMN_CHART",
			expectedDimension:  "iPad mini",
			expectedQuery:      "SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria;Italy') GROUP BY device",
		},

		{
			name:               "valid - SINGLE_VALUE with should have no dimension",
			inputQuery:         "USQL;SINGLE_VALUE;;SELECT AVG(duration) FROM usersession WHERE country IN('Austria')",
			expectedResultType: "SINGLE_VALUE",
			expectedDimension:  "",
			expectedQuery:      "SELECT AVG(duration) FROM usersession WHERE country IN('Austria')",
		},
		{
			name:                 "invalid - COLUMN_CHART missing dimension",
			inputQuery:           "USQL;COLUMN_CHART;;SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
			expectError:          true,
			expectedErrorMessage: "dimension should not be empty",
		},
		{
			name:                 "invalid - SINGLE_VALUE with should have no dimension",
			inputQuery:           "USQL;SINGLE_VALUE;dimension;SELECT AVG(duration) FROM usersession WHERE country IN('Austria')",
			expectError:          true,
			expectedErrorMessage: "dimension should be empty",
		},
		{
			name:                 "invalid - unknown result type",
			inputQuery:           "USQL;unknown;dimension;SELECT AVG(duration) FROM usersession WHERE country IN('Austria')",
			expectError:          true,
			expectedErrorMessage: "unknown result type",
		},
		{
			name:                 "invalid - empty result type",
			inputQuery:           "USQL;;dimension;SELECT AVG(duration) FROM usersession WHERE country IN('Austria')",
			expectError:          true,
			expectedErrorMessage: "result type should not be empty",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewQueryParser(tc.inputQuery).Parse()
			if tc.expectError {
				assert.Nil(t, query)
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage)
				}
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, query) {
					assert.EqualValues(t, tc.expectedResultType, query.GetResultType())
					assert.EqualValues(t, tc.expectedDimension, query.GetDimension())
					assert.EqualValues(t, tc.expectedQuery, query.GetQuery().GetQuery())
				}
			}
		})
	}
}

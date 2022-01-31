package usql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewQuery tests the creation of new USQL Query instances.
func TestNewQuery(t *testing.T) {
	tests := []struct {
		name                 string
		inputQuery           string
		expectedQuery        string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:          "with query",
			inputQuery:    "SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
			expectedQuery: "SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
		},
		// Error cases below:
		{
			name:                 "with no query",
			expectError:          true,
			expectedErrorMessage: "USQL query should not be empty",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewQuery(tc.inputQuery)
			if tc.expectError {
				assert.Nil(t, query)
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage)
				}
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, query) {
					assert.EqualValues(t, tc.expectedQuery, query.GetQuery())
				}
			}
		})
	}
}

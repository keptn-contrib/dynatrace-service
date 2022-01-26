package problemsv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryParser tests the QueryParser
func TestQueryParser(t *testing.T) {
	tests := []struct {
		name                    string
		inputQuery              string
		expectedProblemSelector string
		expectedEntitySelector  string
	}{
		{
			name:                    "valid",
			inputQuery:              "PV2;problemSelector=status(open)&entitySelector=mzId(7030365576649815430)",
			expectedProblemSelector: "status(open)",
			expectedEntitySelector:  "mzId(7030365576649815430)",
		},
		{
			name:       "valid - empty",
			inputQuery: "PV2;",
		},
		{
			name:                   "valid",
			inputQuery:             "PV2;entitySelector=mzId(7030365576649815430)",
			expectedEntitySelector: "mzId(7030365576649815430)",
		},
		{
			name:                    "valid",
			inputQuery:              "PV2;problemSelector=status(open)",
			expectedProblemSelector: "status(open)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewQueryParser(tc.inputQuery).Parse()

			assert.NoError(t, err)
			if assert.NotNil(t, query) {
				assert.EqualValues(t, tc.expectedProblemSelector, query.GetProblemSelector())
				assert.EqualValues(t, tc.expectedEntitySelector, query.GetEntitySelector())
			}
		})
	}
}

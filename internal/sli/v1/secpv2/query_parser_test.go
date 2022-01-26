package secpv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryParser tests the QueryParser
func TestQueryParser(t *testing.T) {
	tests := []struct {
		name                            string
		inputQuery                      string
		expectedSecurityProblemSelector string
	}{
		{
			name:                            "valid",
			inputQuery:                      "SECPV2;securityProblemSelector=status(open)",
			expectedSecurityProblemSelector: "status(open)",
		},
		{
			name:       "valid - empty",
			inputQuery: "SECPV2;",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := NewQueryParser(tc.inputQuery).Parse()
			assert.NoError(t, err)
			if assert.NotNil(t, query) {
				assert.EqualValues(t, tc.expectedSecurityProblemSelector, query.GetSecurityProblemSelector())
			}
		})
	}
}

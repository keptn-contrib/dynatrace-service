package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryParser tests the QueryParser
func TestQueryParser(t *testing.T) {
	tests := []struct {
		name                 string
		inputQuery           string
		expectedSLOID        string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:          "valid",
			inputQuery:    "SLO;524ca177-849b-3e8c-8175-42b93fbc33c5",
			expectedSLOID: "524ca177-849b-3e8c-8175-42b93fbc33c5",
		},
		{
			name:                 "invalid - no SLO prefix",
			inputQuery:           ";524ca177-849b-3e8c-8175-42b93fbc33c5",
			expectError:          true,
			expectedErrorMessage: "SLO queries should start with SLO",
		},
		{
			name:                 "invalid - no SLO ID",
			inputQuery:           "SLO;",
			expectError:          true,
			expectedErrorMessage: "SLO ID should not be empty",
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
					assert.EqualValues(t, tc.expectedSLOID, query.GetSLOID())
				}
			}
		})
	}
}

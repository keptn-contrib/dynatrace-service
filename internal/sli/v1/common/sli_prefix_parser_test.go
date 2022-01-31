package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSLIPrefixParser tests the SLIPrefixParser
func TestSLIPrefixParser(t *testing.T) {
	tests := []struct {
		name                        string
		input                       string
		inputCount                  int
		expectedPiecesAssertionFunc func(t assert.TestingT, p SLIPieces)
		expectError                 bool
		expectedErrorMessage        string
	}{
		{
			name:       "valid empty input",
			inputCount: 1,
			expectedPiecesAssertionFunc: func(t assert.TestingT, p SLIPieces) {
				assert.EqualValues(t, 1, p.Count())
				assertPieceValueAtIndex(t, "", p, 0)
			},
		},
		{
			name:       "valid - input with two pieces",
			input:      "one;two",
			inputCount: 2,
			expectedPiecesAssertionFunc: func(t assert.TestingT, p SLIPieces) {
				assert.EqualValues(t, 2, p.Count())
				assertPieceValueAtIndex(t, "one", p, 0)
				assertPieceValueAtIndex(t, "two", p, 1)
			},
		},
		{
			name:       "valid - too many pieces",
			input:      "one;two;three;four",
			inputCount: 2,
			expectedPiecesAssertionFunc: func(t assert.TestingT, p SLIPieces) {
				assert.EqualValues(t, 2, p.Count())
				assertPieceValueAtIndex(t, "one", p, 0)
				assertPieceValueAtIndex(t, "two;three;four", p, 1)
			},
		},
		{
			name:       "valid - input with three pieces, one empty",
			input:      "one;;three",
			inputCount: 3,
			expectedPiecesAssertionFunc: func(t assert.TestingT, p SLIPieces) {
				assert.EqualValues(t, 3, p.Count())
				assertPieceValueAtIndex(t, "one", p, 0)
				assertPieceValueAtIndex(t, "", p, 1)
				assertPieceValueAtIndex(t, "three", p, 2)
			},
		},
		{
			name:                 "invalid - too few pieces",
			input:                "one;two",
			inputCount:           3,
			expectError:          true,
			expectedErrorMessage: "incorrect prefix",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pieces, err := NewSLIPrefixParser(tc.input, tc.inputCount).Parse()
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, pieces)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, pieces) {
					tc.expectedPiecesAssertionFunc(t, *pieces)
				}
				assert.Empty(t, tc.expectedErrorMessage, "fix test setup")
			}
		})
	}
}

func assertPieceValueAtIndex(t assert.TestingT, expectedValue string, pieces SLIPieces, index int) {
	v, err := pieces.Get(index)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedValue, v)
}

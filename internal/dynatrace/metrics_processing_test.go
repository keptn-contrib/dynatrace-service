package dynatrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateResultName(t *testing.T) {
	tests := []struct {
		name                    string
		dimensionMap            map[string]string
		expectedIndicatorSuffix string
	}{
		{
			name:                    "dimension without name: expect suffix with dimension value",
			dimensionMap:            map[string]string{"dt.entity.application": "APPLICATION-007CAB1ABEACDFE1"},
			expectedIndicatorSuffix: "APPLICATION-007CAB1ABEACDFE1",
		},
		{
			name: "dimension with name available: expect suffix with dimension name",
			dimensionMap: map[string]string{
				"dt.entity.application.name": "easytravel-ang.lab.dynatrace.org",
				"dt.entity.application":      "APPLICATION-007CAB1ABEACDFE1",
			},
			expectedIndicatorSuffix: "easytravel-ang.lab.dynatrace.org",
		},
		{
			name: "multiple dimensions with names: expect suffix with dimension names",
			dimensionMap: map[string]string{
				"dt.entity.application.name": "easytravel-ang.lab.dynatrace.org",
				"dt.entity.application":      "APPLICATION-007CAB1ABEACDFE1",
				"dt.entity.browser.name":     "Synthetic monitor",
				"dt.entity.browser":          "BROWSER-1CFF5AB60CE3BBAF",
			},
			expectedIndicatorSuffix: "easytravel-ang.lab.dynatrace.org Synthetic monitor",
		},
		{
			name: "multiple dimensions, but not all have names: expect suffix with dimension names where available, other dimensions where no names are available",
			dimensionMap: map[string]string{
				"dt.entity.application":  "APPLICATION-007CAB1ABEACDFE1",
				"dt.entity.browser.name": "Synthetic monitor",
				"dt.entity.browser":      "BROWSER-1CFF5AB60CE3BBAF",
			},
			expectedIndicatorSuffix: "APPLICATION-007CAB1ABEACDFE1 Synthetic monitor",
		},

		// Should not occur, but test it works as expected:
		{
			name:                    "no dimensions: expect no suffix",
			expectedIndicatorSuffix: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.expectedIndicatorSuffix, generateResultName(tt.dimensionMap))
		})
	}
}

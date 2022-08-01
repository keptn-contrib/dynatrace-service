package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScaleDataWithUnit(t *testing.T) {
	testConfigs := []struct {
		name           string
		unit           string
		inputValue     float64
		expectedResult float64
	}{
		// micro second
		{
			name:           "MicroSecond works",
			unit:           "MicroSecond",
			inputValue:     1000000.0,
			expectedResult: 1000.0,
		},
		{
			name:           "Microsecond works",
			unit:           "Microsecond",
			inputValue:     1000000.0,
			expectedResult: 1000.0,
		},
		{
			name:           "microsecond works",
			unit:           "microsecond",
			inputValue:     1000000.0,
			expectedResult: 1000.0,
		},
		{
			name:           "microSecond works",
			unit:           "microSecond",
			inputValue:     1000000.0,
			expectedResult: 1000.0,
		},
		// byte
		{
			name:           "Byte works",
			unit:           "Byte",
			inputValue:     1024.0,
			expectedResult: 1.0,
		},
		{
			name:           "byte works",
			unit:           "byte",
			inputValue:     1024.0,
			expectedResult: 1.0,
		},
		// unknown metric is unchanged
		{
			name:           "MicroSeconds does not work",
			unit:           "MicroSeconds",
			inputValue:     123.0,
			expectedResult: 123.0,
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			actual := ScaleData(tc.unit, tc.inputValue)

			assert.EqualValues(t, tc.expectedResult, actual)
		})
	}
}

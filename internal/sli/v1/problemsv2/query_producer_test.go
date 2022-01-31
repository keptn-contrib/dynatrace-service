package problemsv2

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"

	"github.com/stretchr/testify/assert"
)

func TestQueryProducer_Produce(t *testing.T) {
	testConfigs := []struct {
		name                   string
		inputPV2Query          problems.Query
		expectedPV2QueryString string
	}{
		{
			name:                   "valid with no problem or entity selectors",
			inputPV2Query:          problems.NewQuery("", ""),
			expectedPV2QueryString: "PV2;",
		},
		{
			name:                   "valid with both problem and entity selectors",
			inputPV2Query:          problems.NewQuery("status(open)", "mzId(7030365576649815430)"),
			expectedPV2QueryString: "PV2;entitySelector=mzId(7030365576649815430)&problemSelector=status(open)",
		},
		{
			name:                   "valid with just problem selector",
			inputPV2Query:          problems.NewQuery("status(open)", ""),
			expectedPV2QueryString: "PV2;problemSelector=status(open)",
		},
		{
			name:                   "valid with just entity selector",
			inputPV2Query:          problems.NewQuery("", "mzId(7030365576649815430)"),
			expectedPV2QueryString: "PV2;entitySelector=mzId(7030365576649815430)",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			pv2QueryString := NewQueryProducer(tc.inputPV2Query).Produce()
			assert.Equal(t, tc.expectedPV2QueryString, pv2QueryString)
		})
	}
}

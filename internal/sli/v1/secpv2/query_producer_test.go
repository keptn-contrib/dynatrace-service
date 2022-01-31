package secpv2

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/secpv2"

	"github.com/stretchr/testify/assert"
)

func TestQueryProducer_Produce(t *testing.T) {
	testConfigs := []struct {
		name                      string
		inputSECPV2Query          secpv2.Query
		expectedSECPV2QueryString string
	}{
		{
			name:                      "valid with no security problem selectors",
			inputSECPV2Query:          secpv2.NewQuery(""),
			expectedSECPV2QueryString: "SECPV2;",
		},
		{
			name:                      "valid with security problem selector",
			inputSECPV2Query:          secpv2.NewQuery("status(open)"),
			expectedSECPV2QueryString: "SECPV2;securityProblemSelector=status(open)",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			pv2QueryString := NewQueryProducer(tc.inputSECPV2Query).Produce()
			assert.Equal(t, tc.expectedSECPV2QueryString, pv2QueryString)
		})
	}
}

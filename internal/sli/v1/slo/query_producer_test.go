package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryProducer_Produce(t *testing.T) {
	testConfigs := []struct {
		name                   string
		inputSLOQuery          Query
		expectedSLOQueryString string
	}{
		{
			name:                   "valid with security problem selector",
			inputSLOQuery:          newQuery(t, "524ca177-849b-3e8c-8175-42b93fbc33c5"),
			expectedSLOQueryString: "SLO;524ca177-849b-3e8c-8175-42b93fbc33c5",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			pv2QueryString := NewQueryProducer(tc.inputSLOQuery).Produce()
			assert.Equal(t, tc.expectedSLOQueryString, pv2QueryString)
		})
	}
}

func newQuery(t *testing.T, sloID string) Query {
	query, err := NewQuery(sloID)
	assert.NoError(t, err)
	assert.NotNil(t, query)
	return *query
}

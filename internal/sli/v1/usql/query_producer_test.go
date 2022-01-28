package usql

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
	"github.com/stretchr/testify/assert"
)

func TestQueryProducer_Produce(t *testing.T) {
	testConfigs := []struct {
		name                    string
		inputUSQLQuery          Query
		expectedUSQLQueryString string
	}{
		{
			name:                    "valid",
			inputUSQLQuery:          newQuery(t, "COLUMN_CHART", "iPad mini", "SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device"),
			expectedUSQLQueryString: "USQL;COLUMN_CHART;iPad mini;SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device",
		},
		{
			name:                    "valid - extra semi colon in query",
			inputUSQLQuery:          newQuery(t, "COLUMN_CHART", "iPad mini", "SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria;Italy') GROUP BY device"),
			expectedUSQLQueryString: "USQL;COLUMN_CHART;iPad mini;SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria;Italy') GROUP BY device",
		},

		{
			name:                    "valid - SINGLE_VALUE with should have no dimension",
			inputUSQLQuery:          newQuery(t, "SINGLE_VALUE", "", "SELECT AVG(duration) FROM usersession WHERE country IN('Austria')"),
			expectedUSQLQueryString: "USQL;SINGLE_VALUE;;SELECT AVG(duration) FROM usersession WHERE country IN('Austria')",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			usqlQueryString := NewQueryProducer(tc.inputUSQLQuery).Produce()
			assert.Equal(t, tc.expectedUSQLQueryString, usqlQueryString)
		})
	}
}

func newQuery(t *testing.T, resultType string, dimension string, queryString string) Query {
	innerQuery, err := usql.NewQuery(queryString)
	assert.NoError(t, err)
	assert.NotNil(t, innerQuery)

	query, err := NewQuery(resultType, dimension, *innerQuery)
	assert.NoError(t, err)
	assert.NotNil(t, query)
	return *query
}

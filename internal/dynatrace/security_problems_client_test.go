package dynatrace

import (
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestSecurityProblemsClient_GetTotalCountByQuery_None(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/securityProblems?from=1571649084000&to=1571649085000&securityProblemSelector=status(OPEN)", "./testdata/test_securityproblemsclient_gettotalcountbyquery_0.json")

	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	securityProblemQuery := "securityProblemSelector=status(OPEN)"

	totalSecurityProblemCount, err := NewSecurityProblemsClient(dtClient).GetTotalCountByQuery(securityProblemQuery, startTime, endTime)

	assert.NoError(t, err)
	assert.EqualValues(t, 0, totalSecurityProblemCount)
}

func TestSecurityProblemsClient_GetTotalCountByQuery_Some(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/securityProblems?from=1638255600000&to=1638259200000&securityProblemSelector=status(OPEN)", "./testdata/test_securityproblemsclient_gettotalcountbyquery_177.json")

	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	startTime := time.Unix(1638255600, 0).UTC()
	endTime := time.Unix(1638259200, 0).UTC()
	securityProblemQuery := "securityProblemSelector=status(OPEN)"

	totalSecurityProblemCount, err := NewSecurityProblemsClient(dtClient).GetTotalCountByQuery(securityProblemQuery, startTime, endTime)

	assert.NoError(t, err)
	assert.EqualValues(t, 177, totalSecurityProblemCount)
}

package dynatrace

import (
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"testing"
	"time"
)

func TestExecuteGetDynatraceSecurityProblems(t *testing.T) {
	handler := test.NewFileBasedURLHandler()
	handler.AddStartsWith("/api/v2/securityProblems", "./testdata/test_get_securityproblems.json")

	dtClient, _, teardown := createDynatraceClient(handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(OPEN)"

	problemResult, err := NewSecurityProblemsClient(dtClient).GetByQuery(problemQuery, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if problemResult == nil {
		t.Fatal("No Problem Result returned for " + problemQuery)
	}

	if problemResult.TotalCount != 0 {
		t.Error("Not returning expected value for Problem Query")
	}
}

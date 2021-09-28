package dynatrace

import (
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"testing"
	"time"
)

func TestExecuteGetDynatraceProblems(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddStartsWith("/api/v2/problems", "./testdata/test_get_problems.json")

	dtClient, _, teardown := createDynatraceClient(handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(open)"
	problemResult, err := NewProblemsV2Client(dtClient).GetByQuery(problemQuery, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if problemResult == nil {
		t.Fatal("No Problem Result returned for " + problemQuery)
	}

	if problemResult.TotalCount != 1 {
		t.Error("Not returning expected value for Problem Query")
	}
}

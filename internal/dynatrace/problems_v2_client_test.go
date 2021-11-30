package dynatrace

import (
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestExecuteGetDynatraceProblems(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddStartsWith("/api/v2/problems", "./testdata/test_get_problems.json")

	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(open)"
	totalProblemCount, err := NewProblemsV2Client(dtClient).GetTotalCountByQuery(problemQuery, startTime, endTime)

	assert.NoError(t, err)
	assert.EqualValues(t, 1, totalProblemCount)
}

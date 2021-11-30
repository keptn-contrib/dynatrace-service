package dynatrace

import (
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestProblemsV2Client_GetTotalCountByQuery(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/problems?from=1571649084000&to=1571649085000&problemEntity=status(open)", "./testdata/test_problemsv2client_gettotalcountbyquery.json")

	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := "problemEntity=status(open)"
	totalProblemCount, err := NewProblemsV2Client(dtClient).GetTotalCountByQuery(problemQuery, startTime, endTime)

	assert.NoError(t, err)
	assert.EqualValues(t, 1, totalProblemCount)
}

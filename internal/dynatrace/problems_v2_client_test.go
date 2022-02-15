package dynatrace

import (
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestProblemsV2Client_GetTotalCountByQuery(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/problems?from=1571649084000&problemSelector=status%28%22open%22%29&to=1571649085000", "./testdata/test_problemsv2client_gettotalcountbyquery.json")

	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	problemQuery := problems.NewQuery("status(\"open\")", "")
	totalProblemCount, err := NewProblemsV2Client(dtClient).GetTotalCountByQuery(NewProblemsV2ClientQueryParameters(problemQuery, startTime, endTime))

	assert.NoError(t, err)
	assert.EqualValues(t, 1, totalProblemCount)
}

func TestProblemsV2Client_GetStatusById(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/problems/-6004362228644432354_1638271020000V2", "./testdata/test_problemsv2client_getstatusbyid.json")

	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	status, err := NewProblemsV2Client(dtClient).GetStatusByID("-6004362228644432354_1638271020000V2")

	assert.NoError(t, err)
	assert.EqualValues(t, ProblemStatusOpen, status)
}

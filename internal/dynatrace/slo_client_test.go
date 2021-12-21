package dynatrace

import (
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestExecuteGetDynatraceSLO(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(SLOPath+"/524ca177-849b-3e8c-8175-42b93fbc33c5?from=1571649084000&to=1571649085000", "./testdata/test_get_slo_id.json")
	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	sloID := "524ca177-849b-3e8c-8175-42b93fbc33c5"
	sloResult, err := NewSLOClient(dtClient).Get(sloID, startTime, endTime)

	assert.NoError(t, err)
	assert.NotNil(t, sloResult, "No SLO Result returned for "+sloID)
	assert.EqualValues(t, 95.66405076939219, sloResult.EvaluatedPercentage, "Not returning expected value for SLO")
}

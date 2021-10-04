package dynatrace

import (
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestExecuteGetDynatraceSLO(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddStartsWith("/api/v2/slo", "./testdata/test_get_slo_id.json")

	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	sloID := "524ca177-849b-3e8c-8175-42b93fbc33c5"
	sloResult, err := NewSLOClient(dtClient).Get(sloID, startTime, endTime)

	if err != nil {
		t.Error(err)
	}

	if sloResult == nil {
		t.Errorf("No SLO Result returned for " + sloID)
	}

	if sloResult.EvaluatedPercentage != 95.66405076939219 {
		t.Error("Not returning expected value for SLO")
	}
}

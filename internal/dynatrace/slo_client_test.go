package dynatrace

import (
	"context"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestExecuteGetDynatraceSLO(t *testing.T) {
	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(SLOPath+"/524ca177-849b-3e8c-8175-42b93fbc33c5?from=1571649084000&timeFrame=GTF&to=1571649085000", "./testdata/test_get_slo_id.json")
	dtClient, _, teardown := createDynatraceClient(t, handler)
	defer teardown()

	timeframe, err := common.NewTimeframeParser("2019-10-21T09:11:24Z", "2019-10-21T09:11:25Z").Parse()
	assert.NoError(t, err)

	sloID := "524ca177-849b-3e8c-8175-42b93fbc33c5"
	sloResult, err := NewSLOClient(dtClient).Get(context.TODO(), NewSLOClientGetRequest(sloID, *timeframe))

	assert.NoError(t, err)
	assert.NotNil(t, sloResult, "No SLO Result returned for "+sloID)
	assert.EqualValues(t, 95.66405076939219, sloResult.EvaluatedPercentage, "Not returning expected value for SLO")
}

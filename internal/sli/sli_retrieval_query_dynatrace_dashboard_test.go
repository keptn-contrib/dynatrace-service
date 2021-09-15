package sli

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueryDynatraceDashboardForSLIs(t *testing.T) {
	keptnEvent := createKeptnEvent(QUALITYGATE_PROJECT, QUALITYGATE_STAGE, QUALTIYGATE_SERVICE)
	dh, teardown := createRetrieval(keptnEvent)
	defer teardown()

	startTime := time.Unix(1571649084, 0).UTC()
	endTime := time.Unix(1571649085, 0).UTC()
	result, err := dh.QueryDynatraceDashboardForSLIs(keptnEvent, common.DynatraceConfigDashboardQUERY, startTime, endTime)

	assert.Nil(t, err)
	assert.NotNil(t, result, "No result returned")
	assert.NotNil(t, result.dashboardLink, "No dashboard link label generated")
	assert.NotNil(t, result.dashboard, "No Dashboard JSON returned")
	assert.NotNil(t, result.sli, "No SLI returned")
	assert.NotNil(t, result.slo, "No SLO returned")
	assert.NotNil(t, result.sliResults, "No SLI Results returned")

	const expectedSLOs = 14
	assert.Equal(t, expectedSLOs, len(result.sli.Indicators))
	assert.Equal(t, expectedSLOs, len(result.slo.Objectives))
	assert.EqualValues(t, &keptnapi.SLOScore{Pass: "90%", Warning: "70%"}, result.slo.TotalScore)
	assert.EqualValues(
		t,
		&keptnapi.SLOComparison{
			CompareWith:               "single_result",
			IncludeResultWithScore:    "pass",
			NumberOfComparisonResults: 1,
			AggregateFunction:         "avg",
		},
		result.slo.Comparison)
	assert.Equal(t, expectedSLOs, len(result.sliResults))
}

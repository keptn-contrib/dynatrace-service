package dashboard

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

func TestMetricsQueryProcessing_Process(t *testing.T) {
	t.Run("AverageMemoryUsage", testMetricsQueryProcessing_Process_AverageMemoryUsage)
}

func testMetricsQueryProcessing_Process_AverageMemoryUsage(t *testing.T) {

	const metricQuery = "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000"

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact("/api/v2/metrics/query?"+metricQuery,
		"./testdata/metrics_query_processing_test/avg_mem_usage/metrics_get_by_query_containers.memory_usage2.json")

	dtClient, _, teardown := createDynatraceClient(handler)
	defer teardown()

	processing := NewMetricsQueryProcessing(dtClient)
	tileResults := processing.Process(0, &keptncommon.SLO{
		SLI:    "avg_mem_usage",
		Weight: 1,
	}, &queryComponents{
		metricID:              "builtin:containers.memory_usage2:merge(container_id):merge(dt.entity.docker_container_group_instance):avg:names",
		metricUnit:            "Byte",
		metricQuery:           metricQuery,
		fullMetricQueryString: metricQuery,
	})

	assert.EqualValues(t, tileResults, []*TileResult{createExpectedTileResult("avg_mem_usage", 48975.83345935025, "MV2;Byte;entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000")})
}

func createExpectedTileResult(sliName string, sliValue float64, sliQuery string) *TileResult {
	return &TileResult{
		sliResult: &v0_2_0.SLIResult{
			Metric:  sliName,
			Value:   sliValue,
			Success: true,
		},
		objective: &keptncommon.SLO{
			SLI:    sliName,
			Weight: 1,
		},
		sliName:  sliName,
		sliQuery: sliQuery,
	}
}

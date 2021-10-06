package dashboard

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

func TestMetricsQueryProcessing_Process(t *testing.T) {
	type args struct {
		noOfDimensionsInChart int
		sloDefinition         *keptncommon.SLO
		metricQueryComponents *queryComponents
	}
	tests := []struct {
		name                      string
		metricsGetByQueryRequest  string
		metricsGetByQueryFilename string
		args                      args
		want                      []*TileResult
	}{
		{
			name:                      "avg_mem_usage",
			metricsGetByQueryRequest:  "/api/v2/metrics/query?entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			metricsGetByQueryFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_avg_mem_usage1.json",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "avg_mem_usage",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricID:              "builtin:containers.memory_usage2:merge(container_id):merge(dt.entity.docker_container_group_instance):avg:names",
					metricUnit:            "Byte",
					metricQuery:           "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
					fullMetricQueryString: "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
				},
			},
			want: []*TileResult{
				&TileResult{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "avg_mem_usage",
						Value:   48975.83345935025,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "avg_mem_usage",
						Weight: 1,
					},
					sliName:  "avg_mem_usage",
					sliQuery: "MV2;Byte;entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(tt.metricsGetByQueryRequest, tt.metricsGetByQueryFilename)

			dtClient, _, teardown := createDynatraceClient(handler)
			defer teardown()

			r := NewMetricsQueryProcessing(dtClient)
			if got := r.Process(tt.args.noOfDimensionsInChart, tt.args.sloDefinition, tt.args.metricQueryComponents); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricsQueryProcessing.Process() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createDynatraceClient(handler http.Handler) (dynatrace.ClientInterface, string, func()) {
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	dtCredentials := &credentials.DTCredentials{
		Tenant:   url,
		ApiToken: "test",
	}

	dh := dynatrace.NewClientWithHTTP(dtCredentials, httpClient)

	return dh, url, teardown
}

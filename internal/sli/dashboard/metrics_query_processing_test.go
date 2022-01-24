package dashboard

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

func TestMetricsQueryProcessing_Process(t *testing.T) {

	startTime, _ := common.ParseUnixTimestamp("2021-10-05T08:00:00Z") // 1633420800000
	endTime, _ := common.ParseUnixTimestamp("2021-10-06T08:00:00Z")   // 1633507200000

	type args struct {
		noOfDimensionsInChart int
		sloDefinition         *keptncommon.SLO
		metricQueryComponents *queryComponents
	}

	tests := []struct {
		name                        string
		metricQueryResponseFilename string
		fullMetricQueryString       string
		args                        args
		expectedResults             []*TileResult
	}{

		// MV2 prefix tests
		{
			name:                        "csrt - MicroSecond - should MV2 prefix",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_service.response.client.json",
			fullMetricQueryString:       "entitySelector=type%28SERVICE%29&from=1633420800000&metricSelector=builtin%3Aservice.response.client%3Amerge%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "csrt",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricsQuery: createMetricsQuery("builtin:service.response.client:merge(\"dt.entity.service\"):avg:names", "type(SERVICE)"),
					metricUnit:   "MicroSecond",
					startTime:    startTime,
					endTime:      endTime,
				},
			},
			expectedResults: []*TileResult{
				{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "csrt",
						Value:   15.868648438045174,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "csrt",
						Weight: 1,
					},
					sliName:  "csrt",
					sliQuery: "MV2;MicroSecond;entitySelector=type(SERVICE)&metricSelector=builtin:service.response.client:merge(\"dt.entity.service\"):avg:names",
				},
			},
		},
		{
			name:                        "cmu - Byte - should MV2 prefix",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_containers.memory_usage2.json",
			fullMetricQueryString:       "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "cmu",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricsQuery: createMetricsQuery("builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names", "type(DOCKER_CONTAINER_GROUP_INSTANCE)"),
					metricUnit:   "Byte",
					startTime:    startTime,
					endTime:      endTime,
				},
			},
			expectedResults: []*TileResult{
				{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "cmu",
						Value:   48975.83345935025,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "cmu",
						Weight: 1,
					},
					sliName:  "cmu",
					sliQuery: "MV2;Byte;entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)&metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names",
				},
			},
		},
		{
			name:                        "hdqc - Count - should not MV2 prefix",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_host.dns.queryCount.json",
			fullMetricQueryString:       "entitySelector=type%28HOST%29&from=1633420800000&metricSelector=builtin%3Ahost.dns.queryCount%3Amerge%28%22dnsServerIp%22%29%3Amerge%28%22dt.entity.host%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "hdqc",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricsQuery: createMetricsQuery("builtin:host.dns.queryCount:merge(\"dnsServerIp\"):merge(\"dt.entity.host\"):avg:names", "type(HOST)"),
					metricUnit:   "Count",
					startTime:    startTime,
					endTime:      endTime,
				},
			},
			expectedResults: []*TileResult{
				{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "hdqc",
						Value:   96.94525462962963,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "hdqc",
						Weight: 1,
					},
					sliName:  "hdqc",
					sliQuery: "entitySelector=type(HOST)&metricSelector=builtin:host.dns.queryCount:merge(\"dnsServerIp\"):merge(\"dt.entity.host\"):avg:names",
				},
			},
		},

		// Metric ID syntax variants tests
		{
			name:                        "cmu - All dimensions quotes",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_containers.memory_usage2.json",
			fullMetricQueryString:       "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "cmu",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricsQuery: createMetricsQuery("builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names", "type(DOCKER_CONTAINER_GROUP_INSTANCE)"),
					metricUnit:   "Byte",
					startTime:    startTime,
					endTime:      endTime,
				},
			},
			expectedResults: []*TileResult{
				{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "cmu",
						Value:   48975.83345935025,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "cmu",
						Weight: 1,
					},
					sliName:  "cmu",
					sliQuery: "MV2;Byte;entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)&metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names",
				},
			},
		},
		{
			name:                        "cmu - No dimensions quotes",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_containers.memory_usage2.json",
			fullMetricQueryString:       "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28container_id%29%3Amerge%28dt.entity.docker_container_group_instance%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "cmu",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{

					metricsQuery: createMetricsQuery("builtin:containers.memory_usage2:merge(container_id):merge(dt.entity.docker_container_group_instance):avg:names", "type(DOCKER_CONTAINER_GROUP_INSTANCE)"),
					metricUnit:   "Byte",
					startTime:    startTime,
					endTime:      endTime,
				},
			},
			expectedResults: []*TileResult{
				{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "cmu",
						Value:   48975.83345935025,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "cmu",
						Weight: 1,
					},
					sliName:  "cmu",
					sliQuery: "MV2;Byte;entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)&metricSelector=builtin:containers.memory_usage2:merge(container_id):merge(dt.entity.docker_container_group_instance):avg:names",
				},
			},
		},
		{
			name:                        "cmu - Just entity dimensions quotes",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_containers.memory_usage2.json",
			fullMetricQueryString:       "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28container_id%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "cmu",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricsQuery: createMetricsQuery("builtin:containers.memory_usage2:merge(container_id):merge(\"dt.entity.docker_container_group_instance\"):avg:names", "type(DOCKER_CONTAINER_GROUP_INSTANCE)"),
					metricUnit:   "Byte",
					startTime:    startTime,
					endTime:      endTime,
				},
			},
			expectedResults: []*TileResult{
				{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "cmu",
						Value:   48975.83345935025,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "cmu",
						Weight: 1,
					},
					sliName:  "cmu",
					sliQuery: "MV2;Byte;entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)&metricSelector=builtin:containers.memory_usage2:merge(container_id):merge(\"dt.entity.docker_container_group_instance\"):avg:names",
				},
			},
		},
		{
			name:                        "cmu - Just non-entity dimensions quotes",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_containers.memory_usage2.json",
			fullMetricQueryString:       "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28dt.entity.docker_container_group_instance%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "cmu",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricsQuery: createMetricsQuery("builtin:containers.memory_usage2:merge(\"container_id\"):merge(dt.entity.docker_container_group_instance):avg:names", "type(DOCKER_CONTAINER_GROUP_INSTANCE)"),
					metricUnit:   "Byte",
					startTime:    startTime,
					endTime:      endTime,
				},
			},
			expectedResults: []*TileResult{
				{
					sliResult: &v0_2_0.SLIResult{
						Metric:  "cmu",
						Value:   48975.83345935025,
						Success: true,
					},
					objective: &keptncommon.SLO{
						SLI:    "cmu",
						Weight: 1,
					},
					sliName:  "cmu",
					sliQuery: "MV2;Byte;entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)&metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(dt.entity.docker_container_group_instance):avg:names",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact("/api/v2/metrics/query?"+tt.fullMetricQueryString,
				tt.metricQueryResponseFilename)

			dtClient, _, teardown := createDynatraceClient(t, handler)
			defer teardown()

			processing := NewMetricsQueryProcessing(dtClient)
			tileResults := processing.Process(tt.args.noOfDimensionsInChart, tt.args.sloDefinition, tt.args.metricQueryComponents)

			assert.EqualValues(t, tt.expectedResults, tileResults)
		})
	}
}

func createMetricsQuery(metricSelector string, entitySelector string) *metrics.Query {
	query, _ := metrics.NewQuery(metricSelector, entitySelector)
	return query
}

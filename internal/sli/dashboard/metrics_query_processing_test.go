package dashboard

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"
)

func TestMetricsQueryProcessing_Process(t *testing.T) {
	type args struct {
		noOfDimensionsInChart int
		sloDefinition         *keptncommon.SLO
		metricQueryComponents *queryComponents
	}

	tests := []struct {
		name                        string
		metricQueryResponseFilename string
		args                        args
		expectedResults             []*TileResult
	}{

		// MV2 prefix tests
		{
			name:                        "csrt - MicroSecond - should MV2 prefix",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_service.response.client.json",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "csrt",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricID:              "builtin:service.response.client:merge(\"dt.entity.service\"):avg:names",
					metricUnit:            "MicroSecond",
					metricQuery:           "metricSelector=builtin:service.response.client:merge(\"dt.entity.service\"):avg:names&entitySelector=type(SERVICE)",
					fullMetricQueryString: "entitySelector=type%28SERVICE%29&from=1633420800000&metricSelector=builtin%3Aservice.response.client%3Amerge%28%22dt.entity.service%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
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
					sliQuery: "MV2;MicroSecond;metricSelector=builtin:service.response.client:merge(\"dt.entity.service\"):avg:names&entitySelector=type(SERVICE)",
				},
			},
		},
		{
			name:                        "cmu - Byte - should MV2 prefix",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_containers.memory_usage2.json",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "cmu",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricID:              "builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names",
					metricUnit:            "Byte",
					metricQuery:           "metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names&entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)",
					fullMetricQueryString: "entitySelector=type%28DOCKER_CONTAINER_GROUP_INSTANCE%29&from=1633420800000&metricSelector=builtin%3Acontainers.memory_usage2%3Amerge%28%22container_id%22%29%3Amerge%28%22dt.entity.docker_container_group_instance%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
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
					sliQuery: "MV2;Byte;metricSelector=builtin:containers.memory_usage2:merge(\"container_id\"):merge(\"dt.entity.docker_container_group_instance\"):avg:names&entitySelector=type(DOCKER_CONTAINER_GROUP_INSTANCE)",
				},
			},
		},
		{
			name:                        "hdqc - Count - should not MV2 prefix",
			metricQueryResponseFilename: "./testdata/metrics_query_processing_test/metrics_get_by_query_builtin_host.dns.queryCount.json",
			args: args{
				noOfDimensionsInChart: 0,
				sloDefinition: &keptncommon.SLO{
					SLI:    "hdqc",
					Weight: 1,
				},
				metricQueryComponents: &queryComponents{
					metricID:              "builtin:host.dns.queryCount:merge(\"dnsServerIp\"):merge(\"dt.entity.host\"):avg:names",
					metricUnit:            "Count",
					metricQuery:           "metricSelector=builtin:host.dns.queryCount:merge(\"dnsServerIp\"):merge(\"dt.entity.host\"):avg:names&entitySelector=type(HOST)",
					fullMetricQueryString: "entitySelector=type%28HOST%29&from=1633420800000&metricSelector=builtin%3Ahost.dns.queryCount%3Amerge%28%22dnsServerIp%22%29%3Amerge%28%22dt.entity.host%22%29%3Aavg%3Anames&resolution=Inf&to=1633507200000",
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
					sliQuery: "metricSelector=builtin:host.dns.queryCount:merge(\"dnsServerIp\"):merge(\"dt.entity.host\"):avg:names&entitySelector=type(HOST)",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact("/api/v2/metrics/query?"+tt.args.metricQueryComponents.fullMetricQueryString,
				tt.metricQueryResponseFilename)

			dtClient, _, teardown := createDynatraceClient(handler)
			defer teardown()

			processing := NewMetricsQueryProcessing(dtClient)
			tileResults := processing.Process(tt.args.noOfDimensionsInChart, tt.args.sloDefinition, tt.args.metricQueryComponents)

			assert.EqualValues(t, tt.expectedResults, tileResults)
		})
	}
}

package dashboard

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
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
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact("/api/v2/metrics/query?"+tt.args.metricQueryComponents.fullMetricQueryString,
				tt.metricQueryResponseFilename)

			dtClient, _, teardown := createDynatraceClient(handler)
			defer teardown()

			processing := NewMetricsQueryProcessing(dtClient)
			tileResults := processing.Process(tt.args.noOfDimensionsInChart, tt.args.sloDefinition, tt.args.metricQueryComponents)

			assert.EqualValues(t, tileResults, tt.expectedResults)
		})
	}
}

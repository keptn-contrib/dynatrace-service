package dashboard

import (
	"context"
	"errors"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
)

type MetricsQueryProcessing struct {
	metricsProcessing dynatrace.MetricsProcessingInterface
}

func NewMetricsQueryProcessing(processing dynatrace.MetricsProcessingInterface) *MetricsQueryProcessing {
	return &MetricsQueryProcessing{
		metricsProcessing: processing,
	}
}

// Process generates SLI & SLO definitions based on the metric query and the number of dimensions in the chart definition.
func (r *MetricsQueryProcessing) Process(ctx context.Context, sloDefinition keptncommon.SLO, metricsQuery metrics.Query, timeframe common.Timeframe) []TileResult {
	request := dynatrace.NewMetricsClientQueryRequest(metricsQuery, timeframe)

	processingResultsSet, err := r.metricsProcessing.ProcessRequest(ctx, request)
	if err != nil {
		var qpErrorType *dynatrace.MetricsQueryProcessingError
		var qrmvErrorType *dynatrace.MetricsQueryReturnedMultipleValuesError
		if errors.As(err, &qpErrorType) || errors.As(err, &qrmvErrorType) {
			return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
		}
		return []TileResult{newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
	}
	return processResults(sloDefinition, request, processingResultsSet.Results())
}

func processResults(sloDefinition keptncommon.SLO, request dynatrace.MetricsClientQueryRequest, results []dynatrace.MetricsProcessingResult) []TileResult {
	if len(results) == 0 {
		return []TileResult{}
	}

	if len(results) == 1 {
		return []TileResult{newSuccessfulTileResult(sloDefinition, results[0].Value(), request.RequestString())}
	}

	var tileResults []TileResult
	for _, result := range results {
		tileResults = append(tileResults, newSuccessfulTileResult(createSLODefinitionForName(sloDefinition, result.Name()), result.Value(), request.RequestString()))
	}

	return tileResults
}

func createSLODefinitionForName(baseSLODefinition keptncommon.SLO, name string) keptncommon.SLO {
	return keptncommon.SLO{
		SLI:         baseSLODefinition.SLI + "_" + cleanIndicatorName(name),
		DisplayName: baseSLODefinition.DisplayName + " (" + name + ")",
		Weight:      baseSLODefinition.Weight,
		KeySLI:      baseSLODefinition.KeySLI,
		Pass:        baseSLODefinition.Pass,
		Warning:     baseSLODefinition.Warning,
	}
}

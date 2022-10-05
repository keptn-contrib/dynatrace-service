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

func NewMetricsQueryProcessing(client dynatrace.ClientInterface) *MetricsQueryProcessing {
	return &MetricsQueryProcessing{
		metricsProcessing: dynatrace.NewMetricsProcessing(dynatrace.NewMetricsClient(client)),
	}
}

func NewMetricsQueryProcessingThatAllowsOnlyOneResult(client dynatrace.ClientInterface) *MetricsQueryProcessing {
	return &MetricsQueryProcessing{
		metricsProcessing: dynatrace.NewMetricsProcessingThatAllowsOnlyOneResult(dynatrace.NewMetricsClient(client)),
	}
}

// Process generates SLI & SLO definitions based on the metric query and the number of dimensions in the chart definition.
func (r *MetricsQueryProcessing) Process(ctx context.Context, sloDefinition keptncommon.SLO, metricsQuery metrics.Query, timeframe common.Timeframe) []TileResult {
	request := dynatrace.NewMetricsClientQueryRequest(metricsQuery, timeframe)
	processingResults, err := r.metricsProcessing.ProcessRequest(ctx, request)
	if err != nil {
		return r.createTileResultsForError(sloDefinition, request, err)
	}
	return r.processResults(sloDefinition, processingResults)
}

func (r *MetricsQueryProcessing) createTileResultsForError(sloDefinition keptncommon.SLO, request dynatrace.MetricsClientQueryRequest, err error) []TileResult {
	messagePrefix := "Could not process tile: "
	var qpErrorType *dynatrace.MetricsQueryProcessingError
	if errors.As(err, &qpErrorType) {
		return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), messagePrefix+err.Error())}
	}
	return []TileResult{newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), messagePrefix+err.Error())}

}

func (r *MetricsQueryProcessing) processResults(sloDefinition keptncommon.SLO, processingResults *dynatrace.MetricsProcessingResults) []TileResult {
	request := processingResults.Request()
	results := processingResults.Results()
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

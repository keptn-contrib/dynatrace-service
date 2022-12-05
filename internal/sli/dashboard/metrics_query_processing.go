package dashboard

import (
	"context"
	"errors"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
)

type MetricsQueryProcessing struct {
	metricsProcessing dynatrace.MetricsProcessingInterface
}

func NewMetricsQueryProcessing(client dynatrace.ClientInterface, targetUnitID string) *MetricsQueryProcessing {
	metricsClient := dynatrace.NewMetricsClient(client)
	unitsClient := dynatrace.NewEnhancedMetricsUnitsDecorator(dynatrace.NewMetricsUnitsClient(client))
	return &MetricsQueryProcessing{
		metricsProcessing: dynatrace.NewConvertUnitMetricsProcessingDecorator(
			metricsClient,
			unitsClient,
			targetUnitID,
			dynatrace.NewRetryForSingleValueMetricsProcessingDecorator(
				metricsClient,
				dynatrace.NewMetricsProcessing(metricsClient),
			),
		),
	}
}

func NewMetricsQueryProcessingThatAllowsOnlyOneResult(client dynatrace.ClientInterface, targetUnitID string) *MetricsQueryProcessing {
	metricsClient := dynatrace.NewMetricsClient(client)
	unitsClient := dynatrace.NewMetricsUnitsClient(client)
	return &MetricsQueryProcessing{
		metricsProcessing: dynatrace.NewConvertUnitMetricsProcessingDecorator(
			metricsClient,
			unitsClient,
			targetUnitID,
			dynatrace.NewRetryForSingleValueMetricsProcessingDecorator(
				metricsClient,
				dynatrace.NewMetricsProcessingThatAllowsOnlyOneResult(metricsClient),
			),
		),
	}
}

// Process generates SLI & SLO definitions based on the metric query and the number of dimensions in the chart definition.
func (r *MetricsQueryProcessing) Process(ctx context.Context, sloDefinition keptncommon.SLO, metricsQuery metrics.Query, timeframe common.Timeframe) []result.SLIWithSLO {
	request := dynatrace.NewMetricsClientQueryRequest(metricsQuery, timeframe)
	processingResults, err := r.metricsProcessing.ProcessRequest(ctx, request)
	if err != nil {
		return r.createTileResultsForError(sloDefinition, request, err)
	}
	return r.processResults(sloDefinition, processingResults)
}

func (r *MetricsQueryProcessing) createTileResultsForError(sloDefinition keptncommon.SLO, request dynatrace.MetricsClientQueryRequest, err error) []result.SLIWithSLO {
	messagePrefix := "Could not process tile: "
	var qpErrorType *dynatrace.MetricsQueryProcessingError
	if errors.As(err, &qpErrorType) {
		return []result.SLIWithSLO{result.NewWarningSLIWithSLOAndQuery(sloDefinition, request.RequestString(), messagePrefix+err.Error())}
	}
	return []result.SLIWithSLO{result.NewFailedSLIWithSLOAndQuery(sloDefinition, request.RequestString(), messagePrefix+err.Error())}

}

func (r *MetricsQueryProcessing) processResults(sloDefinition keptncommon.SLO, processingResults *dynatrace.MetricsProcessingResults) []result.SLIWithSLO {
	request := processingResults.Request()
	results := processingResults.Results()
	if len(results) == 1 {
		return []result.SLIWithSLO{result.NewSuccessfulSLIWithSLOAndQuery(sloDefinition, results[0].Value(), request.RequestString())}
	}

	var tileResults []result.SLIWithSLO
	for _, r := range results {
		tileResults = append(tileResults, result.NewSuccessfulSLIWithSLOAndQuery(createSLODefinitionForName(sloDefinition, r.Name()), r.Value(), request.RequestString()))
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

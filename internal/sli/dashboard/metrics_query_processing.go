package dashboard

import (
	"context"
	"errors"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/ff"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
)

type MetricsQueryProcessing struct {
	metricsProcessing dynatrace.MetricsProcessingInterface
	featureFlags      ff.GetSLIFeatureFlags
}

func NewMetricsQueryProcessing(client dynatrace.ClientInterface, targetUnitID string, flags ff.GetSLIFeatureFlags) *MetricsQueryProcessing {
	metricsClient := dynatrace.NewMetricsClient(client)

	return newMetricsQueryProcessing(
		dynatrace.NewConvertUnitsAndRetryForSingleValueMetricsProcessingDecorator(
			metricsClient,
			targetUnitID,
			dynatrace.NewMetricsProcessingThatAllowsMultipleResults(metricsClient)),
		flags)
}

func NewMetricsQueryProcessingThatAllowsOnlyOneResult(client dynatrace.ClientInterface, targetUnitID string, flags ff.GetSLIFeatureFlags) *MetricsQueryProcessing {
	metricsClient := dynatrace.NewMetricsClient(client)

	return newMetricsQueryProcessing(
		dynatrace.NewConvertUnitsAndRetryForSingleValueMetricsProcessingDecorator(
			metricsClient,
			targetUnitID,
			dynatrace.NewMetricsProcessingThatAllowsOnlyOneResult(metricsClient)),
		flags)
}

func newMetricsQueryProcessing(metricsProcessing dynatrace.MetricsProcessingInterface, flags ff.GetSLIFeatureFlags) *MetricsQueryProcessing {
	return &MetricsQueryProcessing{
		metricsProcessing: metricsProcessing,
		featureFlags:      flags,
	}
}

// Process generates SLI & SLO definitions based on the metric query and the number of dimensions in the chart definition.
func (r *MetricsQueryProcessing) Process(ctx context.Context, sloDefinition result.SLO, metricsQuery metrics.Query, timeframe common.Timeframe) []result.SLIWithSLO {
	request := dynatrace.NewMetricsClientQueryRequest(metricsQuery, timeframe)
	processingResults, err := r.metricsProcessing.ProcessRequest(ctx, request)
	if err != nil {
		return r.createTileResultsForError(sloDefinition, request, err)
	}
	return r.processResults(sloDefinition, processingResults)
}

func (r *MetricsQueryProcessing) createTileResultsForError(sloDefinition result.SLO, request dynatrace.MetricsClientQueryRequest, err error) []result.SLIWithSLO {
	var qpErrorType *dynatrace.MetricsQueryProcessingError
	if errors.As(err, &qpErrorType) {
		return []result.SLIWithSLO{result.NewWarningSLIWithSLOAndQuery(sloDefinition, request.RequestString(), err.Error())}
	}
	return []result.SLIWithSLO{result.NewFailedSLIWithSLOAndQuery(sloDefinition, request.RequestString(), err.Error())}

}

func (r *MetricsQueryProcessing) processResults(sloDefinition result.SLO, processingResults *dynatrace.MetricsProcessingResults) []result.SLIWithSLO {
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

func createSLODefinitionForName(baseSLODefinition result.SLO, name string) result.SLO {
	return result.SLO{
		SLI:         baseSLODefinition.SLI + "_" + cleanIndicatorName(name),
		DisplayName: baseSLODefinition.DisplayName + " (" + name + ")",
		Weight:      baseSLODefinition.Weight,
		KeySLI:      baseSLODefinition.KeySLI,
		Pass:        baseSLODefinition.Pass,
		Warning:     baseSLODefinition.Warning,
	}
}

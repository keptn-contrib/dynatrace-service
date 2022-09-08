package dashboard

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"golang.org/x/exp/maps"
)

type MetricsQueryProcessing struct {
	client dynatrace.ClientInterface
}

func NewMetricsQueryProcessing(client dynatrace.ClientInterface) *MetricsQueryProcessing {
	return &MetricsQueryProcessing{
		client: client,
	}
}

// Process generates SLI & SLO definitions based on the metric query and the number of dimensions in the chart definition.
func (r *MetricsQueryProcessing) Process(ctx context.Context, sloDefinition keptncommon.SLO, metricsQuery metrics.Query, timeframe common.Timeframe) []TileResult {
	request := dynatrace.NewMetricsClientQueryRequest(metricsQuery, timeframe)
	singleResult, err := dynatrace.NewMetricsClient(r.client).GetSingleMetricSeriesCollectionByQuery(ctx, request)
	if err != nil {
		var qpErrorType *dynatrace.MetricsQueryProcessingError
		if errors.As(err, &qpErrorType) {
			return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
		}
		return []TileResult{newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
	}

	return r.processMetricSeries(sloDefinition, singleResult.Data, request)
}

func (r *MetricsQueryProcessing) processMetricSeries(sloDefinition keptncommon.SLO, metricSeries []dynatrace.MetricSeries, request dynatrace.MetricsClientQueryRequest) []TileResult {
	if len(metricSeries) == 0 {
		return []TileResult{}
	}

	if len(metricSeries) == 1 {
		return []TileResult{processValues(metricSeries[0].Values, sloDefinition, request)}
	}

	var tileResults []TileResult
	for _, singleMetricSeries := range metricSeries {
		tileResults = append(tileResults, processValues(singleMetricSeries.Values, createSLODefinitionForDimensionMap(sloDefinition, singleMetricSeries.DimensionMap), request))
	}
	return tileResults
}

func processValues(values []*float64, sloDefinition keptncommon.SLO, request dynatrace.MetricsClientQueryRequest) TileResult {
	if len(values) != 1 {
		return newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), fmt.Sprintf("Expected a single value but retrieved %d", len(values)))
	}

	if values[0] == nil {
		return newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), "Expected a value but retrieved 'null'")
	}

	return newSuccessfulTileResult(sloDefinition, *values[0], request.RequestString())
}

func createSLODefinitionForDimensionMap(baseSLODefinition keptncommon.SLO, dimensionMap map[string]string) keptncommon.SLO {
	suffix := generateIndicatorSuffix(dimensionMap)
	return keptncommon.SLO{
		SLI:         baseSLODefinition.SLI + "_" + cleanIndicatorName(suffix),
		DisplayName: baseSLODefinition.DisplayName + " (" + suffix + ")",
		Weight:      baseSLODefinition.Weight,
		KeySLI:      baseSLODefinition.KeySLI,
		Pass:        baseSLODefinition.Pass,
		Warning:     baseSLODefinition.Warning,
	}
}

// generateIndicatorSuffix generates an indicator suffix based on all dimensions.
// As this is used for both indicator and display names, it must then be cleaned before use in indicator names.
func generateIndicatorSuffix(dimensionMap map[string]string) string {
	const nameSuffix = ".name"

	// take all dimension values except where both names and IDs are available, in that case only take the names
	suffixComponents := map[string]string{}
	for key, value := range dimensionMap {
		if value == "" {
			continue
		}

		if strings.HasSuffix(key, nameSuffix) {
			keyWithoutNameSuffix := strings.TrimSuffix(key, nameSuffix)
			suffixComponents[keyWithoutNameSuffix] = value
			continue
		}

		_, found := suffixComponents[key]
		if !found {
			suffixComponents[key] = value
		}
	}

	// ensure suffix component values are ordered by key alphabetically
	keys := maps.Keys(suffixComponents)
	sort.Strings(keys)
	sortedSuffixComponentValues := make([]string, 0, len(keys))
	for _, k := range keys {
		sortedSuffixComponentValues = append(sortedSuffixComponentValues, suffixComponents[k])
	}

	return strings.Join(sortedSuffixComponentValues, " ")
}

package dashboard

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
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
	singleResult, err := dynatrace.NewMetricsClient(r.client).GetSingleResultByQuery(ctx, request)
	if err != nil {
		var qpErrorType *dynatrace.MetricsQueryProcessingError
		if errors.As(err, &qpErrorType) {
			return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
		}
		return []TileResult{newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
	}

	return r.processSingleResult(sloDefinition, singleResult.Data, request)
}

func (r *MetricsQueryProcessing) processSingleResult(sloDefinition keptncommon.SLO, singleResultData []dynatrace.MetricQueryResultNumbers, request dynatrace.MetricsClientQueryRequest) []TileResult {
	var tileResults []TileResult

	for _, singleDataEntry := range singleResultData {

		indicatorName := cleanIndicatorName(sloDefinition.SLI)
		displayName := sloDefinition.DisplayName
		if len(singleResultData) > 1 {
			suffix := generateIndicatorSuffix(singleDataEntry.DimensionMap)
			indicatorName = indicatorName + "_" + cleanIndicatorName(suffix)
			displayName = displayName + " (" + suffix + ")"
		}

		tileResult := newSuccessfulTileResult(
			keptncommon.SLO{
				SLI:         indicatorName,
				DisplayName: displayName,
				Weight:      sloDefinition.Weight,
				KeySLI:      sloDefinition.KeySLI,
				Pass:        sloDefinition.Pass,
				Warning:     sloDefinition.Warning,
			},
			averageValues(singleDataEntry.Values),
			request.RequestString(),
		)

		// we got our metric, SLOs and the value
		log.WithFields(
			log.Fields{
				"tileResult": tileResult,
			}).Debug("Got indicator value")

		tileResults = append(tileResults, tileResult)
	}

	return tileResults
}

// averageValues returns the arithmetic average of the values.
func averageValues(values []float64) float64 {
	value := 0.0
	for _, singleValue := range values {
		value = value + singleValue
	}
	return value / float64(len(values))
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

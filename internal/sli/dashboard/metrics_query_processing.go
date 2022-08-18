package dashboard

import (
	"context"
	"errors"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
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
func (r *MetricsQueryProcessing) Process(ctx context.Context, noOfDimensionsInChart int, sloDefinition keptncommon.SLO, metricQueryComponents *queryComponents) []TileResult {

	// Lets run the Query and iterate through all data per dimension. Each Dimension will become its own indicator
	request := dynatrace.NewMetricsClientQueryRequest(metricQueryComponents.metricsQuery, metricQueryComponents.timeframe)
	singleResult, err := dynatrace.NewMetricsClient(r.client).GetSingleResultByQuery(ctx, request)
	if err != nil {
		var qpErrorType *dynatrace.MetricsQueryProcessingError
		if errors.As(err, &qpErrorType) {
			return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
		}
		return []TileResult{newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), err.Error())}
	}

	return r.processSingleResult(noOfDimensionsInChart, sloDefinition, metricQueryComponents, singleResult.Data, request)
}

func (r *MetricsQueryProcessing) processSingleResult(noOfDimensionsInChart int, sloDefinition keptncommon.SLO, metricQueryComponents *queryComponents, singleResultData []dynatrace.MetricQueryResultNumbers, request dynatrace.MetricsClientQueryRequest) []TileResult {
	var tileResults []TileResult
	for _, singleDataEntry := range singleResultData {
		//
		// we need to generate the indicator name based on the base name + all dimensions, e.g: teststep_MYTESTSTEP, teststep_MYOTHERTESTSTEP
		// EXCEPTION: If there is only ONE data value then we skip this and just use the base SLI name
		indicatorName := sloDefinition.SLI

		if len(singleResultData) > 1 {
			// because we use the ":names" transformation we always get two dimension entries for entity dimensions, e.g: Host, Service .... First is the Name of the entity, then the ID of the Entity
			// lets first validate that we really received Dimension Names
			dimensionCount := len(singleDataEntry.Dimensions)
			dimensionIncrement := 2
			if dimensionCount != (noOfDimensionsInChart * 2) {
				dimensionIncrement = 1
			}

			// lets iterate through the list and get all names
			for dimIx := 0; dimIx < len(singleDataEntry.Dimensions); dimIx = dimIx + dimensionIncrement {
				indicatorName = indicatorName + "_" + singleDataEntry.Dimensions[dimIx]
			}
		}

		// make sure we have a valid indicator name by getting rid of special characters
		indicatorName = cleanIndicatorName(indicatorName)

		// calculating the value
		value := 0.0
		for _, singleValue := range singleDataEntry.Values {
			value = value + singleValue
		}
		value = value / float64(len(singleDataEntry.Values))

		// we got our metric, SLOs and the value
		log.WithFields(
			log.Fields{
				"name":  indicatorName,
				"value": value,
			}).Debug("Got indicator value")

		tileResult := newSuccessfulTileResult(
			keptncommon.SLO{
				SLI:         indicatorName,
				DisplayName: sloDefinition.DisplayName,
				Weight:      sloDefinition.Weight,
				KeySLI:      sloDefinition.KeySLI,
				Pass:        sloDefinition.Pass,
				Warning:     sloDefinition.Warning,
			},
			value,
			request.RequestString(),
		)

		tileResults = append(tileResults, tileResult)
	}

	return tileResults
}

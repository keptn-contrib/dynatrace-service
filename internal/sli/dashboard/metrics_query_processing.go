package dashboard

import (
	"context"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/unit"
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
	queryResult, err := dynatrace.NewMetricsClient(r.client).GetByQuery(ctx, request)

	// ERROR-CASE: Metric API return no values or an error
	// we could not query data - so - we return the error back as part of our SLIResults
	if err != nil {
		return []TileResult{newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), "error querying Metrics API v2: "+err.Error())}
	}

	// TODO 2021-10-12: Check if having a query result with zero results is even plausable
	if len(queryResult.Result) == 0 {
		return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), "Metrics API v2 returned zero results")}
	}

	if len(queryResult.Result) > 1 {
		return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), "Metrics API v2 returned more than one result")}
	}

	// SUCCESS-CASE: we retrieved values - now create an indicator result for every dimension
	singleResult := queryResult.Result[0]
	log.WithFields(
		log.Fields{
			"metricId": singleResult.MetricID,
		}).Debug("Processing result")

	if len(singleResult.Data) == 0 {
		if len(singleResult.Warnings) > 0 {
			return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), "Metrics API v2 returned zero data points. Warnings: "+strings.Join(singleResult.Warnings, ", "))}
		}
		return []TileResult{newWarningTileResultFromSLODefinitionAndQuery(sloDefinition, request.RequestString(), "Metrics API v2 returned zero data points")}
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

		// lets scale the metric
		value = unit.ScaleData(metricQueryComponents.metricUnit, value)

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

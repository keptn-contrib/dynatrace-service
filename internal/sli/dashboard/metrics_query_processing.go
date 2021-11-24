package dashboard

import (
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/unit"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
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

// Process Generates the relevant SLIs & SLO definitions based on the metric query
// noOfDimensionsInChart: how many dimensions did we have in the chart definition
func (r *MetricsQueryProcessing) Process(noOfDimensionsInChart int, sloDefinition *keptncommon.SLO, metricQueryComponents *queryComponents) []*TileResult {

	// Lets run the Query and iterate through all data per dimension. Each Dimension will become its own indicator
	queryResult, err := dynatrace.NewMetricsClient(r.client).GetByQuery(metricQueryComponents.fullMetricQueryString)

	// ERROR-CASE: Metric API return no values or an error
	// we could not query data - so - we return the error back as part of our SLIResults
	if err != nil {
		log.WithError(err).Debug("No result for query")
		return createFailedTileResultFromSLODefinition(sloDefinition, metricQueryComponents.metricQuery, err.Error())
	}

	// TODO 2021-10-12: Check if having a query result with zero results is even plausable
	if len(queryResult.Result) == 0 {
		const errorMessage = "Expected a single result but got no result for metric ID"

		log.WithFields(
			log.Fields{
				"wantedMetricId": metricQueryComponents.metricID,
			}).Error(errorMessage)
		return createFailedTileResultFromSLODefinition(sloDefinition, metricQueryComponents.metricQuery, errorMessage)
	}

	if len(queryResult.Result) > 1 {
		const errorMessage = "Expected a result only for a single metric ID but got multiple results"
		log.WithFields(
			log.Fields{
				"wantedMetricId": metricQueryComponents.metricID,
			}).Error(errorMessage)
		return createFailedTileResultFromSLODefinition(sloDefinition, metricQueryComponents.metricQuery, errorMessage)
	}

	var tileResults []*TileResult

	// SUCCESS-CASE: we retrieved values - now create an indicator result for every dimension
	singleResult := queryResult.Result[0]
	log.WithFields(
		log.Fields{
			"metricId":                    singleResult.MetricID,
			"metricSelectorTargetSnippet": metricQueryComponents.metricSelectorTargetSnippet,
			"entitySelectorTargetSnippet": metricQueryComponents.entitySelectorTargetSnippet,
		}).Debug("Processing result")

	dataResultCount := len(singleResult.Data)
	if dataResultCount == 0 {
		return createFailedTileResultFromSLODefinition(sloDefinition, metricQueryComponents.metricQuery, "Metrics query result has no data")
	}

	for _, singleDataEntry := range singleResult.Data {
		//
		// we need to generate the indicator name based on the base name + all dimensions, e.g: teststep_MYTESTSTEP, teststep_MYOTHERTESTSTEP
		// EXCEPTION: If there is only ONE data value then we skip this and just use the base SLI name
		indicatorName := sloDefinition.SLI

		metricQueryForSLI := metricQueryComponents.metricQuery

		// we need this one to "fake" the MetricQuery for the SLi.yaml to include the dynamic dimension name for each value
		// we initialize it with ":names" as this is the part of the metric query string we will replace
		filterSLIDefinitionAggregatorValue := ":names"

		if dataResultCount > 1 {
			// because we use the ":names" transformation we always get two dimension entries for entity dimensions, e.g: Host, Service .... First is the Name of the entity, then the ID of the Entity
			// lets first validate that we really received Dimension Names
			dimensionCount := len(singleDataEntry.Dimensions)
			dimensionIncrement := 2
			if dimensionCount != (noOfDimensionsInChart * 2) {
				// ph.Logger.Debug(fmt.Sprintf("DIDNT RECEIVE ID and Names. Lets assume we just received the dimension IDs"))
				dimensionIncrement = 1
			}

			// lets iterate through the list and get all names
			for dimIx := 0; dimIx < len(singleDataEntry.Dimensions); dimIx = dimIx + dimensionIncrement {
				dimensionValue := singleDataEntry.Dimensions[dimIx]
				indicatorName = indicatorName + "_" + dimensionValue

				filterSLIDefinitionAggregatorValue = ":names" + strings.Replace(metricQueryComponents.metricSelectorTargetSnippet, "FILTERDIMENSIONVALUE", dimensionValue, 1)

				if metricQueryComponents.entitySelectorTargetSnippet != "" && dimensionIncrement == 2 {
					dimensionEntityID := singleDataEntry.Dimensions[dimIx+1]
					metricQueryForSLI = metricQueryForSLI + strings.Replace(metricQueryComponents.entitySelectorTargetSnippet, "FILTERDIMENSIONVALUE", dimensionEntityID, 1)
				}
			}
		}

		// make sure we have a valid indicator name by getting rid of special characters
		indicatorName = common.CleanIndicatorName(indicatorName)

		// calculating the value
		value := 0.0
		for _, singleValue := range singleDataEntry.Values {
			value = value + singleValue
		}
		value = value / float64(len(singleDataEntry.Values))

		// lets scale the metric
		value = unit.ScaleData(metricQueryComponents.metricID, metricQueryComponents.metricUnit, value)

		// we got our metric, slos and the value
		log.WithFields(
			log.Fields{
				"name":  indicatorName,
				"value": value,
			}).Debug("Got indicator value")

		// we use ":names" to find the right spot to add our custom dimension filter
		sliQuery := strings.Replace(metricQueryForSLI, ":names", filterSLIDefinitionAggregatorValue, 1)

		finalSLIQuery, err := unit.ConvertToMV2Query(sliQuery, metricQueryComponents.metricUnit)
		if err != nil {
			finalSLIQuery = sliQuery
		}

		// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
		// we also add the SLO definition in case we need to generate an SLO.yaml
		tileResults = append(
			tileResults,
			&TileResult{
				sliResult: &keptnv2.SLIResult{
					Metric:  indicatorName,
					Value:   value,
					Success: true,
				},
				objective: &keptncommon.SLO{
					SLI:     indicatorName,
					Weight:  sloDefinition.Weight,
					KeySLI:  sloDefinition.KeySLI,
					Pass:    sloDefinition.Pass,
					Warning: sloDefinition.Warning,
				},
				sliName:  indicatorName,
				sliQuery: finalSLIQuery,
			})
	}

	return tileResults
}

func createFailedTileResultFromSLODefinition(sloDefinition *keptncommon.SLO, sliQuery string, message string) []*TileResult {
	return []*TileResult{
		{
			sliResult: &keptnv2.SLIResult{
				Metric:  sloDefinition.SLI,
				Value:   0,
				Success: false,
				Message: message,
			},
			objective: sloDefinition,
			sliName:   sloDefinition.SLI,
			sliQuery:  sliQuery,
		},
	}
}

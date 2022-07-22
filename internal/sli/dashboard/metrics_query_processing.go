package dashboard

import (
	"context"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/unit"
	v1metrics "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/metrics"
	v1mv2 "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/mv2"
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
	queryResult, err := dynatrace.NewMetricsClient(r.client).GetByQuery(ctx, dynatrace.NewMetricsClientQueryParameters(metricQueryComponents.metricsQuery, metricQueryComponents.timeframe))

	// ERROR-CASE: Metric API return no values or an error
	// we could not query data - so - we return the error back as part of our SLIResults
	if err != nil {
		return []TileResult{newFailedTileResultFromSLODefinitionAndSLIQuery(sloDefinition, v1metrics.NewQueryProducer(metricQueryComponents.metricsQuery).Produce(), "error querying Metrics API v2: "+err.Error())}
	}

	// TODO 2021-10-12: Check if having a query result with zero results is even plausable
	if len(queryResult.Result) == 0 {
		return []TileResult{newWarningTileResultFromSLODefinitionAndSLIQuery(sloDefinition, v1metrics.NewQueryProducer(metricQueryComponents.metricsQuery).Produce(), "Metrics API v2 returned zero results")}
	}

	if len(queryResult.Result) > 1 {
		return []TileResult{newWarningTileResultFromSLODefinitionAndSLIQuery(sloDefinition, v1metrics.NewQueryProducer(metricQueryComponents.metricsQuery).Produce(), "Metrics API v2 returned more than one result")}
	}

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
		if len(singleResult.Warnings) > 0 {
			return []TileResult{newWarningTileResultFromSLODefinitionAndSLIQuery(sloDefinition, v1metrics.NewQueryProducer(metricQueryComponents.metricsQuery).Produce(), "Metrics API v2 returned zero data points. Warnings: "+strings.Join(singleResult.Warnings, ", "))}
		}
		return []TileResult{newWarningTileResultFromSLODefinitionAndSLIQuery(sloDefinition, v1metrics.NewQueryProducer(metricQueryComponents.metricsQuery).Produce(), "Metrics API v2 returned zero data points")}
	}

	var tileResults []TileResult
	for _, singleDataEntry := range singleResult.Data {
		//
		// we need to generate the indicator name based on the base name + all dimensions, e.g: teststep_MYTESTSTEP, teststep_MYOTHERTESTSTEP
		// EXCEPTION: If there is only ONE data value then we skip this and just use the base SLI name
		indicatorName := sloDefinition.SLI

		metricSelectorForSLI := metricQueryComponents.metricsQuery.GetMetricSelector()
		entitySelectorForSLI := metricQueryComponents.metricsQuery.GetEntitySelector()

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
					entitySelectorForSLI = entitySelectorForSLI + strings.Replace(metricQueryComponents.entitySelectorTargetSnippet, "FILTERDIMENSIONVALUE", dimensionEntityID, 1)
				}
			}
		}

		// we use ":names" to find the right spot to add our custom dimension filter
		metricSelectorForSLI = strings.Replace(metricSelectorForSLI, ":names", filterSLIDefinitionAggregatorValue, 1)

		metricQueryForSLI, err := metrics.NewQuery(metricSelectorForSLI, entitySelectorForSLI)
		if err != nil {
			return []TileResult{newFailedTileResultFromSLODefinitionAndSLIQuery(sloDefinition, v1metrics.NewQueryProducer(metricQueryComponents.metricsQuery).Produce(), "error creating Metrics v2 query for SLI")}
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
		value = unit.ScaleData(metricQueryComponents.metricsQuery.GetMetricSelector(), metricQueryComponents.metricUnit, value)

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
			getMetricsQueryString(metricQueryComponents.metricUnit, *metricQueryForSLI),
		)

		tileResults = append(tileResults, tileResult)
	}

	return tileResults
}

// getMetricsQueryString gets the query string for the metrics query, either MV2 or normal.
func getMetricsQueryString(unit string, query metrics.Query) string {
	mv2Query, err := v1mv2.NewQuery(unit, query)
	if err == nil {
		return v1mv2.NewQueryProducer(*mv2Query).Produce()
	}

	return v1metrics.NewQueryProducer(query).Produce()
}

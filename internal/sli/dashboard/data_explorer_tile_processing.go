package dashboard

import (
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type DataExplorerTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	timeframe     common.Timeframe
}

func NewDataExplorerTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe) *DataExplorerTileProcessing {
	return &DataExplorerTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
	}
}

func (p *DataExplorerTileProcessing) Process(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) []*TileResult {
	// first - lets figure out if this tile should be included in SLI validation or not - we parse the title and look for "sli=sliname"
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(tile.Name)
	if sloDefinition.SLI == "" {
		log.WithField("tileName", tile.Name).Debug("Data explorer tile not included as name doesnt include sli=SLINAME")
		return nil
	}

	if len(tile.Queries) != 1 {
		unsuccessfulTileResult := newUnsuccessfulTileResultFromSLODefinition(sloDefinition, "Data Explorer tile must have exactly one query")
		return []*TileResult{&unsuccessfulTileResult}
	}

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	managementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	return p.processQuery(sloDefinition, tile.Queries[0], managementZoneFilter)
}

func (p *DataExplorerTileProcessing) processQuery(sloDefinition *keptnapi.SLO, dataQuery dynatrace.DataExplorerQuery, managementZoneFilter *ManagementZoneFilter) []*TileResult {
	log.WithField("metric", dataQuery.Metric).Debug("Processing data explorer query")

	metricQuery, err := p.generateMetricQueryFromDataExplorerQuery(dataQuery, managementZoneFilter)
	if err != nil {
		log.WithError(err).Warn("generateMetricQueryFromDataExplorerQuery returned an error, SLI will not be used")
		unsuccessfulTileResult := newUnsuccessfulTileResultFromSLODefinition(sloDefinition, "Data Explorer tile could not be converted to a metric query: "+err.Error())
		return []*TileResult{&unsuccessfulTileResult}
	}

	return NewMetricsQueryProcessing(p.client).Process(len(dataQuery.SplitBy), sloDefinition, metricQuery)
}

// Looks at the DataExplorerQuery configuration of a data explorer chart and generates the Metrics Query.
//
// Returns a queryComponents object
//   - metricId, e.g: built-in:mymetric
//   - metricUnit, e.g: MilliSeconds
//   - metricQuery, e.g: metricSelector=metric&filter...
//   - fullMetricQuery, e.g: metricQuery&from=123213&to=2323
//   - entitySelectorSLIDefinition, e.g: ,entityid(FILTERDIMENSIONVALUE)
//   - filterSLIDefinitionAggregator, e.g: , filter(eq(Test Step,FILTERDIMENSIONVALUE))
func (p *DataExplorerTileProcessing) generateMetricQueryFromDataExplorerQuery(dataQuery dynatrace.DataExplorerQuery, managementZoneFilter *ManagementZoneFilter) (*queryComponents, error) {

	// TODO 2021-08-04: there are too many return values and they are have the same type

	if dataQuery.Metric == "" {
		return nil, fmt.Errorf("Metric query generation requires that data explorer query has a metric")
	}

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := dynatrace.NewMetricsClient(p.client).GetByID(dataQuery.Metric)
	if err != nil {
		return nil, err
	}

	processedFilter := &processedFilterComponents{}

	// Create the right entity Selectors for the queries execute
	if dataQuery.FilterBy != nil && len(dataQuery.FilterBy.NestedFilters) > 0 {

		// TODO: 2021-10-29: consider supporting more than a single filter with a single criterion
		if len(dataQuery.FilterBy.NestedFilters) != 1 {
			return nil, fmt.Errorf("only a single filter is supported")
		}

		if len(dataQuery.FilterBy.NestedFilters[0].Criteria) != 1 {
			return nil, fmt.Errorf("only a single filter criterion is supported")
		}

		if len(dataQuery.FilterBy.NestedFilters[0].NestedFilters) > 0 {
			return nil, fmt.Errorf("nested filters are not permitted")
		}

		entityType := strings.ToUpper(strings.TrimPrefix(dataQuery.FilterBy.NestedFilters[0].Filter, "dt.entity."))
		if len(metricDefinition.EntityType) > 0 {
			entityType = metricDefinition.EntityType[0]
		}

		processedFilter, err = processFilter(entityType, &dataQuery.FilterBy.NestedFilters[0])
		if err != nil {
			return nil, err
		}
	}

	// optionally split by a single dimension
	// TODO: 2021-10-29: consider adding support for more than one split dimension
	var splitBy string
	if len(dataQuery.SplitBy) > 1 {
		return nil, fmt.Errorf("only a single splitBy dimension is supported")
	} else if len(dataQuery.SplitBy) == 1 {
		splitBy = fmt.Sprintf(":splitBy(\"%s\")", dataQuery.SplitBy[0])
		// if we split by a dimension we need to include that dimension in our individual SLI query definitions - thats why we hand this back in the filter clause
		processedFilter.metricSelectorTargetSnippet = fmt.Sprintf("%s:filter(eq(%s,FILTERDIMENSIONVALUE))", processedFilter.metricSelectorTargetSnippet, dataQuery.SplitBy[0])
	} else {
		splitBy = ":splitBy()"
	}

	metricAggregation, err := getSpaceAggregationTransformation(dataQuery.SpaceAggregation)
	if err != nil {
		return nil, err
	}

	// optionally add management zone filter to entity selector filter
	managementZoneFilterString := managementZoneFilter.ForEntitySelector()
	if managementZoneFilterString != "" {
		if processedFilter.entitySelectorFilter == "" {
			processedFilter.entitySelectorFilter = fmt.Sprintf("type(%s)", metricDefinition.EntityType[0])
		}
		processedFilter.entitySelectorFilter = processedFilter.entitySelectorFilter + managementZoneFilterString
	}

	// NOTE: add :names so we also get the names of the dimensions and not just the entities. This means we get two values for each dimension
	metricSelector := fmt.Sprintf("%s%s%s:%s:names",
		dataQuery.Metric, processedFilter.metricSelectorFilter, splitBy, strings.ToLower(metricAggregation))

	metricsQuery, err := metrics.NewQuery(metricSelector, processedFilter.entitySelectorFilter)
	if err != nil {
		return nil, err
	}

	return &queryComponents{
		metricsQuery:                *metricsQuery,
		timeframe:                   p.timeframe,
		metricUnit:                  metricDefinition.Unit,
		entitySelectorTargetSnippet: processedFilter.entitySelectorTargetSnippet,
		metricSelectorTargetSnippet: processedFilter.metricSelectorTargetSnippet,
	}, nil

}

type processedFilterComponents struct {
	metricSelectorFilter        string
	metricSelectorTargetSnippet string
	entitySelectorFilter        string
	entitySelectorTargetSnippet string
}

func processFilter(entityType string, filter *dynatrace.DataExplorerFilter) (*processedFilterComponents, error) {
	switch filter.FilterType {
	case "ID":
		return &processedFilterComponents{
			entitySelectorFilter:        fmt.Sprintf("entityId(%s)", filter.Criteria[0].Value),
			entitySelectorTargetSnippet: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	case "NAME":
		return &processedFilterComponents{
			entitySelectorFilter:        fmt.Sprintf("type(%s),entityName(\"%s\")", entityType, filter.Criteria[0].Value),
			entitySelectorTargetSnippet: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	case "TAG":
		return &processedFilterComponents{
			entitySelectorFilter:        fmt.Sprintf("type(%s),tag(\"%s\")", entityType, filter.Criteria[0].Value),
			entitySelectorTargetSnippet: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	case "ENTITY_ATTRIBUTE":
		return &processedFilterComponents{
			entitySelectorFilter:        fmt.Sprintf("type(%s),%s(\"%s\")", entityType, filter.EntityAttribute, filter.Criteria[0].Value),
			entitySelectorTargetSnippet: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	case "DIMENSION":
		return &processedFilterComponents{
			metricSelectorFilter:        fmt.Sprintf(":filter(%s(\"%s\",\"%s\"))", filter.Criteria[0].Evaluator, filter.Filter, filter.Criteria[0].Value),
			metricSelectorTargetSnippet: fmt.Sprintf(":filter(eq(\"%s\",\"FILTERDIMENSIONVALUE\"))", filter.Filter),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported filter type: %s", filter.FilterType)
	}
}

func getSpaceAggregationTransformation(spaceAggregation string) (string, error) {
	switch spaceAggregation {
	case "AVG":
		return "avg", nil
	case "SUM":
		return "sum", nil
	case "MIN":
		return "min", nil
	case "MAX":
		return "max", nil
	case "COUNT":
		return "count", nil
	case "MEDIAN":
		return "median", nil
	case "PERCENTILE_10":
		return "percentile(10)", nil
	case "PERCENTILE_75":
		return "percentile(75)", nil
	case "PERCENTILE_90":
		return "percentile(90)", nil
	case "VALUE":
		return "value", nil
	default:
		return "", fmt.Errorf("unknown space aggregation: %s", spaceAggregation)
	}
}

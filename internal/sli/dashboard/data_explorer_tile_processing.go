package dashboard

import (
	"fmt"
	"strings"
	"time"

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
	startUnix     time.Time
	endUnix       time.Time
}

func NewDataExplorerTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, startUnix time.Time, endUnix time.Time) *DataExplorerTileProcessing {
	return &DataExplorerTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		startUnix:     startUnix,
		endUnix:       endUnix,
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
		return createFailureTileResult(sloDefinition.SLI, "", "Data Explorer tile must have exactly one query")
	}

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	return p.processQuery(sloDefinition, tile.Queries[0], tileManagementZoneFilter)
}

func (p *DataExplorerTileProcessing) processQuery(sloDefinition *keptnapi.SLO, dataQuery dynatrace.DataExplorerQuery, tileManagementZoneFilter *ManagementZoneFilter) []*TileResult {
	log.WithField("metric", dataQuery.Metric).Debug("Processing data explorer query")

	metricQuery, err := p.generateMetricQueryFromDataExplorerQuery(dataQuery, tileManagementZoneFilter)
	if err != nil {
		log.WithError(err).Warn("generateMetricQueryFromDataExplorerQuery returned an error, SLI will not be used")
		return createFailureTileResult(sloDefinition.SLI, "", "Data Explorer tile could not be converted to a metric query: "+err.Error())
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
func (p *DataExplorerTileProcessing) generateMetricQueryFromDataExplorerQuery(dataQuery dynatrace.DataExplorerQuery, tileManagementZoneFilter *ManagementZoneFilter) (*queryComponents, error) {

	// TODO 2021-08-04: there are too many return values and they are have the same type

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := dynatrace.NewMetricsClient(p.client).GetByID(dataQuery.Metric)
	if err != nil {
		return nil, err
	}

	// building the merge aggregator string, e.g: merge("dt.entity.disk"):merge("dt.entity.host") - or merge("dt.entity.service")
	// TODO: 2021-09-20: Check for redundant code after update to use dimension keys rather than indexes
	metricDimensionCount := len(metricDefinition.DimensionDefinitions)

	// we need to merge all those dimensions based on the metric definition that are not included in the "splitBy"
	// so - we iterate through the dimensions based on the metric definition from the back to front - and then merge those not included in splitBy
	mergeAggregator := ""
	for metricDimIx := metricDimensionCount - 1; metricDimIx >= 0; metricDimIx-- {
		log.WithField("metricDimIx", metricDimIx).Debug("Processing Dimension Ix")

		doMergeDimension := true
		for _, splitDimension := range dataQuery.SplitBy {
			log.WithFields(
				log.Fields{
					"dimension1": splitDimension,
					"dimension2": metricDefinition.DimensionDefinitions[metricDimIx].Key,
				}).Debug("Comparing Dimensions %")

			if strings.Compare(splitDimension, metricDefinition.DimensionDefinitions[metricDimIx].Key) == 0 {
				doMergeDimension = false
			}
		}

		if doMergeDimension {
			// this is a dimension we want to merge as it is not split by in the chart
			log.WithField("dimension", metricDefinition.DimensionDefinitions[metricDimIx].Key).Debug("merging dimension")
			mergeAggregator = mergeAggregator + fmt.Sprintf(":merge(\"%s\")", metricDefinition.DimensionDefinitions[metricDimIx].Key)
		}
	}

	filterAggregator := &filterAggregator{}

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

		filterAggregator, err = makeFilter(entityType, &dataQuery.FilterBy.NestedFilters[0])
		if err != nil {
			return nil, err
		}
	}

	// optionally split by a single dimentision
	// TODO: 2021-10-29: consider adding support for more than one split dimension
	if len(dataQuery.SplitBy) > 1 {
		return nil, fmt.Errorf("only a single splitBy dimension is supported")
	}

	// if we split by a dimension we need to include that dimension in our individual SLI query definitions - thats why we hand this back in the filter clause
	if len(dataQuery.SplitBy) == 1 {
		filterAggregator.filterSLIDefinitionAggregator = fmt.Sprintf("%s:filter(eq(%s,FILTERDIMENSIONVALUE))", filterAggregator.filterSLIDefinitionAggregator, dataQuery.SplitBy[0])
	}

	metricAggregation, err := getSpaceAggregationTransformation(dataQuery.SpaceAggregation)
	if err != nil {
		metricAggregation = metricDefinition.DefaultAggregation.Type
	}

	// lets create the metricSelector and entitySelector
	// ATTENTION: adding :names so we also get the names of the dimensions and not just the entities. This means we get two values for each dimension
	metricQuery := fmt.Sprintf("metricSelector=%s%s%s:%s:names%s%s",
		dataQuery.Metric, mergeAggregator, filterAggregator.filterAggregator, strings.ToLower(metricAggregation),
		filterAggregator.entityFilter, tileManagementZoneFilter.ForEntitySelector())

	// lets build the Dynatrace API Metric query for the proposed timeframe and additonal filters!
	fullMetricQuery, metricID, err := metrics.NewQueryBuilder(p.eventData, p.customFilters).Build(metricQuery, p.startUnix, p.endUnix)
	if err != nil {
		return nil, err
	}

	return &queryComponents{
		metricID:                      metricID,
		metricUnit:                    metricDefinition.Unit,
		metricQuery:                   metricQuery,
		fullMetricQueryString:         fullMetricQuery,
		entitySelectorSLIDefinition:   filterAggregator.entitySelectorSLIDefinition,
		filterSLIDefinitionAggregator: filterAggregator.filterSLIDefinitionAggregator,
	}, nil

}

type filterAggregator struct {
	filterAggregator              string
	filterSLIDefinitionAggregator string
	entityFilter                  string
	entitySelectorSLIDefinition   string
}

// TODO: 2021-11-09: Investigate adding support for other filter types, e.g. DIMENSION
func makeFilter(entityType string, nestedFilter *dynatrace.DataExplorerFilter) (*filterAggregator, error) {
	switch nestedFilter.FilterType {
	case "ID":
		return &filterAggregator{
			entityFilter:                fmt.Sprintf("&entitySelector=entityId(%s)", nestedFilter.Criteria[0].Value),
			entitySelectorSLIDefinition: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	case "NAME":
		return &filterAggregator{
			entityFilter:                fmt.Sprintf("&entitySelector=type(%s),entityName(\"%s\")", entityType, nestedFilter.Criteria[0].Value),
			entitySelectorSLIDefinition: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	case "TAG":
		return &filterAggregator{
			entityFilter:                fmt.Sprintf("&entitySelector=type(%s),tag(\"%s\")", entityType, nestedFilter.Criteria[0].Value),
			entitySelectorSLIDefinition: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	case "ENTITY_ATTRIBUTE":
		return &filterAggregator{
			entityFilter:                fmt.Sprintf("&entitySelector=type(%s),%s(\"%s\")", entityType, nestedFilter.EntityAttribute, nestedFilter.Criteria[0].Value),
			entitySelectorSLIDefinition: ",entityId(FILTERDIMENSIONVALUE)",
		}, nil

	default:
		return nil, fmt.Errorf("unsupported filter type")
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

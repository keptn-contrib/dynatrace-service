package dashboard

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type CustomChartingTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	startUnix     time.Time
	endUnix       time.Time
}

func NewCustomChartingTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, startUnix time.Time, endUnix time.Time) *CustomChartingTileProcessing {
	return &CustomChartingTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		startUnix:     startUnix,
		endUnix:       endUnix,
	}
}

func (p *CustomChartingTileProcessing) Process(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) []*TileResult {
	tileTitle := tile.Title()

	// first - lets figure out if this tile should be included in SLI validation or not - we parse the title and look for "sli=sliname"
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(tileTitle)
	if sloDefinition.SLI == "" {
		log.WithField("tileTitle", tileTitle).Debug("Tile not included as name doesnt include sli=SLINAME")
		return nil
	}

	log.WithFields(
		log.Fields{
			"tileTitle":         tileTitle,
			"baseIndicatorName": sloDefinition.SLI,
		}).Debug("Processing custom chart")

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	if tile.FilterConfig == nil {
		return nil
	}

	var tileResults []*TileResult

	// we can potentially have multiple series on that chart
	for _, series := range tile.FilterConfig.ChartConfig.Series {

		// First lets generate the query and extract all important metric information we need for generating SLIs & SLOs
		metricQuery, err := p.generateMetricQueryFromChart(series, tileManagementZoneFilter, tile.FilterConfig.FiltersPerEntityType, p.startUnix, p.endUnix)

		// if there was no error we generate the SLO & SLO definition
		if err != nil {
			log.WithError(err).Warn("generateMetricQueryFromChart returned an error, SLI will not be used")
			continue
		}

		results := NewMetricsQueryProcessing(p.client).Process(len(series.Dimensions), sloDefinition, metricQuery)
		tileResults = append(tileResults, results...)
	}

	return tileResults
}

// Looks at the ChartSeries configuration of a regular chart and generates the Metrics Query
//
// Returns a queryComponents object
//   - metricId, e.g: built-in:mymetric
//   - metricUnit, e.g: MilliSeconds
//   - metricQuery, e.g: metricSelector=metric&filter...
//   - fullMetricQuery, e.g: metricQuery&from=123213&to=2323
//   - entitySelectorSLIDefinition, e.g: ,entityid(FILTERDIMENSIONVALUE)
//   - filterSLIDefinitionAggregator, e.g: , filter(eq(Test Step,FILTERDIMENSIONVALUE))
func (p *CustomChartingTileProcessing) generateMetricQueryFromChart(series dynatrace.Series, tileManagementZoneFilter *ManagementZoneFilter, filtersPerEntityType map[string]map[string][]string, startUnix time.Time, endUnix time.Time) (*queryComponents, error) {

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := dynatrace.NewMetricsClient(p.client).GetByID(series.Metric)
	if err != nil {
		log.WithError(err).WithField("metric", series.Metric).Debug("Error retrieving metric description")
		return nil, err
	}

	// building the merge aggregator string, e.g: merge("dt.entity.disk"):merge("dt.entity.host") - or merge("dt.entity.service")
	// TODO: 2021-09-20: Check for redundant code after update to use dimension keys rather than indexes
	metricDimensionCount := len(metricDefinition.DimensionDefinitions)
	metricAggregation := metricDefinition.DefaultAggregation.Type
	mergeAggregator := ""
	filterAggregator := ""
	filterSLIDefinitionAggregator := ""
	entitySelectorSLIDefinition := ""

	// now we need to merge all the dimensions that are not part of the series.dimensions, e.g: if the metric has two dimensions but only one dimension is used in the chart we need to merge the others
	// as multiple-merges are possible but as they are executed in sequence we have to use the right index
	for metricDimIx := metricDimensionCount - 1; metricDimIx >= 0; metricDimIx-- {
		doMergeDimension := true
		metricDimIxAsString := strconv.Itoa(metricDimIx)
		// lets check if this dimension is in the chart
		for _, seriesDim := range series.Dimensions {
			log.WithFields(
				log.Fields{
					"seriesDim.id": seriesDim.ID,
					"metricDimIx":  metricDimIxAsString,
				}).Debug("check")
			if strings.Compare(seriesDim.ID, metricDimIxAsString) == 0 {
				// this is a dimension we want to keep and not merge
				log.WithField("dimension", metricDefinition.DimensionDefinitions[metricDimIx].Name).Debug("not merging dimension")
				doMergeDimension = false

				// lets check if we need to apply a dimension filter
				// TODO: support multiple filters - right now we only support 1
				if len(seriesDim.Values) > 0 {
					filterAggregator = fmt.Sprintf(":filter(eq(%s,%s))", seriesDim.Name, seriesDim.Values[0])
				} else {
					// we need this for the generation of the SLI for each individual dimension value
					// if the dimension is a dt.entity we have to add an addiotnal entityId to the entitySelector - otherwise we add a filter for the dimension
					if strings.HasPrefix(seriesDim.Name, "dt.entity.") {
						entitySelectorSLIDefinition = fmt.Sprintf(",entityId(FILTERDIMENSIONVALUE)")
					} else {
						filterSLIDefinitionAggregator = fmt.Sprintf(":filter(eq(%s,FILTERDIMENSIONVALUE))", seriesDim.Name)
					}
				}
			}
		}

		if doMergeDimension {
			// this is a dimension we want to merge as it is not split by in the chart
			log.WithField("dimension", metricDefinition.DimensionDefinitions[metricDimIx].Name).Debug("merging dimension")
			mergeAggregator = mergeAggregator + fmt.Sprintf(":merge(\"%s\")", metricDefinition.DimensionDefinitions[metricDimIx].Key)
		}
	}

	// handle aggregation. If "NONE" is specified we go to the defaultAggregration
	if series.Aggregation != "NONE" {
		metricAggregation = series.Aggregation
	}
	// for percentile we need to specify the percentile itself
	if metricAggregation == "PERCENTILE" {
		metricAggregation = fmt.Sprintf("%s(%f)", metricAggregation, series.Percentile)
	}
	// for rate measures such as failure rate we take average if it is "OF_INTEREST_RATIO"
	if metricAggregation == "OF_INTEREST_RATIO" {
		metricAggregation = "avg"
	}
	// for rate measures charting also provides the "OTHER_RATIO" option which is the inverse
	// TODO: not supported via API - so we default to avg
	if metricAggregation == "OTHER_RATIO" {
		metricAggregation = "avg"
	}

	// TODO - handle aggregation rates -> probably doesnt make sense as we always evalute a short timeframe
	// if series.AggregationRate

	// lets get the true entity type as the one in the dashboard might not be accurate, e.g: IOT might be used instead of CUSTOM_DEVICE
	// so - if the metric definition has EntityTypes defined we take the first one
	entityType := series.EntityType
	if len(metricDefinition.EntityType) > 0 {
		entityType = metricDefinition.EntityType[0]
	}

	// Need to implement chart filters per entity type, e.g: its possible that a chart has a filter on entites or tags
	// lets see if we have a FiltersPerEntityType for the tiles EntityType
	entityTileFilter := getEntitySelectorFromEntityFilter(filtersPerEntityType, entityType)

	// lets create the metricSelector and entitySelector
	// ATTENTION: adding :names so we also get the names of the dimensions and not just the entities. This means we get two values for each dimension
	metricQuery := fmt.Sprintf("metricSelector=%s%s%s:%s:names&entitySelector=type(%s)%s%s",
		series.Metric, mergeAggregator, filterAggregator, strings.ToLower(metricAggregation),
		entityType, entityTileFilter, tileManagementZoneFilter.ForEntitySelector())

	// lets build the Dynatrace API Metric query for the proposed timeframe and additional filters!
	fullMetricQuery, metricID, err := metrics.NewQueryBuilder(p.eventData, p.customFilters).Build(metricQuery, startUnix, endUnix)
	if err != nil {
		return nil, err
	}

	return &queryComponents{
		metricID:                      metricID,
		metricUnit:                    metricDefinition.Unit,
		metricQuery:                   metricQuery,
		fullMetricQueryString:         fullMetricQuery,
		entitySelectorSLIDefinition:   entitySelectorSLIDefinition,
		filterSLIDefinitionAggregator: filterSLIDefinitionAggregator,
	}, nil
}

// getEntitySelectorFromEntityFilter Parses the filtersPerEntityType dashboard definition and returns the entitySelector query filter -
// the return value always starts with a , (comma)
//   return example: ,entityId("ABAD-222121321321")
func getEntitySelectorFromEntityFilter(filtersPerEntityType map[string]map[string][]string, entityType string) string {
	entityTileFilter := ""
	if filtersPerEntityType, containsEntityType := filtersPerEntityType[entityType]; containsEntityType {
		// Check for SPECIFIC_ENTITIES - if we have an array then we filter for each entity
		if entityArray, containsSpecificEntities := filtersPerEntityType["SPECIFIC_ENTITIES"]; containsSpecificEntities {
			for _, entityId := range entityArray {
				entityTileFilter = entityTileFilter + ","
				entityTileFilter = entityTileFilter + fmt.Sprintf("entityId(\"%s\")", entityId)
			}
		}
		// Check for SPECIFIC_ENTITIES - if we have an array then we filter for each entity
		if tagArray, containsAutoTags := filtersPerEntityType["AUTO_TAGS"]; containsAutoTags {
			for _, tag := range tagArray {
				entityTileFilter = entityTileFilter + ","
				entityTileFilter = entityTileFilter + fmt.Sprintf("tag(\"%s\")", tag)
			}
		}
	}
	return entityTileFilter
}

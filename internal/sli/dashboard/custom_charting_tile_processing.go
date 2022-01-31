package dashboard

import (
	"errors"
	"fmt"
	"sort"
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
		return createFailedTileResultFromSLODefinition(sloDefinition, "Custom charting tile is missing a filterConfig element")
	}

	if len(tile.FilterConfig.ChartConfig.Series) != 1 {
		return createFailedTileResultFromSLODefinition(sloDefinition, "Custom charting tile must have exactly one series")
	}

	return p.processSeries(sloDefinition, &tile.FilterConfig.ChartConfig.Series[0], tileManagementZoneFilter, tile.FilterConfig.FiltersPerEntityType)
}

func (p *CustomChartingTileProcessing) processSeries(sloDefinition *keptnapi.SLO, series *dynatrace.Series, tileManagementZoneFilter *ManagementZoneFilter, filtersPerEntityType map[string]dynatrace.FilterMap) []*TileResult {

	metricQuery, err := p.generateMetricQueryFromChart(series, tileManagementZoneFilter, filtersPerEntityType, p.startUnix, p.endUnix)

	if err != nil {
		log.WithError(err).Warn("generateMetricQueryFromChart returned an error, SLI will not be used")
		return createFailedTileResultFromSLODefinition(sloDefinition, "Custom charting tile could not be converted to a metric query: "+err.Error())
	}

	return NewMetricsQueryProcessing(p.client).Process(len(series.Dimensions), sloDefinition, metricQuery)
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
func (p *CustomChartingTileProcessing) generateMetricQueryFromChart(series *dynatrace.Series, tileManagementZoneFilter *ManagementZoneFilter, filtersPerEntityType map[string]dynatrace.FilterMap, startUnix time.Time, endUnix time.Time) (*queryComponents, error) {

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := dynatrace.NewMetricsClient(p.client).GetByID(series.Metric)
	if err != nil {
		log.WithError(err).WithField("metric", series.Metric).Debug("Error retrieving metric description")
		return nil, err
	}

	// handle aggregation. If "NONE" is specified we go to the defaultAggregration
	metricAggregation := metricDefinition.DefaultAggregation.Type
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

	// Need to implement chart filters per entity type, e.g: its possible that a chart has a filter on entites or tags
	// lets see if we have a FiltersPerEntityType for the tiles EntityType
	entityTileFilter, err := getEntitySelectorFromEntityFilter(filtersPerEntityType, series.EntityType)
	if err != nil {
		return nil, fmt.Errorf("could not get filter for entity type %s: %w", series.EntityType, err)
	}

	// lets get the true entity type as the one in the dashboard might not be accurate, e.g: IOT might be used instead of CUSTOM_DEVICE
	// so - if the metric definition has EntityTypes defined we take the first one
	entityType := series.EntityType
	if len(metricDefinition.EntityType) > 0 {
		entityType = metricDefinition.EntityType[0]
	}

	// build split by
	splitBy := ""
	filterAggregator := ""
	metricSelectorTargetSnippet := ""
	entitySelectorTargetSnippet := ""
	if len(series.Dimensions) > 1 {
		return nil, errors.New("only a single dimension is supported")
	} else if len(series.Dimensions) == 1 {
		seriesDim := series.Dimensions[0]
		splitBy = fmt.Sprintf(":splitBy(\"%s\")", seriesDim.Name)

		// lets check if we need to apply a dimension filter
		// TODO: support multiple filters - right now we only support 1
		if len(seriesDim.Values) > 1 {
			return nil, errors.New("only a single dimension filter is supported")
		} else if len(seriesDim.Values) == 1 {
			filterAggregator = fmt.Sprintf(":filter(eq(%s,%s))", seriesDim.Name, seriesDim.Values[0])
		} else {
			// we need this for the generation of the SLI for each individual dimension value
			// if the dimension is a dt.entity we have to add an additional entityId to the entitySelector - otherwise we add a filter for the dimension
			if strings.HasPrefix(seriesDim.Name, "dt.entity.") {
				entitySelectorTargetSnippet = fmt.Sprintf(",entityId(\"FILTERDIMENSIONVALUE\")")
			} else {
				metricSelectorTargetSnippet = fmt.Sprintf(":filter(eq(%s,FILTERDIMENSIONVALUE))", seriesDim.Name)
			}
		}
	} else {
		splitBy = ":splitBy()"
	}

	// NOTE: add :names so we also get the names of the dimensions and not just the entities. This means we get two values for each dimension
	metricSelector := fmt.Sprintf("%s%s%s:%s:names",
		series.Metric, filterAggregator, splitBy, strings.ToLower(metricAggregation))
	entitySelector := fmt.Sprintf("type(%s)%s%s",
		entityType, entityTileFilter, tileManagementZoneFilter.ForEntitySelector())
	metricsQuery, err := metrics.NewQuery(metricSelector, entitySelector)
	if err != nil {
		return nil, err
	}

	return &queryComponents{
		metricsQuery:                *metricsQuery,
		startTime:                   startUnix,
		endTime:                     endUnix,
		metricUnit:                  metricDefinition.Unit,
		entitySelectorTargetSnippet: entitySelectorTargetSnippet,
		metricSelectorTargetSnippet: metricSelectorTargetSnippet,
	}, nil
}

// getEntitySelectorFromEntityFilter Parses the filtersPerEntityType dashboard definition and returns the entitySelector query filter -
// the return value always starts with a , (comma)
//   return example: ,entityId("ABAD-222121321321")
func getEntitySelectorFromEntityFilter(filtersPerEntityType map[string]dynatrace.FilterMap, entityType string) (string, error) {
	filterMap, containsEntityType := filtersPerEntityType[entityType]
	if !containsEntityType {
		return "", nil
	}

	filter, err := makeEntitySelectorForFilterMap(filterMap)
	if err != nil {
		return "", err
	}

	if entityType == "SERVICE_KEY_REQUEST" {
		filter = ",fromRelationships.isServiceMethodOfService(type(SERVICE)" + filter + ")"
	}
	return filter, nil
}

func makeEntitySelectorForFilterMap(filterMap dynatrace.FilterMap) (string, error) {
	unknownFilters := []string{}
	for k := range filterMap {
		switch k {
		case "SPECIFIC_ENTITIES", "AUTO_TAGS":
			// do nothing - these are fine and will be used later

		default:
			unknownFilters = append(unknownFilters, k)
		}
	}

	if len(unknownFilters) > 0 {
		sort.Strings(unknownFilters)
		return "", fmt.Errorf("unknown filters: %s", strings.Join(unknownFilters, ", "))
	}

	return makeSpecificEntitiesFilter(filterMap["SPECIFIC_ENTITIES"]) + makeAutoTagsFilter(filterMap["AUTO_TAGS"]), nil
}

func makeSpecificEntitiesFilter(specificEntities []string) string {
	specificEntityFilter := ""
	for _, entityId := range specificEntities {
		specificEntityFilter = specificEntityFilter + fmt.Sprintf(",entityId(\"%s\")", entityId)
	}
	return specificEntityFilter
}

func makeAutoTagsFilter(autoTags []string) string {
	autoTagsFilter := ""
	for _, tag := range autoTags {
		autoTagsFilter = autoTagsFilter + fmt.Sprintf(",tag(\"%s\")", tag)
	}
	return autoTagsFilter
}

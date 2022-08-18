package dashboard

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
)

// CustomChartingTileProcessing represents the processing of a Custom Charting dashboard tile.
type CustomChartingTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	timeframe     common.Timeframe
}

// NewCustomChartingTileProcessing creates a new CustomChartingTileProcessing.
func NewCustomChartingTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe) *CustomChartingTileProcessing {
	return &CustomChartingTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
	}
}

// Process processes the specified Custom Charting dashboard tile.
func (p *CustomChartingTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) []TileResult {
	if tile.FilterConfig == nil {
		log.Debug("Skipping custom charting tile as it is missing a filterConfig element")
		return nil
	}

	sloDefinitionParsingResult, err := parseSLODefinition(tile.FilterConfig.CustomName)
	var sloDefError *sloDefinitionError
	if errors.As(err, &sloDefError) {
		return []TileResult{newFailedTileResultFromError(sloDefError.sliNameOrTileTitle(), "Custom charting tile title parsing error", err)}
	}

	if sloDefinitionParsingResult.exclude {
		log.WithField("tile.FilterConfig.CustomName", tile.FilterConfig.CustomName).Debug("Tile excluded as name includes exclude=true")
		return nil
	}

	sloDefinition := sloDefinitionParsingResult.sloDefinition
	if sloDefinition.SLI == "" {
		log.WithField("tile.FilterConfig.CustomName", tile.FilterConfig.CustomName).Debug("Tile not included as name doesnt include sli=SLINAME")
		return nil
	}

	log.WithFields(
		log.Fields{
			"tile.FilterConfig.CustomName": tile.FilterConfig.CustomName,
			"baseIndicatorName":            sloDefinition.SLI,
		}).Debug("Processing custom chart")

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	if len(tile.FilterConfig.ChartConfig.Series) != 1 {
		return []TileResult{newFailedTileResultFromSLODefinition(sloDefinition, "Custom charting tile must have exactly one series")}
	}

	return p.processSeries(ctx, sloDefinition, &tile.FilterConfig.ChartConfig.Series[0], tileManagementZoneFilter, tile.FilterConfig.FiltersPerEntityType)
}

func (p *CustomChartingTileProcessing) processSeries(ctx context.Context, sloDefinition keptnapi.SLO, series *dynatrace.Series, tileManagementZoneFilter *ManagementZoneFilter, filtersPerEntityType map[string]dynatrace.FilterMap) []TileResult {

	metricQuery, err := p.generateMetricQueryFromChartSeries(ctx, series, tileManagementZoneFilter, filtersPerEntityType)

	if err != nil {
		log.WithError(err).Warn("generateMetricQueryFromChart returned an error, SLI will not be used")
		return []TileResult{newFailedTileResultFromSLODefinition(sloDefinition, "Custom charting tile could not be converted to a metric query: "+err.Error())}
	}

	return NewMetricsQueryProcessing(p.client).Process(ctx, sloDefinition, metricQuery)
}

func (p *CustomChartingTileProcessing) generateMetricQueryFromChartSeries(ctx context.Context, series *dynatrace.Series, tileManagementZoneFilter *ManagementZoneFilter, filtersPerEntityType map[string]dynatrace.FilterMap) (*queryComponents, error) {

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := dynatrace.NewMetricsClient(p.client).GetByID(ctx, series.Metric)
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
	if len(series.Dimensions) > 1 {
		return nil, errors.New("only a single dimension is supported")
	} else if len(series.Dimensions) == 1 {
		seriesDim := series.Dimensions[0]
		splitBy = fmt.Sprintf(":splitBy(\"%s\")", seriesDim.Name)

		// lets check if we need to apply a dimension filter
		// TODO: support multiple filters - right now we only support 1
		if len(seriesDim.Values) > 1 {
			return nil, errors.New("only a single dimension filter is supported")
		}

		if len(seriesDim.Values) == 1 {
			filterAggregator = fmt.Sprintf(":filter(eq(%s,%s))", seriesDim.Name, seriesDim.Values[0])
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
		metricsQuery: *metricsQuery,
		timeframe:    p.timeframe,
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

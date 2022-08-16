package dashboard

import (
	"context"
	"errors"
	"fmt"
	"strings"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
)

// DataExplorerTileProcessing represents the processing of a Data Explorer dashboard tile.
type DataExplorerTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	timeframe     common.Timeframe
}

// NewDataExplorerTileProcessing creates a new DataExplorerTileProcessing.
func NewDataExplorerTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe) *DataExplorerTileProcessing {
	return &DataExplorerTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
	}
}

// Process processes the specified Data Explorer dashboard tile.
func (p *DataExplorerTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) []TileResult {
	sloDefinitionParsingResult, err := parseSLODefinition(tile.Name)
	var sloDefError *sloDefinitionError
	if errors.As(err, &sloDefError) {
		return []TileResult{newFailedTileResultFromError(sloDefError.sliNameOrTileTitle(), "Data Explorer tile title parsing error", err)}
	}

	if sloDefinitionParsingResult.exclude {
		log.WithField("tileName", tile.Name).Debug("Tile excluded as name includes exclude=true")
		return nil
	}

	sloDefinition := sloDefinitionParsingResult.sloDefinition
	if sloDefinition.SLI == "" {
		log.WithField("tileName", tile.Name).Debug("Omitted Data Explorer tile as no SLI name could be derived")
		return nil
	}

	if (len(sloDefinition.Pass) == 0) && (len(sloDefinition.Warning) == 0) {
		criteria, err := tryGetThresholdPassAndWarningCriteria(tile)
		if err != nil {
			return []TileResult{newFailedTileResultFromSLODefinition(sloDefinition, fmt.Sprintf("Invalid Data Explorer tile thresholds: %s", err.Error()))}
		}

		if criteria != nil {
			sloDefinition.Pass = []*keptnapi.SLOCriteria{&criteria.pass}
			sloDefinition.Warning = []*keptnapi.SLOCriteria{&criteria.warning}
		}
	}

	err = validateDataExplorerTile(tile)
	if err != nil {
		return []TileResult{newFailedTileResultFromSLODefinition(sloDefinition, err.Error())}
	}

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	managementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	return p.processQuery(ctx, sloDefinition, tile.Queries[0], managementZoneFilter)
}

func validateDataExplorerTile(tile *dynatrace.Tile) error {
	if len(tile.Queries) != 1 {
		return fmt.Errorf("Data Explorer tile must have exactly one query")
	}

	if tile.VisualConfig == nil {
		return nil
	}

	if len(tile.VisualConfig.Rules) == 0 {
		return nil
	}

	if len(tile.VisualConfig.Rules) > 1 {
		return fmt.Errorf("Data Explorer tile must have exactly one visual configuration rule")
	}

	return validateDataExplorerVisualConfigurationRule(tile.VisualConfig.Rules[0])
}

func validateDataExplorerVisualConfigurationRule(rule dynatrace.VisualConfigRule) error {
	if rule.UnitTransform != "" {
		return fmt.Errorf("Data Explorer query unit must be set to 'Auto' rather than '%s'", rule.UnitTransform)
	}
	return nil
}

func (p *DataExplorerTileProcessing) processQuery(ctx context.Context, sloDefinition keptnapi.SLO, dataQuery dynatrace.DataExplorerQuery, managementZoneFilter *ManagementZoneFilter) []TileResult {
	log.WithField("metric", dataQuery.Metric).Debug("Processing data explorer query")

	metricQuery, err := p.generateMetricQueryFromDataExplorerQuery(ctx, dataQuery, managementZoneFilter)
	if err != nil {
		log.WithError(err).Warn("generateMetricQueryFromDataExplorerQuery returned an error, SLI will not be used")
		return []TileResult{newFailedTileResultFromSLODefinition(sloDefinition, "Data Explorer tile could not be converted to a metric query: "+err.Error())}
	}

	return NewMetricsQueryProcessing(p.client).Process(ctx, len(dataQuery.SplitBy), sloDefinition, metricQuery)
}

func (p *DataExplorerTileProcessing) generateMetricQueryFromDataExplorerQuery(ctx context.Context, dataQuery dynatrace.DataExplorerQuery, managementZoneFilter *ManagementZoneFilter) (*queryComponents, error) {

	// TODO 2021-08-04: there are too many return values and they are have the same type

	if dataQuery.Metric == "" {
		return nil, fmt.Errorf("metric query generation requires that data explorer query has a metric")
	}

	// Lets query the metric definition as we need to know how many dimension the metric has
	metricDefinition, err := dynatrace.NewMetricsClient(p.client).GetByID(ctx, dataQuery.Metric)
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
	}

	if len(dataQuery.SplitBy) == 1 {
		splitBy = fmt.Sprintf(":splitBy(\"%s\")", dataQuery.SplitBy[0])
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
		entitySelectorFilter, err := ensureEntitySelectorFilter(processedFilter.entitySelectorFilter, metricDefinition)
		if err != nil {
			return nil, err
		}
		processedFilter.entitySelectorFilter = entitySelectorFilter + managementZoneFilterString
	}

	// NOTE: add :names so we also get the names of the dimensions and not just the entities. This means we get two values for each dimension
	metricSelector := fmt.Sprintf("%s%s%s:%s:names",
		dataQuery.Metric, processedFilter.metricSelectorFilter, splitBy, strings.ToLower(metricAggregation))

	metricsQuery, err := metrics.NewQuery(metricSelector, processedFilter.entitySelectorFilter)
	if err != nil {
		return nil, err
	}

	return &queryComponents{
		metricsQuery: *metricsQuery,
		timeframe:    p.timeframe,
	}, nil

}

func ensureEntitySelectorFilter(existingEntitySelectorFilter string, metricDefinition *dynatrace.MetricDefinition) (string, error) {
	if existingEntitySelectorFilter != "" {
		return existingEntitySelectorFilter, nil
	}

	if len(metricDefinition.EntityType) == 0 {
		return "", fmt.Errorf("metric %s has no entity type", metricDefinition.MetricID)
	}

	return fmt.Sprintf("type(%s)", metricDefinition.EntityType[0]), nil
}

type processedFilterComponents struct {
	metricSelectorFilter string
	entitySelectorFilter string
}

func processFilter(entityType string, filter *dynatrace.DataExplorerFilter) (*processedFilterComponents, error) {
	switch filter.FilterType {
	case "ID":
		return &processedFilterComponents{
			entitySelectorFilter: fmt.Sprintf("entityId(%s)", filter.Criteria[0].Value),
		}, nil

	case "NAME":
		return &processedFilterComponents{
			entitySelectorFilter: fmt.Sprintf("type(%s),entityName(\"%s\")", entityType, filter.Criteria[0].Value),
		}, nil

	case "TAG":
		return &processedFilterComponents{
			entitySelectorFilter: fmt.Sprintf("type(%s),tag(\"%s\")", entityType, filter.Criteria[0].Value),
		}, nil

	case "ENTITY_ATTRIBUTE":
		return &processedFilterComponents{
			entitySelectorFilter: fmt.Sprintf("type(%s),%s(\"%s\")", entityType, filter.EntityAttribute, filter.Criteria[0].Value),
		}, nil

	case "DIMENSION":
		return &processedFilterComponents{
			metricSelectorFilter: fmt.Sprintf(":filter(%s(\"%s\",\"%s\"))", filter.Criteria[0].Evaluator, filter.Filter, filter.Criteria[0].Value),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported filter type: %s", filter.FilterType)
	}
}

func getSpaceAggregationTransformation(spaceAggregation string) (string, error) {
	switch spaceAggregation {
	case "":
		return "auto", nil
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

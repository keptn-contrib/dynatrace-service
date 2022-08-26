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

	tileResults := p.processMetricExpressions(ctx, sloDefinition, tile.MetricExpressions, managementZoneFilter)

	// check for mismatch between number of results and visual config type
	if tileHasSingleValueVisualization(tile) && len(tileResults) > 1 {
		return []TileResult{newFailedTileResultFromSLODefinitionAndQuery(sloDefinition, tileResults[0].sliResult.Query(), fmt.Sprintf("Data Explorer tile is configured for single value but query produced %d results", len(tileResults)))}
	}

	return tileResults
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

func tileHasSingleValueVisualization(tile *dynatrace.Tile) bool {
	if tile.VisualConfig == nil {
		return false
	}

	if tile.VisualConfig.Type == dynatrace.SingleValueVisualConfigType {
		return true
	}

	return false
}

func (p *DataExplorerTileProcessing) processMetricExpressions(ctx context.Context, sloDefinition keptnapi.SLO, metricExpressions []string, managementZoneFilter *ManagementZoneFilter) []TileResult {
	if len(metricExpressions) == 0 {
		log.Warn("processMetricExpressions found no metric expressions, SLI will not be used")
		return []TileResult{newFailedTileResultFromSLODefinition(sloDefinition, "Data Explorer tile has no metric expressions")}
	}

	if len(metricExpressions) > 2 {
		log.WithField("metricExpressions", metricExpressions).Warn("processMetricExpressions found more than 2 metric expressions")
	}

	metricsQuery, err := p.generateMetricQueryFromMetricExpression(ctx, sloDefinition, metricExpressions[0], managementZoneFilter)
	if err != nil {
		log.WithError(err).Warn("processMetricExpressions returned an error, SLI will not be used")
		return []TileResult{newFailedTileResultFromSLODefinition(sloDefinition, "Data Explorer tile could not be converted to a metric query: "+err.Error())}
	}

	return NewMetricsQueryProcessing(p.client).Process(ctx, sloDefinition, *metricsQuery, p.timeframe)
}

func (p *DataExplorerTileProcessing) generateMetricQueryFromMetricExpression(ctx context.Context, sloDefinition keptnapi.SLO, metricExpression string, managementZoneFilter *ManagementZoneFilter) (*metrics.Query, error) {
	pieces := strings.SplitN(metricExpression, "&", 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("metric expression does not contain two components: %s", metricExpression)
	}

	// TODO: 2022-08-24: support resolutions other than auto, encoded as null, assumed here to be the same as resolution inf.
	resolution, err := parseResolutionKeyValuePair(pieces[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse resolution metric expression component: %w", err)
	}

	if resolution != metrics.ResolutionInf {
		return nil, fmt.Errorf("resolution must be set to 'Auto' rather than '%s'", resolution)
	}

	return metrics.NewQueryWithResolutionAndMZSelector(pieces[1], "", resolution, managementZoneFilter.ForMZSelector())
}

func parseResolutionKeyValuePair(keyValuePair string) (string, error) {
	const resolutionPrefix = "resolution="
	if !strings.HasPrefix(keyValuePair, resolutionPrefix) {
		return "", fmt.Errorf("unexpected prefix in key value pair: %s", keyValuePair)
	}

	resolution := strings.TrimPrefix(keyValuePair, resolutionPrefix)
	if resolution == "" {
		return "", errors.New("resolution must not be empty")
	}

	if resolution == "null" {
		return metrics.ResolutionInf, nil
	}

	return resolution, nil
}

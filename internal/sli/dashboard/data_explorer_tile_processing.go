package dashboard

import (
	"context"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/ff"
	"strings"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
)

// DataExplorerTileProcessing represents the processing of a Data Explorer dashboard tile.
type DataExplorerTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	timeframe     common.Timeframe
	featureFlags  ff.GetSLIFeatureFlags
}

// NewDataExplorerTileProcessing creates a new DataExplorerTileProcessing.
func NewDataExplorerTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe, flags ff.GetSLIFeatureFlags) *DataExplorerTileProcessing {
	return &DataExplorerTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
		featureFlags:  flags,
	}
}

// Process processes the specified Data Explorer dashboard tile.
func (p *DataExplorerTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) []result.SLIWithSLO {
	validatedDataExplorerTile, err := newDataExplorerTileValidator(tile, dashboardFilter, p.featureFlags).tryValidate()
	var validationErr *dataExplorerTileValidationError
	if errors.As(err, &validationErr) {
		return []result.SLIWithSLO{result.NewFailedSLIWithSLO(validationErr.sloDefinition, err.Error())}
	}

	if validatedDataExplorerTile == nil {
		return []result.SLIWithSLO{}
	}

	return p.createMetricsQueryProcessing(validatedDataExplorerTile).Process(ctx, validatedDataExplorerTile.sloDefinition, validatedDataExplorerTile.query, p.timeframe)
}

func (p *DataExplorerTileProcessing) createMetricsQueryProcessing(validatedTile *validatedDataExplorerTile) *MetricsQueryProcessing {
	if validatedTile.singleValueVisualization {
		return NewMetricsQueryProcessingThatAllowsOnlyOneResult(p.client, validatedTile.targetUnitID, p.featureFlags)
	}

	return NewMetricsQueryProcessing(p.client, validatedTile.targetUnitID, p.featureFlags)
}

type dataExplorerTileValidationError struct {
	sloDefinition result.SLO
	errors        []error
}

func (err *dataExplorerTileValidationError) Error() string {
	var errStrings = make([]string, len(err.errors))
	for i, e := range err.errors {
		errStrings[i] = e.Error()
	}
	return fmt.Sprintf("error validating Data Explorer tile: %s", strings.Join(errStrings, "; "))
}

type dataExplorerTileValidator struct {
	tile            *dynatrace.Tile
	dashboardFilter *dynatrace.DashboardFilter
	featureFlags    ff.GetSLIFeatureFlags
}

func newDataExplorerTileValidator(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter, flags ff.GetSLIFeatureFlags) *dataExplorerTileValidator {
	return &dataExplorerTileValidator{
		tile:            tile,
		dashboardFilter: dashboardFilter,
		featureFlags:    flags,
	}
}

func (v *dataExplorerTileValidator) tryValidate() (*validatedDataExplorerTile, error) {
	sloDefinitionParsingResult, err := parseSLODefinition(v.featureFlags, v.tile.Name)
	if (err == nil) && (sloDefinitionParsingResult.exclude) {
		log.WithField("tileName", v.tile.Name).Debug("Tile excluded as name includes exclude=true")
		return nil, nil
	}

	sloDefinition := sloDefinitionParsingResult.sloDefinition

	if sloDefinition.SLI == "" {
		log.WithField("tileName", v.tile.Name).Debug("Omitted Data Explorer tile as no SLI name could be derived")
		return nil, nil
	}

	var errs []error
	if err != nil {
		errs = append(errs, err)
	}

	queryID, err := getQueryID(v.tile.Queries)
	if err != nil {
		errs = append(errs, err)
	}

	if (len(sloDefinition.Pass) == 0) && (len(sloDefinition.Warning) == 0) {
		passAndWarningProvider, err := tryGetThresholdPassAndWarningProvider(v.tile)
		if err != nil {
			errs = append(errs, err)
		}

		if passAndWarningProvider != nil {
			sloDefinition.Pass = passAndWarningProvider.getPass()
			sloDefinition.Warning = passAndWarningProvider.getWarning()
		}
	}

	query, err := createMetricsQueryForMetricExpressions(v.tile.MetricExpressions, NewManagementZoneFilter(v.dashboardFilter, v.tile.TileFilter.ManagementZone))
	if err != nil {
		log.WithError(err).Warn("createMetricsQueryForMetricExpressions returned an error, SLI will not be used")
		errs = append(errs, err)
	}

	targetUnitID, err := getUnitTransform(v.tile.VisualConfig, queryID)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, &dataExplorerTileValidationError{
			sloDefinition: sloDefinition,
			errors:        errs,
		}
	}

	return &validatedDataExplorerTile{
		sloDefinition:            sloDefinition,
		targetUnitID:             targetUnitID,
		singleValueVisualization: isSingleValueVisualizationType(v.tile.VisualConfig),
		query:                    *query,
	}, nil
}

// getQueryID gets the single enabled query ID or returns an error.
func getQueryID(queries []dynatrace.DataExplorerQuery) (string, error) {
	if len(queries) == 0 {
		return "", errors.New("Data Explorer tile has no query")
	}

	enabledQueryIDs := make([]string, 0, len(queries))
	for _, q := range queries {
		if q.Enabled {
			enabledQueryIDs = append(enabledQueryIDs, q.ID)
		}
	}

	if len(enabledQueryIDs) == 0 {
		return "", errors.New("Data Explorer tile has no query enabled")
	}

	if len(enabledQueryIDs) > 1 {
		return "", fmt.Errorf("Data Explorer tile has %d queries enabled but only one is supported", len(enabledQueryIDs))
	}

	return enabledQueryIDs[0], nil
}

func getUnitTransform(visualConfig *dynatrace.VisualizationConfiguration, queryID string) (string, error) {
	if visualConfig == nil {
		return "", nil
	}

	if queryID == "" {
		return "", nil
	}

	matchingRule, err := tryGetMatchingVisualizationRule(visualConfig.Rules, queryID)
	if err != nil {
		return "", fmt.Errorf("could not get unit transform: %w", err)
	}

	if matchingRule == nil {
		return "", nil
	}

	return matchingRule.UnitTransform, nil
}

func tryGetMatchingVisualizationRule(rules []dynatrace.VisualizationRule, queryID string) (*dynatrace.VisualizationRule, error) {
	queryMatcher := createQueryMatcher(queryID)
	var matchingRules []dynatrace.VisualizationRule
	for _, r := range rules {
		if r.Matcher == queryMatcher {
			matchingRules = append(matchingRules, r)
		}
	}

	if len(matchingRules) == 0 {
		return nil, nil
	}

	if len(matchingRules) > 1 {
		return nil, fmt.Errorf("expected one visualization rule for query '%s' but found %d", queryID, len(matchingRules))
	}

	return &matchingRules[0], nil
}

func createQueryMatcher(queryID string) string {
	return queryID + ":"
}

func isSingleValueVisualizationType(visualConfig *dynatrace.VisualizationConfiguration) bool {
	if visualConfig == nil {
		return false
	}

	return visualConfig.Type == dynatrace.SingleValueVisualizationConfigurationType
}

func createMetricsQueryForMetricExpressions(metricExpressions []string, managementZoneFilter *ManagementZoneFilter) (*metrics.Query, error) {
	if len(metricExpressions) == 0 {
		return nil, errors.New("Data Explorer tile has no metric expressions")
	}

	if len(metricExpressions) > 2 {
		log.WithField("metricExpressions", metricExpressions).Warn("processMetricExpressions found more than 2 metric expressions")
	}

	return createMetricsQueryForMetricExpression(metricExpressions[0], managementZoneFilter)
}

func createMetricsQueryForMetricExpression(metricExpression string, managementZoneFilter *ManagementZoneFilter) (*metrics.Query, error) {
	pieces := strings.SplitN(metricExpression, "&", 2)
	if len(pieces) != 2 {
		return nil, fmt.Errorf("metric expression does not contain two components: %s", metricExpression)
	}

	resolution, err := parseResolutionKeyValuePair(pieces[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse resolution metric expression component: %w", err)
	}

	return metrics.NewQuery(pieces[1], "", resolution, managementZoneFilter.ForMZSelector())
}

// parseResolutionKeyValuePair parses the resolution key value pair, returning resolution or error. In the case that no resolution is set in UI, i.e. resolution=null, an empty string is returned.
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
		return "", nil
	}

	return resolution, nil
}

type validatedDataExplorerTile struct {
	sloDefinition            result.SLO
	targetUnitID             string
	singleValueVisualization bool
	query                    metrics.Query
}

package dynatrace

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
)

const (
	emptyUnitID       = ""
	autoUnitID        = "auto"
	noneUnitID        = "none"
	countUnitID       = "Count"
	unspecifiedUnitID = "Unspecified"

	kiloUnitID     = "Kilo"
	millionUnitID  = "Million"
	billionUnitID  = "Billion"
	trillionUnitID = "Trillion"
)

// unitlessConversionSnippetsMap lists all support conversions from Count or Unspecified to target via the map'S key-
var unitlessConversionSnippetsMap map[string]string = map[string]string{
	kiloUnitID:     "/1000",
	millionUnitID:  "/1000000",
	billionUnitID:  "/1000000000",
	trillionUnitID: "/1000000000000",
}

// UnknownUnitlessConversionError represents the error that an invalid target unit was specified for a unitless conversion.
type UnknownUnitlessConversionError struct {
	unitID string
}

func (e *UnknownUnitlessConversionError) Error() string {
	return fmt.Sprintf("unknown unit '%s'", e.unitID)
}

// UnableToApplyFoldToValueDefaultAggregationError represents the error that a fold is unable to be applied to a metric selector as the default aggregation type is 'value'.
type UnableToApplyFoldToValueDefaultAggregationError struct{}

func (e *UnableToApplyFoldToValueDefaultAggregationError) Error() string {
	return "unable to apply ':fold' to the metric selector as the default aggregation type is 'value'"
}

// metricQueryModifier modifies a metrics query to return a single result and / or to convert the unit of values returned.
// It assumes that the metric definition obtained from the original metric selector is valid also for the modified ones and thus that it can be cached.
type metricsQueryModifier struct {
	metricsClient          MetricsClientInterface
	query                  metrics.Query
	includeFold            bool
	unitsConversionSnippet string
	setResolutionToInf     bool
	cachedMetricDefinition *MetricDefinition
}

func newMetricsQueryModifier(metricsClient MetricsClientInterface, query metrics.Query) *metricsQueryModifier {
	return &metricsQueryModifier{
		metricsClient: metricsClient,
		query:         query,
	}
}

// applyUnitConversion modifies the metric selector such that the result has the specified unit.
func (u *metricsQueryModifier) applyUnitConversion(ctx context.Context, targetUnitID string) error {
	if !doesTargetUnitRequireConversion(targetUnitID) {
		return nil
	}

	metricDefinition, err := u.getCachedMetricDefinition(ctx)
	if err != nil {
		return err
	}

	if metricDefinition.Unit == targetUnitID {
		return nil
	}

	unitsConversionSnippet, err := getConversionSnippet(metricDefinition, targetUnitID)
	if err != nil {
		return err
	}

	u.unitsConversionSnippet = unitsConversionSnippet
	return nil
}

// doesTargetUnitRequireConversion checks if the target unit ID requires conversion or not. Currently, "Auto" (default empty value and explicit `auto` value) and "None" require no conversion.
func doesTargetUnitRequireConversion(targetUnitID string) bool {
	switch targetUnitID {
	case emptyUnitID, autoUnitID, noneUnitID:
		return false
	default:
		return true
	}
}

// applyFoldOrResolutionInf modifies the query to use resolution Inf or a fold such that each metric series returns a single value.
func (u *metricsQueryModifier) applyFoldOrResolutionInf(ctx context.Context) error {
	metricDefinition, err := u.getCachedMetricDefinition(ctx)
	if err != nil {
		return err
	}

	if (u.query.GetResolution() == "") && metricDefinition.ResolutionInfSupported {
		u.setResolutionToInf = true
		return nil
	}

	if metricDefinition.DefaultAggregation.Type == AggregationTypeValue {
		return &UnableToApplyFoldToValueDefaultAggregationError{}
	}

	u.includeFold = true
	return nil
}

// getModifiedQuery gets the modified query with any resolution change, or fold or units conversion.
func (u *metricsQueryModifier) getModifiedQuery() (*metrics.Query, error) {
	return metrics.NewQuery(u.getModifiedMetricSelector(), u.query.GetEntitySelector(), u.getModifiedResolution(), u.query.GetMZSelector())
}

func (u *metricsQueryModifier) getModifiedMetricSelector() string {
	modifiedMetricSelector := u.query.GetMetricSelector()
	if u.includeFold {
		modifiedMetricSelector = "(" + modifiedMetricSelector + "):fold"
	}

	if u.unitsConversionSnippet != "" {
		modifiedMetricSelector = "(" + modifiedMetricSelector + ")" + u.unitsConversionSnippet
	}
	return modifiedMetricSelector
}

func (u *metricsQueryModifier) getModifiedResolution() string {
	if u.setResolutionToInf {
		return metrics.ResolutionInf
	}
	return u.query.GetResolution()
}

func (u *metricsQueryModifier) getCachedMetricDefinition(ctx context.Context) (*MetricDefinition, error) {
	if u.cachedMetricDefinition != nil {
		return u.cachedMetricDefinition, nil
	}

	metricDefinition, err := u.metricsClient.GetMetricDefinitionByID(ctx, u.query.GetMetricSelector())
	if err != nil {
		return nil, err
	}

	u.cachedMetricDefinition = metricDefinition
	return u.cachedMetricDefinition, nil
}

func doesMetricDefinitionSupportToUnitTransformation(metricDefinition *MetricDefinition) bool {
	const toUnitTransformation = "toUnit"

	for _, t := range metricDefinition.Transformations {
		if t == toUnitTransformation {
			return true
		}
	}
	return false
}

func getConversionSnippet(metricDefinition *MetricDefinition, targetUnitID string) (string, error) {
	sourceUnitID := metricDefinition.Unit

	if shouldDoUnitlessConversion(sourceUnitID) {
		return getUnitlessConversionSnippet(targetUnitID)
	}

	if doesMetricDefinitionSupportToUnitTransformation(metricDefinition) {
		return getToUnitConversionSnippet(sourceUnitID, targetUnitID), nil
	} else {
		return getAutoToUnitConversionSnippet(sourceUnitID, targetUnitID), nil
	}
}

func shouldDoUnitlessConversion(sourceUnitID string) bool {
	switch sourceUnitID {
	case countUnitID, unspecifiedUnitID:
		return true

	default:
		return false
	}
}

func getUnitlessConversionSnippet(targetUnitID string) (string, error) {
	snippet, ok := unitlessConversionSnippetsMap[targetUnitID]
	if !ok {
		return "", &UnknownUnitlessConversionError{unitID: targetUnitID}
	}

	return snippet, nil
}

func getAutoToUnitConversionSnippet(sourceUnitID, targetUnitID string) string {
	return fmt.Sprintf(":auto%s", getToUnitConversionSnippet(sourceUnitID, targetUnitID))
}

func getToUnitConversionSnippet(sourceUnitID, targetUnitID string) string {
	return fmt.Sprintf(":toUnit(%s,%s)", sourceUnitID, targetUnitID)
}

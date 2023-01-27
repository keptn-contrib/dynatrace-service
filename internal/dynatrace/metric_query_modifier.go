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

// unitlessConversions lists all support conversions from Count or Unspecified to target via the map's key.
var unitlessConversions = map[string]string{
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
	metricsClient                       MetricsClientInterface
	query                               metrics.Query
	includeFold                         bool
	unitsConversionMetricSelectorSuffix string
	setResolutionToInf                  bool
	metricDefinition                    *MetricDefinition
}

func newMetricsQueryModifier(metricsClient MetricsClientInterface, query metrics.Query) *metricsQueryModifier {
	return &metricsQueryModifier{
		metricsClient: metricsClient,
		query:         query,
	}
}

// applyUnitConversion modifies the metric selector such that the result has the specified unit and returns the current modified query or an error.
func (u *metricsQueryModifier) applyUnitConversion(ctx context.Context, targetUnitID string) (*metrics.Query, error) {
	if !doesTargetUnitRequireConversion(targetUnitID) {
		return u.getModifiedQuery()
	}

	metricDefinition, err := u.getMetricDefinition(ctx)
	if err != nil {
		return nil, err
	}

	if metricDefinition.Unit == targetUnitID {
		return u.getModifiedQuery()
	}

	unitsConversionSnippet, err := getConversionMetricSelectorSuffix(metricDefinition, targetUnitID)
	if err != nil {
		return nil, err
	}

	u.unitsConversionMetricSelectorSuffix = unitsConversionSnippet
	return u.getModifiedQuery()
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

// applyFoldOrResolutionInf modifies the query to use resolution Inf or a fold such that each metric series returns a single value and returns the current modified query or an error.
func (u *metricsQueryModifier) applyFoldOrResolutionInf(ctx context.Context) (*metrics.Query, error) {
	metricDefinition, err := u.getMetricDefinition(ctx)
	if err != nil {
		return nil, err
	}

	if (u.query.GetResolution() == "") && metricDefinition.ResolutionInfSupported {
		u.setResolutionToInf = true
		return u.getModifiedQuery()
	}

	if metricDefinition.DefaultAggregation.Type == AggregationTypeValue {
		return nil, &UnableToApplyFoldToValueDefaultAggregationError{}
	}

	u.includeFold = true
	return u.getModifiedQuery()
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

	if u.unitsConversionMetricSelectorSuffix != "" {
		modifiedMetricSelector = "(" + modifiedMetricSelector + ")" + u.unitsConversionMetricSelectorSuffix
	}
	return modifiedMetricSelector
}

func (u *metricsQueryModifier) getModifiedResolution() string {
	if u.setResolutionToInf {
		return metrics.ResolutionInf
	}
	return u.query.GetResolution()
}

func (u *metricsQueryModifier) getMetricDefinition(ctx context.Context) (*MetricDefinition, error) {
	if u.metricDefinition != nil {
		return u.metricDefinition, nil
	}

	metricDefinition, err := u.metricsClient.GetMetricDefinitionByID(ctx, u.query.GetMetricSelector())
	if err != nil {
		return nil, err
	}

	u.metricDefinition = metricDefinition
	return u.metricDefinition, nil
}

func doesMetricKeySupportToUnitTransformation(metricDefinition *MetricDefinition) bool {
	const toUnitTransformation = "toUnit"

	for _, t := range metricDefinition.Transformations {
		if t == toUnitTransformation {
			return true
		}
	}
	return false
}

func getConversionMetricSelectorSuffix(metricDefinition *MetricDefinition, targetUnitID string) (string, error) {
	sourceUnitID := metricDefinition.Unit

	if shouldDoUnitlessConversion(sourceUnitID) {
		return getUnitlessConversionMetricSelectorSuffix(targetUnitID)
	}

	if doesMetricKeySupportToUnitTransformation(metricDefinition) {
		return getToUnitConversionMetricSelectorSuffix(sourceUnitID, targetUnitID), nil
	}

	return getAutoToUnitConversionMetricSelectorSuffix(sourceUnitID, targetUnitID), nil
}

func shouldDoUnitlessConversion(sourceUnitID string) bool {
	switch sourceUnitID {
	case countUnitID, unspecifiedUnitID:
		return true

	default:
		return false
	}
}

func getUnitlessConversionMetricSelectorSuffix(targetUnitID string) (string, error) {
	snippet, ok := unitlessConversions[targetUnitID]
	if !ok {
		return "", &UnknownUnitlessConversionError{unitID: targetUnitID}
	}

	return snippet, nil
}

func getAutoToUnitConversionMetricSelectorSuffix(sourceUnitID, targetUnitID string) string {
	return fmt.Sprintf(":auto%s", getToUnitConversionMetricSelectorSuffix(sourceUnitID, targetUnitID))
}

func getToUnitConversionMetricSelectorSuffix(sourceUnitID, targetUnitID string) string {
	return fmt.Sprintf(":toUnit(%s,%s)", sourceUnitID, targetUnitID)
}

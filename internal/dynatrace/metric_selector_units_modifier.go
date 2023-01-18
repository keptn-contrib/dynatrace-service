package dynatrace

import (
	"context"
	"fmt"
)

// metricSelectorUnitsModifier modifies a metric selector to return results in the specified unit.
type metricSelectorUnitsModifier struct {
	metricsClient MetricsClientInterface
}

func newMetricSelectorUnitsModifier(metricsClient MetricsClientInterface) *metricSelectorUnitsModifier {
	return &metricSelectorUnitsModifier{
		metricsClient: metricsClient,
	}
}

// UnknownUnitlessConversionError represents the error that an invalid target unit was specified for a unitless conversion.
type UnknownUnitlessConversionError struct {
	unitID string
}

func (e *UnknownUnitlessConversionError) Error() string {
	return fmt.Sprintf("unknown unit '%s'", e.unitID)
}

const (
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

// applyUnit modifies the metric selector to return results in the specified unit.
func (u *metricSelectorUnitsModifier) applyUnit(ctx context.Context, metricSelector string, targetUnitID string) (string, error) {
	metricDefinition, err := u.metricsClient.GetMetricDefinitionByID(ctx, metricSelector)
	if err != nil {
		return "", err
	}

	sourceUnitID := metricDefinition.Unit
	if sourceUnitID == targetUnitID {
		return metricSelector, nil
	}

	snippet, err := u.getConversionSnippet(
		metricSelector,
		doesMetricDefinitionSupportToUnitTransformation(metricDefinition),
		sourceUnitID,
		targetUnitID,
	)
	if err != nil {
		return "", err
	}

	return "(" + metricSelector + ")" + snippet, nil
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

func (u *metricSelectorUnitsModifier) getConversionSnippet(metricSelector string, supportsToUnitTransformation bool, sourceUnitID string, targetUnitID string) (string, error) {
	if shouldDoUnitlessConversion(sourceUnitID) {
		return getUnitlessConversionSnippet(targetUnitID)
	}

	if supportsToUnitTransformation {
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

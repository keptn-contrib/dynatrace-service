package dynatrace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// MetricsUnitsPath is the base endpoint for Metrics Units API
const MetricsUnitsPath = "/api/v2/units"

// UnitConversionResult is the result of a conversion.
type UnitConversionResult struct {
	UnitID      string  `json:"unitId"`
	ResultValue float64 `json:"resultValue"`
}

const (
	valueKey      = "value"
	targetUnitKey = "targetUnit"
)

// MetricsUnitsClientConvertRequest encapsulates the request for the MetricsUnitsClient's Convert method.
type MetricsUnitsClientConvertRequest struct {
	sourceUnitID string
	value        float64
	targetUnitID string
}

// NewMetricsUnitsClientConvertRequest creates a new MetricsUnitsClientConvertRequest.
func NewMetricsUnitsClientConvertRequest(sourceUnitID string, value float64, targetUnitID string) MetricsUnitsClientConvertRequest {
	return MetricsUnitsClientConvertRequest{
		sourceUnitID: sourceUnitID,
		value:        value,
		targetUnitID: targetUnitID,
	}
}

// RequestString encodes MetricsUnitsClientConvertRequest into a request string.
func (q *MetricsUnitsClientConvertRequest) RequestString() string {
	queryParameters := newQueryParameters()
	queryParameters.add(valueKey, strconv.FormatFloat(q.value, 'f', -1, 64))
	queryParameters.add(targetUnitKey, q.targetUnitID)

	return MetricsUnitsPath + "/" + url.PathEscape(q.sourceUnitID) + "/convert?" + queryParameters.encode()
}

// MetricsUnitsClientInterface defines functions for the Dynatrace Metrics Units endpoint.
type MetricsUnitsClientInterface interface {
	// Convert converts a value between the specified units.
	Convert(ctx context.Context, request MetricsUnitsClientConvertRequest) (float64, error)
}

// MetricsUnitsClient is a client for interacting with Dynatrace Metrics Units endpoint.
type MetricsUnitsClient struct {
	client ClientInterface
}

// NewMetricsUnitsClient creates a new MetricsUnitsClient
func NewMetricsUnitsClient(client ClientInterface) *MetricsUnitsClient {
	return &MetricsUnitsClient{
		client: client,
	}
}

// Convert converts a value between the specified units.
func (c *MetricsUnitsClient) Convert(ctx context.Context, request MetricsUnitsClientConvertRequest) (float64, error) {
	body, err := c.client.Get(ctx, request.RequestString())
	if err != nil {
		return 0, err
	}

	var result UnitConversionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, err
	}

	return result.ResultValue, nil
}

// EnhancedMetricsUnitsDecorator builds on MetricsUnitsClient by providing local conversion for display units for "Count" and "Unspecified" units.
type EnhancedMetricsUnitsDecorator struct {
	baseMetricsUnitsClient MetricsUnitsClientInterface
}

// NewEnhancedMetricsUnitsDecorator creates a new EnhancedMetricsUnitsDecorator
func NewEnhancedMetricsUnitsDecorator(baseMetricsUnitsClient MetricsUnitsClientInterface) *EnhancedMetricsUnitsDecorator {
	return &EnhancedMetricsUnitsDecorator{
		baseMetricsUnitsClient: baseMetricsUnitsClient,
	}
}

// Convert converts a value between the specified units.
func (c *EnhancedMetricsUnitsDecorator) Convert(ctx context.Context, request MetricsUnitsClientConvertRequest) (float64, error) {
	if shouldDoLocalConversion(request.sourceUnitID) {
		return scaleValueForUnitID(request.value, request.targetUnitID)
	}

	return c.baseMetricsUnitsClient.Convert(ctx, request)
}

func scaleValueForUnitID(value float64, targetUnitID string) (float64, error) {
	scaleFactor, err := getScaleFactor(targetUnitID)
	if err != nil {
		return 0, err
	}

	return value / scaleFactor, nil
}

const (
	kiloUnitID     = "Kilo"
	millionUnitID  = "Million"
	billionUnitID  = "Billion"
	trillionUnitID = "Trillion"
)

const (
	countUnitID       = "Count"
	unspecifiedUnitID = "Unspecified"
)

func shouldDoLocalConversion(sourceUnitID string) bool {
	switch sourceUnitID {
	case countUnitID, unspecifiedUnitID:
		return true

	default:
		return false
	}
}

func getScaleFactor(unitID string) (float64, error) {
	switch unitID {
	case kiloUnitID:
		return 1000, nil

	case millionUnitID:
		return 1000000, nil

	case billionUnitID:
		return 1000000000, nil

	case trillionUnitID:
		return 1000000000000, nil

	default:
		return 0, errors.New(fmt.Sprintf("unknown unit '%s'", unitID))
	}
}

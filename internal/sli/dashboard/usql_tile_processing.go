package dashboard

import (
	"context"
	"errors"
	"fmt"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
	v1usql "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/usql"
)

// USQLTileProcessing represents the processing of a USQL dashboard tile.
type USQLTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	timeframe     common.Timeframe
}

// NewUSQLTileProcessing creates a new USQLTileProcessing.
func NewUSQLTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe) *USQLTileProcessing {
	return &USQLTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
	}
}

// Process processes the specified USQL dashboard tile.
// TODO: 2022-03-07: Investigate if all error and warning cases are covered. E.g. what happens if a query returns no results?
func (p *USQLTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile) []*TileResult {
	sloDefinition, err := common.ParseSLOFromString(tile.CustomName)
	var sloDefError *common.SLODefinitionError
	if errors.As(err, &sloDefError) {
		failedTileResult := newFailedTileResultFromError(sloDefError.SLINameOrTileTitle(), "User Sessions Query tile title parsing error", err)
		return []*TileResult{&failedTileResult}
	}

	query, err := usql.NewQuery(tile.Query)
	if err != nil {
		failedTileResult := newFailedTileResultFromSLODefinition(sloDefinition, "error creating USQL query: "+err.Error())
		return []*TileResult{&failedTileResult}
	}

	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(ctx, dynatrace.NewUSQLClientQueryParameters(*query, p.timeframe))
	if err != nil {
		failedTileResult := newFailedTileResultFromSLODefinition(sloDefinition, "error querying User sessions API: "+err.Error())
		return []*TileResult{&failedTileResult}
	}

	switch tile.Type {
	case dynatrace.SingleValueVisualizationType:
		tileResult := processQueryResultForSingleValue(*usqlResult, sloDefinition, *query)
		return []*TileResult{&tileResult}
	case dynatrace.ColumnChartVisualizationType, dynatrace.LineChartVisualizationType, dynatrace.PieChartVisualizationType, dynatrace.TableVisualizationType:
		return processQueryResultForMultipleValues(*usqlResult, sloDefinition, tile.Type, *query)
	default:
		// generate failed tile result specifically because it is unsupported
		failedTileResult := newFailedTileResultFromSLODefinition(sloDefinition, "unsupported USQL visualization type: "+tile.Type)
		return []*TileResult{&failedTileResult}
	}
}

func processQueryResultForSingleValue(usqlResult dynatrace.DTUSQLResult, sloDefinition *keptncommon.SLO, baseQuery usql.Query) TileResult {
	if len(usqlResult.Values) == 0 {
		return newWarningTileResultFromSLODefinition(sloDefinition, "User sessions API returned zero values")
	}

	if len(usqlResult.ColumnNames) != 1 || len(usqlResult.Values) != 1 {
		return newWarningTileResultFromSLODefinition(sloDefinition, fmt.Sprintf("USQL visualization type %s should only return a single result", dynatrace.SingleValueVisualizationType))
	}
	dimensionValue, err := tryCastDimensionValueToNumeric(usqlResult.Values[0][0])
	if err != nil {
		return newWarningTileResultFromSLODefinition(sloDefinition, err.Error())
	}

	return createSuccessfulTileResultForDimensionNameAndValue("", dimensionValue, sloDefinition, dynatrace.SingleValueVisualizationType, baseQuery)
}

func processQueryResultForMultipleValues(usqlResult dynatrace.DTUSQLResult, sloDefinition *keptncommon.SLO, visualizationType string, baseQuery usql.Query) []*TileResult {
	if len(usqlResult.Values) == 0 {
		warningTileResult := newWarningTileResultFromSLODefinition(sloDefinition, "User sessions API returned zero values")
		return []*TileResult{&warningTileResult}
	}

	if len(usqlResult.ColumnNames) < 2 {
		warningTileResult := newWarningTileResultFromSLODefinition(sloDefinition, fmt.Sprintf("USQL result type %s should have at least two columns", visualizationType))
		return []*TileResult{&warningTileResult}
	}

	var tileResults []*TileResult
	for _, rowValue := range usqlResult.Values {
		dimensionName, err := tryCastDimensionNameToString(rowValue[0])
		if err != nil {
			warningTileResult := newWarningTileResultFromSLODefinition(sloDefinition, err.Error())
			tileResults = append(tileResults, &warningTileResult)
			continue
		}

		dimensionValue, err := tryGetDimensionValueForVisualizationType(rowValue, visualizationType)
		if err != nil {
			warningTileResult := newWarningTileResultFromSLODefinition(sloDefinition, err.Error())
			tileResults = append(tileResults, &warningTileResult)
			continue
		}

		tileResult := createSuccessfulTileResultForDimensionNameAndValue(dimensionName, dimensionValue, sloDefinition, visualizationType, baseQuery)
		tileResults = append(tileResults, &tileResult)
	}
	return tileResults
}

func tryGetDimensionValueForVisualizationType(rowValue []interface{}, visualizationType string) (float64, error) {
	var rawValue interface{}
	switch visualizationType {
	case dynatrace.ColumnChartVisualizationType, dynatrace.LineChartVisualizationType, dynatrace.PieChartVisualizationType:
		rawValue = rowValue[1]
	case dynatrace.TableVisualizationType:
		rawValue = rowValue[len(rowValue)-1]
	default:
		return 0, fmt.Errorf("unsupported USQL visualization type: %s", visualizationType)
	}

	value, err := tryCastDimensionValueToNumeric(rawValue)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func tryCastDimensionValueToNumeric(dimensionValue interface{}) (float64, error) {
	value, ok := dimensionValue.(float64)
	if ok {
		return value, nil
	}

	return 0, errors.New("dimension value should be a number")
}

func tryCastDimensionNameToString(dimensionName interface{}) (string, error) {
	value, ok := dimensionName.(string)
	if ok {
		return value, nil
	}

	return "", errors.New("dimension name should be a string")
}

func createSuccessfulTileResultForDimensionNameAndValue(dimensionName string, dimensionValue float64, sloDefinition *keptncommon.SLO, visualizationType string, baseQuery usql.Query) TileResult {
	indicatorName := sloDefinition.SLI
	if dimensionName != "" {
		indicatorName = common.CleanIndicatorName(indicatorName + "_" + dimensionName)
	}

	v1USQLQuery, err := v1usql.NewQuery(visualizationType, dimensionName, baseQuery)
	if err != nil {
		return newFailedTileResultFromSLODefinition(sloDefinition, "could not create USQL v1 query: "+err.Error())
	}

	return TileResult{
		sliResult: result.NewSuccessfulSLIResult(indicatorName, dimensionValue),
		objective: &keptncommon.SLO{
			SLI:     indicatorName,
			Weight:  sloDefinition.Weight,
			KeySLI:  sloDefinition.KeySLI,
			Pass:    sloDefinition.Pass,
			Warning: sloDefinition.Warning,
		},
		sliName:  indicatorName,
		sliQuery: v1usql.NewQueryProducer(*v1USQLQuery).Produce(),
	}
}

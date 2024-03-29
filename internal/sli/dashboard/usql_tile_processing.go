package dashboard

import (
	"context"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/ff"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
)

const zeroValuesMessage = "User sessions API returned zero values"

// USQLTileProcessing represents the processing of a USQL dashboard tile.
type USQLTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	timeframe     common.Timeframe
	featureFlags  ff.GetSLIFeatureFlags
}

// NewUSQLTileProcessing creates a new USQLTileProcessing.
func NewUSQLTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe, flags ff.GetSLIFeatureFlags) *USQLTileProcessing {
	return &USQLTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
		featureFlags:  flags,
	}
}

// Process processes the specified USQL dashboard tile.
// TODO: 2022-03-07: Investigate if all error and warning cases are covered. E.g. what happens if a query returns no results?
func (p *USQLTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile) []result.SLIWithSLO {
	sloDefinitionParsingResult, err := parseSLODefinition(p.featureFlags, tile.CustomName)
	if (err == nil) && (sloDefinitionParsingResult.exclude) {
		log.WithField("tile.CustomName", tile.Name).Debug("Tile excluded as name includes exclude=true")
		return nil
	}

	sloDefinition := sloDefinitionParsingResult.sloDefinition
	if sloDefinition.SLI == "" {
		log.WithField("tile.CustomName", tile.Name).Debug("Omitted User Sessions Query tile as no SLI name could be derived")
		return nil
	}

	if err != nil {
		return []result.SLIWithSLO{result.NewFailedSLIWithSLO(sloDefinition, "User Sessions Query tile title parsing error: "+err.Error())}
	}

	query, err := usql.NewQuery(tile.Query)
	if err != nil {
		return []result.SLIWithSLO{result.NewFailedSLIWithSLO(sloDefinition, "error creating USQL query: "+err.Error())}
	}

	request := dynatrace.NewUSQLClientQueryRequest(*query, p.timeframe)
	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(ctx, request)
	if err != nil {
		return []result.SLIWithSLO{result.NewFailedSLIWithSLOAndQuery(sloDefinition, request.RequestString(), "error querying User sessions API: "+err.Error())}
	}

	switch tile.Type {
	case dynatrace.SingleValueVisualizationType:
		return []result.SLIWithSLO{processQueryResultForSingleValue(*usqlResult, sloDefinition, request)}
	case dynatrace.ColumnChartVisualizationType, dynatrace.LineChartVisualizationType, dynatrace.PieChartVisualizationType, dynatrace.TableVisualizationType:
		return processQueryResultForMultipleValues(*usqlResult, sloDefinition, tile.Type, request, p.featureFlags)
	default:
		// generate failed tile result specifically because it is unsupported
		return []result.SLIWithSLO{result.NewFailedSLIWithSLOAndQuery(sloDefinition, request.RequestString(), "unsupported USQL visualization type: "+tile.Type)}
	}
}

func processQueryResultForSingleValue(usqlResult dynatrace.DTUSQLResult, sloDefinition result.SLO, request dynatrace.USQLClientQueryRequest) result.SLIWithSLO {
	if len(usqlResult.Values) == 0 {
		return result.NewWarningSLIWithSLOAndQuery(sloDefinition, request.RequestString(), zeroValuesMessage)
	}

	if len(usqlResult.ColumnNames) != 1 || len(usqlResult.Values) != 1 {
		return result.NewWarningSLIWithSLOAndQuery(sloDefinition, request.RequestString(), fmt.Sprintf("USQL visualization type %s should only return a single result", dynatrace.SingleValueVisualizationType))
	}
	dimensionValue, err := tryCastDimensionValueToNumeric(usqlResult.Values[0][0])
	if err != nil {
		return result.NewWarningSLIWithSLOAndQuery(sloDefinition, request.RequestString(), err.Error())
	}

	return result.NewSuccessfulSLIWithSLOAndQuery(sloDefinition, dimensionValue, request.RequestString())
}

func processQueryResultForMultipleValues(usqlResult dynatrace.DTUSQLResult, sloDefinition result.SLO, visualizationType string, request dynatrace.USQLClientQueryRequest, flags ff.GetSLIFeatureFlags) []result.SLIWithSLO {
	if len(usqlResult.Values) == 0 {
		return []result.SLIWithSLO{result.NewWarningSLIWithSLOAndQuery(sloDefinition, request.RequestString(), zeroValuesMessage)}
	}

	if len(usqlResult.ColumnNames) < 2 {
		return []result.SLIWithSLO{result.NewWarningSLIWithSLOAndQuery(sloDefinition, request.RequestString(), fmt.Sprintf("USQL result type %s should have at least two columns", visualizationType))}
	}

	var tileResults []result.SLIWithSLO
	for index, rowValue := range usqlResult.Values {
		dimensionName, err := tryCastDimensionNameToString(rowValue[0])
		if err != nil {
			tileResults = append(tileResults, newWarningTileResultWithIndexFromSLODefinitionAndQuery(index, sloDefinition, request, err.Error(), flags))
			continue
		}

		dimensionValue, err := tryGetDimensionValueForVisualizationType(rowValue, visualizationType)
		if err != nil {
			tileResults = append(tileResults, newWarningTileResultWithIndexFromSLODefinitionAndQuery(index, sloDefinition, request, err.Error(), flags))
			continue
		}

		tileResults = append(tileResults, newSuccessfulTileResultForDimensionNameAndValue(dimensionName, dimensionValue, sloDefinition, request, flags))
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

func newSuccessfulTileResultForDimensionNameAndValue(dimensionName string, dimensionValue float64, sloDefinition result.SLO, request dynatrace.USQLClientQueryRequest, flags ff.GetSLIFeatureFlags) result.SLIWithSLO {
	return result.NewSuccessfulSLIWithSLOAndQuery(
		result.SLO{
			SLI:         buildIndicatorNameWithDimensionName(flags, sloDefinition.SLI, dimensionName),
			DisplayName: buildDisplayNameWithDimensionName(sloDefinition.DisplayName, dimensionName),
			Weight:      sloDefinition.Weight,
			KeySLI:      sloDefinition.KeySLI,
			Pass:        sloDefinition.Pass,
			Warning:     sloDefinition.Warning,
		},
		dimensionValue,
		request.RequestString(),
	)
}

func newWarningTileResultWithIndexFromSLODefinitionAndQuery(index int, sloDefinition result.SLO, request dynatrace.USQLClientQueryRequest, message string, flags ff.GetSLIFeatureFlags) result.SLIWithSLO {
	return result.NewWarningSLIWithSLOAndQuery(
		result.SLO{
			SLI:         cleanIndicatorName(flags.SkipLowercaseSLINames(), fmt.Sprintf("%s_%d", sloDefinition.SLI, index+1)),
			DisplayName: fmt.Sprintf("%s (%d)", sloDefinition.DisplayName, index+1),
			Weight:      sloDefinition.Weight,
			KeySLI:      sloDefinition.KeySLI,
			Pass:        sloDefinition.Pass,
			Warning:     sloDefinition.Warning,
		},
		request.RequestString(),
		message,
	)
}

func buildIndicatorNameWithDimensionName(flags ff.GetSLIFeatureFlags, baseIndicatorName string, dimensionName string) string {
	if dimensionName == "" {
		return baseIndicatorName
	}

	return cleanIndicatorName(flags.SkipLowercaseSLINames(), baseIndicatorName+"_"+dimensionName)
}

func buildDisplayNameWithDimensionName(baseDisplayName string, dimensionName string) string {
	if dimensionName == "" {
		return baseDisplayName
	}

	return baseDisplayName + " (" + dimensionName + ")"
}

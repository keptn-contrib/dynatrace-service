package query

import (
	"errors"
	"fmt"
	"strings"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/unit"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
)

type Processing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	customQueries *keptn.CustomQueries
	startUnix     time.Time
	endUnix       time.Time
}

func NewProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, customQueries *keptn.CustomQueries, startUnix time.Time, endUnix time.Time) *Processing {
	return &Processing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		customQueries: customQueries,
		startUnix:     startUnix,
		endUnix:       endUnix,
	}
}

// GetSLIValue queries a single metric value from Dynatrace API.
// Can handle both Metric Queries as well as USQL
func (p *Processing) GetSLIValue(name string) (float64, error) {
	// first we get the query from the SLI configuration based on its logical name
	// no default values here anymore if indicator could not be matched (e.g. due to a misspelling) and custom SLIs were defined
	rawQuery, err := p.customQueries.GetQueryByNameOrDefaultIfEmpty(name)
	if err != nil {
		return 0, err
	}

	sliQuery := common.ReplaceQueryParameters(rawQuery, p.customFilters, p.eventData)

	log.WithFields(
		log.Fields{
			"name":     name,
			"rawQuery": rawQuery,
			"query":    sliQuery,
		}).Debug("Retrieved SLI query")

	switch {
	case strings.HasPrefix(sliQuery, "USQL;"):
		return p.executeUSQLQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, "SLO;"):
		return p.executeSLOQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, "PV2;"):
		return p.executeProblemQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, "SECPV2;"):
		return p.executeSecurityProblemQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, "MV2;"):
		return p.executeMetricsV2Query(sliQuery, p.startUnix, p.endUnix)
	default:
		return p.executeMetricsQuery(sliQuery, "", p.startUnix, p.endUnix)
	}
}

// USQL query
func (p *Processing) executeUSQLQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	// In this case we need to parse USQL;TYPE;DIMENSION;QUERY
	const resultTypeSingleValue = "SINGLE_VALUE"
	const resultTypeTable = "TABLE"
	const resultTypeColumnChart = "COLUMN_CHART"
	const resultTypePieChart = "PIE_CHART"

	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 4 {
		return 0, fmt.Errorf("USQL Query incorrect format: %s", metricsQuery)
	}

	resultType := querySplits[1]
	if resultType == "" {
		return 0, fmt.Errorf("USQL result type is empty - only %s, %s, %s and %s allowed",
			resultTypeSingleValue, resultTypeColumnChart, resultTypePieChart, resultTypeTable)
	}

	requestedDimensionName := querySplits[2]
	if requestedDimensionName == "" && resultType != resultTypeSingleValue {
		return 0, fmt.Errorf("USQL dimension should not be empty unless result type is %s", resultTypeSingleValue)
	}

	usqlRawQuery := querySplits[3]
	if usqlRawQuery == "" {
		return 0, fmt.Errorf("USQL query is emtpy")
	}

	usqlQuery := usql.NewQueryBuilder(p.eventData, p.customFilters).Build(usqlRawQuery, startUnix, endUnix)
	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(usqlQuery)
	if err != nil {
		return 0, fmt.Errorf("error executing USQL Query: %v", err)
	}

	if resultType == resultTypeSingleValue {
		if len(usqlResult.ColumnNames) != 1 || len(usqlResult.Values) != 1 {
			return 0, fmt.Errorf("USQL result type %s should only return a single result", resultTypeSingleValue)
		}
		return tryCastDimensionValueToNumeric(usqlResult.Values[0][0])
	}

	// all other types must at least have 2 columns to work properly
	if len(usqlResult.ColumnNames) < 2 {
		return 0, fmt.Errorf("USQL result type %s should at least have two columns", resultType)
	}

	for _, rowValue := range usqlResult.Values {
		var dimensionName interface{}
		var dimensionValue interface{}

		switch resultType {
		case resultTypePieChart, resultTypeColumnChart:
			dimensionName = rowValue[0]
			dimensionValue = rowValue[1]
		case resultTypeTable:
			dimensionName = rowValue[0]
			dimensionValue = rowValue[len(rowValue)-1]
		default:
			return 0, fmt.Errorf("unknown USQL result type: %s", resultType)
		}

		name, err := tryCastDimensionNameToString(dimensionName)
		if err != nil {
			return 0, err
		}

		if name == requestedDimensionName {
			return tryCastDimensionValueToNumeric(dimensionValue)
		}
	}

	return 0, fmt.Errorf("could not find dimension name '%s' in result", requestedDimensionName)
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

	return "", errors.New("dimension name should be a string ")
}

// query a specific SLO
func (p *Processing) executeSLOQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {

	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 2 {
		return 0, fmt.Errorf("SLO Indicator query has wrong format. Should be SLO;<SLID> but is: %s", metricsQuery)
	}

	sloID := querySplits[1]
	sloResult, err := dynatrace.NewSLOClient(p.client).Get(sloID, startUnix, endUnix)
	if err != nil {
		return 0, err
	}

	return sloResult.EvaluatedPercentage, nil
}

func (p *Processing) executeProblemQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	// we query number of problems
	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 2 {
		return 0, fmt.Errorf("Problemv2 Indicator query has wrong format. Should be PV2;entitySelectory=selector&problemSelector=selector but is: %s", metricsQuery)
	}

	problemQuery := querySplits[1]
	totalProblemCount, err := dynatrace.NewProblemsV2Client(p.client).GetTotalCountByQuery(problemQuery, startUnix, endUnix)
	if err != nil {
		return 0, fmt.Errorf("Error executing Dynatrace Problem v2 Query %v", err)
	}

	return float64(totalProblemCount), nil
}

//  query number of problems
func (p *Processing) executeSecurityProblemQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {

	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 2 {
		return 0, fmt.Errorf("Security Problemv2 Indicator query has wrong format. Should be SECPV2;securityProblemSelector=selector but is: %s", metricsQuery)
	}

	securityProblemQuery := querySplits[1]
	totalSecurityProblemCount, err := dynatrace.NewSecurityProblemsClient(p.client).GetTotalCountByQuery(securityProblemQuery, startUnix, endUnix)
	if err != nil {
		return 0, err
	}

	return float64(totalSecurityProblemCount), nil
}

func (p *Processing) executeMetricsV2Query(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	metricsQuery, metricUnit, err := unit.ParseMV2Query(metricsQuery)
	if err != nil {
		return 0, err
	}

	return p.executeMetricsQuery(metricsQuery, metricUnit, startUnix, endUnix)
}

func (p *Processing) executeMetricsQuery(metricsQueryString string, metricUnit string, startUnix time.Time, endUnix time.Time) (float64, error) {
	// try to do the legacy query transformation
	transformedQueryString, err := metrics.NewLegacyQueryTransformation(metricsQueryString).Transform()
	if err != nil {
		return 0, fmt.Errorf("could not parse old format metrics query: %v, %w", metricsQueryString, err)
	}

	metricsQuery, err := metrics.ParseQuery(transformedQueryString)
	if err != nil {
		return 0, fmt.Errorf("could not parse metrics query: %v, %w", metricsQuery, err)
	}

	result, err := dynatrace.NewMetricsClient(p.client).GetByQuery(metricsQuery.GetMetricSelector(), metricsQuery.GetEntitySelector(), startUnix, endUnix)

	if err != nil {
		return 0, fmt.Errorf("Dynatrace Metrics API returned an error: %s. This was the query executed: %s", err.Error(), metricsQuery)
	}

	// TODO 2021-10-13: Collect and log all warnings

	// TODO 2021-10-13: Check if having a query result with zero results is even plausable
	if len(result.Result) == 0 {
		return 0, fmt.Errorf("Dynatrace Metrics API failed to return a result")
	}

	if len(result.Result) > 1 {
		return 0, fmt.Errorf("expected only a single result from Dynatrace Metrics API but got multiple")
	}

	singleResult := result.Result[0]

	if len(singleResult.Data) == 0 {
		if len(singleResult.Warnings) > 0 {
			return 0, fmt.Errorf("Dynatrace Metrics API returned zero data points. Warnings: %s, Query: %s", strings.Join(singleResult.Warnings, ", "), metricsQuery)
		}
		return 0, fmt.Errorf("Dynatrace Metrics API returned zero data points")
	}

	if len(singleResult.Data) > 1 {
		if len(singleResult.Warnings) > 0 {
			return 0, fmt.Errorf("expected only a single data point from Dynatrace Metrics API but got multiple. Warnings: %s, Query: %s", strings.Join(singleResult.Warnings, ", "), metricsQuery)
		}
		return 0, fmt.Errorf("expected only a single data point from Dynatrace Metrics API but got multiple")
	}

	singleDataPoint := singleResult.Data[0]

	// TODO 2021-10-13: Check if having a query result with zero values is even plausable
	if len(singleDataPoint.Values) == 0 {
		if len(singleResult.Warnings) > 0 {
			return 0, fmt.Errorf("Dynatrace Metrics API returned zero data point values. Warnings: %s, Query: %s", strings.Join(singleResult.Warnings, ", "), metricsQuery)
		}
		return 0, fmt.Errorf("Dynatrace Metrics API returned zero data point values")
	}

	if len(singleDataPoint.Values) > 1 {
		if len(singleResult.Warnings) > 0 {
			return 0, fmt.Errorf("expected only a single data point value from Dynatrace Metrics API but got multiple. Warnings: %s, Query: %s", strings.Join(singleResult.Warnings, ", "), metricsQuery)
		}
		return 0, fmt.Errorf("expected only a single data point value from Dynatrace Metrics API but got multiple")
	}

	singleValue := singleDataPoint.Values[0]
	return unit.ScaleData(metricsQuery.GetMetricSelector(), metricUnit, singleValue), nil
}

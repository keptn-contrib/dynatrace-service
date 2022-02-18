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
	v1metrics "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/metrics"
	v1mv2 "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/mv2"
	v1problems "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/problemsv2"
	v1secpv2 "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/secpv2"
	v1slo "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/slo"
	v1usql "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/usql"
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

// GetSLIValue queries a single SLI value ultimately from the Dynatrace API.
// TODO: 2022-01-28: Refactoring needed: this is currently SLI v1 format processing, it should moved to the v1 package, separating it from the general logic.
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
	case strings.HasPrefix(sliQuery, v1usql.USQLPrefix):
		return p.executeUSQLQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, v1slo.SLOPrefix):
		return p.executeSLOQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, v1problems.ProblemsV2Prefix):
		return p.executeProblemQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, v1secpv2.SecurityProblemsV2Prefix):
		return p.executeSecurityProblemQuery(sliQuery, p.startUnix, p.endUnix)
	case strings.HasPrefix(sliQuery, v1mv2.MV2Prefix):
		return p.executeMetricsV2Query(sliQuery, p.startUnix, p.endUnix)
	default:
		return p.executeMetricsQuery(sliQuery, p.startUnix, p.endUnix)
	}
}

// USQL query
func (p *Processing) executeUSQLQuery(usqlQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {

	query, err := v1usql.NewQueryParser(usqlQuery).Parse()
	if err != nil {
		return 0, fmt.Errorf("error parsing USQL query: %w", err)
	}

	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(dynatrace.NewUSQLClientQueryParameters(query.GetQuery(), startUnix, endUnix))
	if err != nil {
		return 0, fmt.Errorf("error executing USQL query: %w", err)
	}

	if query.GetResultType() == v1usql.SingleValueResultType {
		if len(usqlResult.ColumnNames) != 1 || len(usqlResult.Values) != 1 {
			return 0, fmt.Errorf("USQL result type %s should only return a single result", v1usql.SingleValueResultType)
		}
		return tryCastDimensionValueToNumeric(usqlResult.Values[0][0])
	}

	// all other types must at least have 2 columns to work properly
	if len(usqlResult.ColumnNames) < 2 {
		return 0, fmt.Errorf("USQL result type %s should at least have two columns", query.GetResultType())
	}

	for _, rowValue := range usqlResult.Values {
		var dimensionName interface{}
		var dimensionValue interface{}

		switch query.GetResultType() {
		case v1usql.PieChartResultType, v1usql.ColumnChartResultType, v1usql.LineChartResultType:
			dimensionName = rowValue[0]
			dimensionValue = rowValue[1]
		case v1usql.TableResultType:
			dimensionName = rowValue[0]
			dimensionValue = rowValue[len(rowValue)-1]
		default:
			return 0, fmt.Errorf("unknown USQL result type: %s", query.GetResultType())
		}

		name, err := tryCastDimensionNameToString(dimensionName)
		if err != nil {
			return 0, err
		}

		if name == query.GetDimension() {
			return tryCastDimensionValueToNumeric(dimensionValue)
		}
	}

	return 0, fmt.Errorf("could not find dimension name '%s' in result", query.GetDimension())
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
func (p *Processing) executeSLOQuery(sloQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	query, err := v1slo.NewQueryParser(sloQuery).Parse()
	if err != nil {
		return 0, fmt.Errorf("error parsing SLO query: %w", err)
	}

	sloResult, err := dynatrace.NewSLOClient(p.client).Get(dynatrace.NewSLOClientGetParameters(query.GetSLOID(), startUnix, endUnix))
	if err != nil {
		return 0, err
	}

	return sloResult.EvaluatedPercentage, nil
}

func (p *Processing) executeProblemQuery(problemsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	query, err := v1problems.NewQueryParser(problemsQuery).Parse()
	if err != nil {
		return 0, fmt.Errorf("error parsing Problems V2 query: %w", err)
	}

	totalProblemCount, err := dynatrace.NewProblemsV2Client(p.client).GetTotalCountByQuery(dynatrace.NewProblemsV2ClientQueryParameters(*query, startUnix, endUnix))
	if err != nil {
		return 0, fmt.Errorf("Error executing Dynatrace Problem v2 Query %v", err)
	}

	return float64(totalProblemCount), nil
}

//  query number of problems
func (p *Processing) executeSecurityProblemQuery(queryString string, startUnix time.Time, endUnix time.Time) (float64, error) {
	query, err := v1secpv2.NewQueryParser(queryString).Parse()
	if err != nil {
		return 0, fmt.Errorf("error parsing Security Problems V2 query: %w", err)
	}

	totalSecurityProblemCount, err := dynatrace.NewSecurityProblemsClient(p.client).GetTotalCountByQuery(dynatrace.NewSecurityProblemsV2ClientQueryParameters(*query, startUnix, endUnix))
	if err != nil {
		return 0, err
	}

	return float64(totalSecurityProblemCount), nil
}

func (p *Processing) executeMetricsV2Query(queryString string, startUnix time.Time, endUnix time.Time) (float64, error) {

	query, err := v1mv2.NewQueryParser(queryString).Parse()
	if err != nil {
		return 0, fmt.Errorf("could not parse MV2 query: %v, %w", queryString, err)
	}

	return p.processMetricsQuery(query.GetQuery(), query.GetUnit(), startUnix, endUnix)
}

func (p *Processing) executeMetricsQuery(queryString string, startUnix time.Time, endUnix time.Time) (float64, error) {
	query, err := v1metrics.NewQueryParser(queryString).Parse()
	if err == nil {
		return p.processMetricsQuery(*query, "", startUnix, endUnix)
	}

	query, legacyErr := v1metrics.NewLegacyQueryParser(queryString).Parse()
	if legacyErr != nil {
		return 0, fmt.Errorf("could not parse metrics query: %v, %w", queryString, err)
	}
	return p.processMetricsQuery(*query, "", startUnix, endUnix)
}

func (p *Processing) processMetricsQuery(query metrics.Query, metricUnit string, startUnix time.Time, endUnix time.Time) (float64, error) {
	result, err := dynatrace.NewMetricsClient(p.client).GetByQuery(dynatrace.NewMetricsClientQueryParameters(query, startUnix, endUnix))
	if err != nil {
		return 0, fmt.Errorf("Dynatrace Metrics API returned an error: %w", err)
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
			return 0, fmt.Errorf("Dynatrace Metrics API returned zero data points. Warnings: %s", strings.Join(singleResult.Warnings, ", "))
		}
		return 0, fmt.Errorf("Dynatrace Metrics API returned zero data points")
	}

	if len(singleResult.Data) > 1 {
		if len(singleResult.Warnings) > 0 {
			return 0, fmt.Errorf("expected only a single data point from Dynatrace Metrics API but got multiple. Warnings: %s", strings.Join(singleResult.Warnings, ", "))
		}
		return 0, fmt.Errorf("expected only a single data point from Dynatrace Metrics API but got multiple")
	}

	singleDataPoint := singleResult.Data[0]

	// TODO 2021-10-13: Check if having a query result with zero values is even plausable
	if len(singleDataPoint.Values) == 0 {
		if len(singleResult.Warnings) > 0 {
			return 0, fmt.Errorf("Dynatrace Metrics API returned zero data point values. Warnings: %s", strings.Join(singleResult.Warnings, ", "))
		}
		return 0, fmt.Errorf("Dynatrace Metrics API returned zero data point values")
	}

	if len(singleDataPoint.Values) > 1 {
		if len(singleResult.Warnings) > 0 {
			return 0, fmt.Errorf("expected only a single data point value from Dynatrace Metrics API but got multiple. Warnings: %s", strings.Join(singleResult.Warnings, ", "))
		}
		return 0, fmt.Errorf("expected only a single data point value from Dynatrace Metrics API but got multiple")
	}

	singleValue := singleDataPoint.Values[0]
	return unit.ScaleData(query.GetMetricSelector(), metricUnit, singleValue), nil
}

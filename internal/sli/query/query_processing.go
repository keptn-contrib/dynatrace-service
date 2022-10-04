package query

import (
	"context"
	"errors"
	"fmt"
	"strings"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/unit"
	v1metrics "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/metrics"
	v1mv2 "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/mv2"
	v1problems "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/problemsv2"
	v1secpv2 "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/secpv2"
	v1slo "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/slo"
	v1usql "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/usql"
)

// Processing representing the processing of custom SLI queries.
type Processing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	customQueries *CustomQueries
	timeframe     common.Timeframe
}

// NewProcessing creates a new Processing.
func NewProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, customQueries *CustomQueries, timeframe common.Timeframe) *Processing {
	return &Processing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		customQueries: customQueries,
		timeframe:     timeframe,
	}
}

// GetSLIResultFromIndicator queries a single SLI value ultimately from the Dynatrace API and returns an SLIResult.
// TODO: 2022-01-28: Refactoring needed: this is currently SLI v1 format processing, it should moved to the v1 package, separating it from the general logic.
func (p *Processing) GetSLIResultFromIndicator(ctx context.Context, name string) result.SLIResult {

	// first we get the query from the SLI configuration based on its logical name
	// no default values here anymore if indicator could not be matched (e.g. due to a misspelling) and custom SLIs were defined
	rawQuery, err := p.customQueries.GetQueryByNameOrDefaultIfEmpty(name)
	if err != nil {
		return result.NewFailedSLIResult(name, err.Error())
	}

	sliQuery := common.ReplaceQueryParameters(rawQuery, p.customFilters, p.eventData)

	switch {
	case strings.HasPrefix(sliQuery, v1usql.USQLPrefix):
		return p.executeUSQLQuery(ctx, name, sliQuery)
	case strings.HasPrefix(sliQuery, v1slo.SLOPrefix):
		return p.executeSLOQuery(ctx, name, sliQuery)
	case strings.HasPrefix(sliQuery, v1problems.ProblemsV2Prefix):
		return p.executeProblemQuery(ctx, name, sliQuery)
	case strings.HasPrefix(sliQuery, v1secpv2.SecurityProblemsV2Prefix):
		return p.executeSecurityProblemQuery(ctx, name, sliQuery)
	case strings.HasPrefix(sliQuery, v1mv2.MV2Prefix):
		return p.executeMetricsV2Query(ctx, name, sliQuery)
	default:
		return p.executeMetricsQuery(ctx, name, sliQuery)
	}
}

func (p *Processing) executeUSQLQuery(ctx context.Context, name string, usqlQuery string) result.SLIResult {

	query, err := v1usql.NewQueryParser(usqlQuery).Parse()
	if err != nil {
		return result.NewFailedSLIResult(name, "error parsing USQL query: "+err.Error())
	}

	request := dynatrace.NewUSQLClientQueryRequest(query.GetQuery(), p.timeframe)
	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(ctx, request)
	if err != nil {
		return result.NewFailedSLIResultWithQuery(name, "error querying User sessions API: "+err.Error(), request.RequestString())
	}

	if query.GetResultType() == v1usql.SingleValueResultType {
		if len(usqlResult.ColumnNames) != 1 || len(usqlResult.Values) != 1 {
			return result.NewWarningSLIResultWithQuery(name, fmt.Sprintf("USQL result type %s should only return a single result", v1usql.SingleValueResultType), request.RequestString())
		}
		value, err := tryCastDimensionValueToNumeric(usqlResult.Values[0][0])
		if err != nil {
			return result.NewWarningSLIResultWithQuery(name, err.Error(), request.RequestString())
		}
		return result.NewSuccessfulSLIResultWithQuery(name, value, request.RequestString())
	}

	// all other types must at least have 2 columns to work properly
	if len(usqlResult.ColumnNames) < 2 {
		return result.NewWarningSLIResultWithQuery(name, fmt.Sprintf("USQL result type %s should at least have two columns", query.GetResultType()), request.RequestString())
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
			// this is unlikely to be reached as it should be handled by the query parser, but a failed result is generated because it is unsupported
			return result.NewFailedSLIResultWithQuery(name, fmt.Sprintf("unknown USQL result type: %s", query.GetResultType()), request.RequestString())
		}

		dimensionNameString, err := tryCastDimensionNameToString(dimensionName)
		if err != nil {
			return result.NewWarningSLIResultWithQuery(name, err.Error(), request.RequestString())
		}

		if dimensionNameString == query.GetDimension() {
			value, err := tryCastDimensionValueToNumeric(dimensionValue)
			if err != nil {
				return result.NewWarningSLIResultWithQuery(name, err.Error(), request.RequestString())
			}
			return result.NewSuccessfulSLIResultWithQuery(name, value, request.RequestString())
		}
	}

	return result.NewWarningSLIResultWithQuery(name, fmt.Sprintf("could not find dimension name '%s' in result", query.GetDimension()), request.RequestString())
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

func (p *Processing) executeSLOQuery(ctx context.Context, name string, sloQuery string) result.SLIResult {
	query, err := v1slo.NewQueryParser(sloQuery).Parse()
	if err != nil {
		return result.NewFailedSLIResult(name, "error parsing SLO query: "+err.Error())
	}

	request := dynatrace.NewSLOClientGetRequest(query.GetSLOID(), p.timeframe)
	sloResult, err := dynatrace.NewSLOClient(p.client).Get(ctx, request)
	if err != nil {
		return result.NewFailedSLIResultWithQuery(name, "error querying Service level objectives API: "+err.Error(), request.RequestString())
	}

	return result.NewSuccessfulSLIResultWithQuery(name, sloResult.EvaluatedPercentage, request.RequestString())
}

func (p *Processing) executeProblemQuery(ctx context.Context, name string, problemsQuery string) result.SLIResult {
	query, err := v1problems.NewQueryParser(problemsQuery).Parse()
	if err != nil {
		return result.NewFailedSLIResult(name, "error parsing Problems v2 query: "+err.Error())
	}

	request := dynatrace.NewProblemsV2ClientQueryRequest(*query, p.timeframe)
	totalProblemCount, err := dynatrace.NewProblemsV2Client(p.client).GetTotalCountByQuery(ctx, request)
	if err != nil {
		return result.NewFailedSLIResultWithQuery(name, "error querying Problems API v2: "+err.Error(), request.RequestString())
	}

	return result.NewSuccessfulSLIResultWithQuery(name, float64(totalProblemCount), request.RequestString())
}

func (p *Processing) executeSecurityProblemQuery(ctx context.Context, name string, queryString string) result.SLIResult {
	query, err := v1secpv2.NewQueryParser(queryString).Parse()
	if err != nil {
		return result.NewFailedSLIResult(name, "error parsing Security Problems v2 query: "+err.Error())
	}

	request := dynatrace.NewSecurityProblemsClientQueryRequest(*query, p.timeframe)
	totalSecurityProblemCount, err := dynatrace.NewSecurityProblemsClient(p.client).GetTotalCountByQuery(ctx, request)
	if err != nil {
		return result.NewFailedSLIResultWithQuery(name, "error querying Security problems API: "+err.Error(), request.RequestString())
	}

	return result.NewSuccessfulSLIResultWithQuery(name, float64(totalSecurityProblemCount), request.RequestString())
}

func (p *Processing) executeMetricsV2Query(ctx context.Context, name string, queryString string) result.SLIResult {
	query, err := v1mv2.NewQueryParser(queryString).Parse()
	if err != nil {
		return result.NewFailedSLIResult(name, "error parsing MV2 query: "+err.Error())
	}

	return p.processMetricsQueryAndMakeSLIResult(ctx, name, query.GetQuery(), query.GetUnit())
}

func (p *Processing) executeMetricsQuery(ctx context.Context, name string, queryString string) result.SLIResult {
	query, err := v1metrics.NewQueryParser(queryString).Parse()
	if err == nil {
		return p.processMetricsQueryAndMakeSLIResult(ctx, name, *query, "")
	}

	query, legacyErr := v1metrics.NewLegacyQueryParser(queryString).Parse()
	if legacyErr != nil {
		return result.NewFailedSLIResult(name, "error parsing Metrics v2 query: "+err.Error())
	}
	return p.processMetricsQueryAndMakeSLIResult(ctx, name, *query, "")
}

func (p *Processing) processMetricsQueryAndMakeSLIResult(ctx context.Context, name string, query metrics.Query, metricUnit string) result.SLIResult {
	request := dynatrace.NewMetricsClientQueryRequest(query, p.timeframe)
	metricsClient := dynatrace.NewMetricsClient(p.client)
	results, err := dynatrace.NewRetryForSingleValueMetricsProcessingDecorator(metricsClient, dynatrace.NewMetricsProcessingThatAllowsOnlyOneResult(metricsClient)).ProcessRequest(ctx, request)
	if err != nil {
		return createSLIResultFromErrorFromMetricsProcessing(err, name, request)
	}

	r, err := results.FirstResultOrError()
	if err != nil {
		return createSLIResultFromErrorFromMetricsProcessing(err, name, request)
	}

	resultsRequest := results.Request()
	return result.NewSuccessfulSLIResultWithQuery(name, unit.ScaleData(metricUnit, r.Value()), resultsRequest.RequestString())
}

func createSLIResultFromErrorFromMetricsProcessing(err error, name string, request dynatrace.MetricsClientQueryRequest) result.SLIResult {
	var qpErrorType *dynatrace.MetricsQueryProcessingError
	if errors.As(err, &qpErrorType) {
		return result.NewWarningSLIResultWithQuery(name, err.Error(), request.RequestString())
	}
	return result.NewFailedSLIResultWithQuery(name, err.Error(), request.RequestString())
}

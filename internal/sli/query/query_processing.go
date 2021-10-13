package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/unit"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
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
	sliQuery, err := p.customQueries.GetQueryByNameOrDefaultIfEmpty(name)
	if err != nil {
		return 0, err
	}

	log.WithFields(
		log.Fields{
			"name":  name,
			"query": sliQuery,
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
	// In this case we need to parse USQL;TILE_TYPE;DIMENSION;QUERY
	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 4 {
		return 0, fmt.Errorf("USQL Query incorrect format: %s", metricsQuery)
	}

	tileName := querySplits[1]
	requestedDimensionName := querySplits[2]
	usqlRawQuery := querySplits[3]

	usqlQuery := usql.NewQueryBuilder(p.eventData, p.customFilters).Build(usqlRawQuery, startUnix, endUnix)
	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(usqlQuery)

	if err != nil {
		return 0, fmt.Errorf("Error executing USQL Query %v", err)
	}

	for _, rowValue := range usqlResult.Values {
		dimensionName := ""
		dimensionValue := 0.0

		if tileName == "SINGLE_VALUE" {
			dimensionValue = rowValue[0].(float64)
		} else if tileName == "PIE_CHART" {
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		} else if tileName == "COLUMN_CHART" {
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		} else if tileName == "TABLE" {
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[len(rowValue)-1].(float64)
		} else {
			log.WithField("tileName", tileName).Debug("Unsupported USQL Tile Type")
			continue
		}

		// did we find the value we were looking for?
		if strings.Compare(dimensionName, requestedDimensionName) == 0 {
			return dimensionValue, nil
		}
	}

	return 0, fmt.Errorf("Error executing USQL Query")
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
	problemQueryResult, err := dynatrace.NewProblemsV2Client(p.client).GetByQuery(problemQuery, startUnix, endUnix)
	if err != nil {
		return 0, fmt.Errorf("Error executing Dynatrace Problem v2 Query %v", err)
	}

	return float64(problemQueryResult.TotalCount), nil
}

//  query number of problems
func (p *Processing) executeSecurityProblemQuery(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {

	querySplits := strings.Split(metricsQuery, ";")
	if len(querySplits) != 2 {
		return 0, fmt.Errorf("Security Problemv2 Indicator query has wrong format. Should be SECPV2;securityProblemSelector=selector but is: %s", metricsQuery)
	}

	problemQuery := querySplits[1]
	problemQueryResult, err := dynatrace.NewSecurityProblemsClient(p.client).GetByQuery(problemQuery, startUnix, endUnix)
	if err != nil {
		return 0, err
	}

	return float64(problemQueryResult.TotalCount), nil
}

func (p *Processing) executeMetricsV2Query(metricsQuery string, startUnix time.Time, endUnix time.Time) (float64, error) {
	metricsQuery, metricUnit, err := unit.ParseMV2Query(metricsQuery)
	if err != nil {
		return 0, err
	}

	return p.executeMetricsQuery(metricsQuery, metricUnit, startUnix, endUnix)
}

func (p *Processing) executeMetricsQuery(metricsQuery string, metricUnit string, startUnix time.Time, endUnix time.Time) (float64, error) {

	metricsQuery, metricSelector, err := metrics.NewQueryBuilder(p.eventData, p.customFilters).Build(metricsQuery, startUnix, endUnix)
	if err != nil {
		return 0, err
	}
	result, err := dynatrace.NewMetricsClient(p.client).GetByQuery(metricsQuery)

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
	return unit.ScaleData(metricSelector, metricUnit, singleValue), nil
}

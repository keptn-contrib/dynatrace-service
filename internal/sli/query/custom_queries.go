package query

import "fmt"

const throughput = "throughput"
const errorRate = "error_rate"
const responseTimeP50 = "response_time_p50"
const responseTimeP90 = "response_time_p90"
const responseTimeP95 = "response_time_p95"

// CustomQueries provides access to user-defined or default SLIs.
type CustomQueries struct {
	values map[string]string
}

// NewEmptyCustomQueries creates a new CustomQueries which will only offer default SLIs.
func NewEmptyCustomQueries() *CustomQueries {
	return &CustomQueries{
		values: make(map[string]string),
	}
}

// NewCustomQueries creates a new CustomQueries which will offer the specified SLIs as well as defaults.
func NewCustomQueries(values map[string]string) *CustomQueries {
	return &CustomQueries{
		values: values,
	}
}

// GetQueryByNameOrDefault returns the custom query with the specified name, a default if available or an error if no such default exists.
func (cq *CustomQueries) GetQueryByNameOrDefault(sliName string) (string, error) {
	query, exists := cq.values[sliName]
	if exists {
		return query, nil
	}

	defaultQuery, err := getDefaultQuery(sliName)
	if err != nil {
		return "", err
	}

	return defaultQuery, nil
}

// GetQueryByNameOrDefaultIfEmpty returns the custom query with the specified name.
// If custom queries have been defined, an error will be returned if an entry for name is not included, or
// if no custom queries have been defined, a default will be returned if available, or an error if no such default exists.
func (cq *CustomQueries) GetQueryByNameOrDefaultIfEmpty(sliName string) (string, error) {
	query, exists := cq.values[sliName]
	if exists {
		return query, nil
	}

	// there are custom SLIs defined, but we could not match it
	if len(cq.values) != 0 {
		return "", fmt.Errorf("SLI definition for '%s' was not found", sliName)
	}

	// no custom SLIs defined - so we fallback to using defaults
	defaultQuery, err := getDefaultQuery(sliName)
	if err != nil {
		return "", err
	}

	return defaultQuery, nil
}

func getDefaultQuery(sliName string) (string, error) {
	// returns new Metrics v2 queries as discussed here: https://github.com/keptn-contrib/dynatrace-sli-service/issues/91
	switch sliName {
	case throughput:
		return "metricSelector=builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case errorRate:
		return "metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP50:
		return "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP90:
		return "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(90)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP95:
		return "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	default:
		return "", fmt.Errorf("unsupported SLI %s", sliName)
	}
}

package metrics

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// url to the metrics api format migration document
const metricsAPIOldFormatNewFormatDoc = "https://github.com/keptn-contrib/dynatrace-sli-service/blob/master/docs/CustomQueryFormatMigration.md"

const (
	scopeKeyPrefix = "scope="
	v1Delimiter    = "?"
	serviceType    = "type(SERVICE)"
)

// LegacyQueryTransformation will parse a un-encoded metric definition query string (usually found in sli.yaml files)
type LegacyQueryTransformation struct {
	query string
}

func NewLegacyQueryTransformation(query string) *LegacyQueryTransformation {
	return &LegacyQueryTransformation{
		query: query,
	}
}

// Transform will try to transform a un-encoded metric definition query string (usually found in sli.yaml files) into a
// valid metrics V2 API query string or return an error in case it could not successfully do that.
// It will return the input query in case it 'assumes' the query is in metrics V2 format.
func (p *LegacyQueryTransformation) Transform() (string, error) {
	query := removeQuestionMark(p.query)

	return transformToMetricsV2QueryFormat(query)
}

func transformToMetricsV2QueryFormat(query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("empty metric selector")
	}

	// support the old format with "<metric_selector>:<some_filters()>?scope=<scope>" as well as the new format with
	// "metricSelector=<metric_selector>&entitySelector=<entity_selector>"
	// split query string by first occurrence of "?"
	querySplit := strings.Split(query, v1Delimiter)

	// new V2 format -> everything within the query string are query parameters
	if len(querySplit) == 1 {
		return query, nil
	}

	log.WithFields(
		log.Fields{
			"query":        query,
			"helpDocument": metricsAPIOldFormatNewFormatDoc,
		}).Warn("COMPATIBILITY WARNING: query uses the old format")

	entitySelector, err := transformScopeToEntitySelector(querySplit[1])
	if err != nil {
		return "", err
	}

	// build the new query format: old format with "?" - everything left of the ? is the identifier, everything right are query params
	return fmt.Sprintf("%s=%s&%s", metricSelectorKey, querySplit[0], entitySelector), nil
}

func transformScopeToEntitySelector(scope string) (string, error) {
	if !strings.HasPrefix(scope, scopeKeyPrefix) {
		return "", fmt.Errorf("invalid metric query - missing 'scope=<scope>' part")
	}

	// compatibility with old scope=... custom queries
	scopeValue := strings.TrimPrefix(scope, scopeKeyPrefix)
	if scopeValue == "" {
		return "", fmt.Errorf("invalid metric query - missing value for 'scope=' key")
	}

	log.WithField("helpDocument", metricsAPIOldFormatNewFormatDoc).Debug("COMPATIBILITY WARNING: querying the new metrics API requires use of entitySelector rather than scope")
	// scope is no longer supported in the new API, it needs to be called "entitySelector" and contain type(SERVICE)
	if !strings.Contains(scopeValue, serviceType) {
		log.WithField("helpDocument", metricsAPIOldFormatNewFormatDoc).Debug("COMPATIBILITY WARNING: Automatically adding type(SERVICE) to entitySelector for compatibility with the new Metrics API")
		scopeValue += "," + serviceType
	}

	return fmt.Sprintf("%s=%s", entitySelectorKey, scopeValue), nil
}

func removeQuestionMark(query string) string {
	if strings.HasPrefix(query, v1Delimiter) {
		log.WithFields(
			log.Fields{
				"query":        query,
				"helpDocument": metricsAPIOldFormatNewFormatDoc,
			}).Warn("COMPATIBILITY WARNING: query string is not compatible. Auto-removing the ? in front.")
		return strings.Replace(query, v1Delimiter, "", 1)
	}

	return query
}

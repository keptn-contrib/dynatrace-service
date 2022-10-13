package metrics

import (
	"errors"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/common"
	log "github.com/sirupsen/logrus"
)

// URL to the metrics api format migration document
const metricsAPIOldFormatNewFormatDoc = "https://github.com/keptn-contrib/dynatrace-sli-service/blob/master/docs/CustomQueryFormatMigration.md"

const (
	scopeKey                = "scope"
	v1Delimiter             = "?"
	serviceType             = "type(SERVICE)"
	metricSelectorKeyPrefix = "metricSelector="
)

// LegacyQueryParser will parse an old format un-encoded metrics query string (usually found in sli.yaml files)
type LegacyQueryParser struct {
	query string
}

// NewLegacyQueryParser creates a new LegacyQueryParser for the specified old format un-encoded metrics query string.
func NewLegacyQueryParser(query string) *LegacyQueryParser {
	return &LegacyQueryParser{
		query: query,
	}
}

// Parse parses the old format with "<metric_selector>:<some_filters()>?scope=<scope>" into a metrics query or returns an error.
// Scope is optional.
func (p *LegacyQueryParser) Parse() (*metrics.Query, error) {
	if p.query == "" {
		return nil, fmt.Errorf("legacy metrics query should not be empty")
	}

	// split query string by first occurrence of "?"
	querySplit := strings.SplitN(p.query, v1Delimiter, 2)

	metricSelector, err := parseLegacySelector(querySplit[0])
	if err != nil {
		return nil, err
	}

	// if only one piece, the whole thing must be the metric selector
	if len(querySplit) == 1 {
		return p.createMetricsQueryAndLog(metricSelector, "")
	}

	entitySelector, err := parsePotentialLegacyScope(querySplit[1])
	if err != nil {
		return nil, err
	}

	return p.createMetricsQueryAndLog(metricSelector, entitySelector)
}

// createMetricsQueryAndLog creates a metrics query and logs a compatibility warning or returns an error.
func (p *LegacyQueryParser) createMetricsQueryAndLog(metricSelector string, entitySelector string) (*metrics.Query, error) {
	query, err := metrics.NewQuery(metricSelector, entitySelector, metrics.ResolutionInf, "")
	if err != nil {
		return nil, err
	}

	log.WithFields(
		log.Fields{
			"oldQuery":     p.query,
			"newQuery":     NewQueryProducer(*query).Produce(),
			"helpDocument": metricsAPIOldFormatNewFormatDoc,
		}).Warn("COMPATIBILITY WARNING: query uses the old format")

	return query, nil
}

// parseLegacySelector tries to parse a legacy selector into a metric selector or returns an error.
func parseLegacySelector(selector string) (string, error) {
	// legacy selector is a just the plain selector, not a key=value pair(s)
	// if it can be parsed as key=value(s) then it is not a legacy selector
	keyValuePairs, err := common.NewSLIParser(selector, &legacyMetricSelectorKeyValidator{}).Parse()
	if err == nil && keyValuePairs.Count() > 0 {
		return "", errors.New("metric selector is not legacy format")
	}
	return selector, nil
}

type legacyMetricSelectorKeyValidator struct{}

// ValidateKey returns true, allowing any key.
func (p *legacyMetricSelectorKeyValidator) ValidateKey(key string) bool {
	return true
}

// parsePotentialLegacyScope parses a potential legacy scope into an entity selector or returns an error.
func parsePotentialLegacyScope(potentialScope string) (string, error) {
	if potentialScope == "" {
		return "", nil
	}

	keyValuePairs, err := common.NewSLIParser(potentialScope, &legacyMetricSelectorKeyValidator{}).Parse()
	if err != nil {
		return "", err
	}

	scope := keyValuePairs.GetValue(scopeKey)
	if !strings.Contains(scope, serviceType) {
		log.WithField("helpDocument", metricsAPIOldFormatNewFormatDoc).Debug("COMPATIBILITY WARNING: Automatically adding type(SERVICE) to entitySelector for compatibility with the new Metrics API")
		scope += "," + serviceType
	}
	return scope, nil
}

type legacyEntitySelectorKeyValidator struct{}

// ValidateKey returns true for key "scope", otherwise false.
func (p *legacyEntitySelectorKeyValidator) ValidateKey(key string) bool {
	return key == scopeKey
}

package dynatrace

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// MetricsQueryFailedError represents an error for a metrics query that could not be retrieved because of an error.
type MetricsQueryFailedError struct {
	cause error
}

// Error returns a string representation of this error.
func (e *MetricsQueryFailedError) Error() string {
	return fmt.Sprintf("error querying Metrics API v2: %v", e.cause)
}

// Unwrap returns the cause of the MetricsQueryFailedError.
func (e *MetricsQueryFailedError) Unwrap() error {
	return e.cause
}

// MetricsQueryProcessingError represents an error that occurred while processing metrics query results.
type MetricsQueryProcessingError struct {
	Message  string
	Warnings []string
}

// Error returns a string representation of this error.
func (e *MetricsQueryProcessingError) Error() string {
	return appendOptionalWarningsToMessage(e.Message, e.Warnings)
}

// MetricsQueryProcessingError represents the specific error that a metrics query returned multiple values.
type MetricsQueryReturnedMultipleValuesError struct {
	ValueCount int
	Warnings   []string
}

// Error returns a string representation of this error.
func (e *MetricsQueryReturnedMultipleValuesError) Error() string {
	return appendOptionalWarningsToMessage(fmt.Sprintf("Metrics API v2 returned %d values", e.ValueCount), e.Warnings)
}

func appendOptionalWarningsToMessage(message string, warnings []string) string {
	if len(warnings) > 0 {
		return fmt.Sprintf("%s. Warnings: %s", message, strings.Join(warnings, ", "))
	}
	return message
}

// MetricsProcessingResultSet groups processing results with warning that occurred.
type MetricsProcessingResultSet struct {
	results  []MetricsProcessingResult
	warnings []string
}

func newMetricsProcessingResultsSet(results []MetricsProcessingResult, warnings []string) *MetricsProcessingResultSet {
	return &MetricsProcessingResultSet{
		results:  results,
		warnings: warnings,
	}
}

// Results gets the results of the MetricsProcessingResultSet.
func (s *MetricsProcessingResultSet) Results() []MetricsProcessingResult {
	return s.results
}

// Warnings gets the warnings of the MetricsProcessingResultSet.
func (s *MetricsProcessingResultSet) Warnings() []string {
	return s.warnings
}

// MetricsProcessingResult associates a value with a name derived from a specific set of dimension values.
type MetricsProcessingResult struct {
	name  string
	value float64
}

func newMetricsProcessingResult(name string, value float64) MetricsProcessingResult {
	return MetricsProcessingResult{name: name, value: value}
}

// Name gets the name of the MetricsProcessingResult.
func (r *MetricsProcessingResult) Name() string {
	return r.name
}

// Value gets the value of the MetricsProcessingResult.
func (r *MetricsProcessingResult) Value() float64 {
	return r.value
}

// MetricsProcessingInterface defines processing of a request into results.
type MetricsProcessingInterface interface {
	// ProcessRequest gets a MetricsProcessingResultSet by query or returns an error.
	ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResultSet, error)
}

// MetricsProcessing offers basic retrieval and processing of metrics.
type MetricsProcessing struct {
	client ClientInterface
}

// NewMetricsProcessing creates a new MetricsProcessing using the specified client interface.
func NewMetricsProcessing(client ClientInterface) *MetricsProcessing {
	return &MetricsProcessing{
		client: client,
	}
}

// ProcessRequest queries and processes metrics using the specified request. It checks for a single metric series collection, and transforms each metric series into a result with a name derived from its dimension values. Each metric series must have exactly one value.
func (p *MetricsProcessing) ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResultSet, error) {
	mc := NewMetricsClient(p.client)
	metricData, err := mc.GetMetricDataByQuery(ctx, request)
	if err != nil {
		return nil, &MetricsQueryFailedError{cause: err}
	}

	if len(metricData.Result) == 0 {
		return nil, &MetricsQueryProcessingError{Message: "Metrics API v2 returned zero metric series collections"}
	}

	if len(metricData.Result) > 1 {
		return nil, &MetricsQueryProcessingError{Message: fmt.Sprintf("Metrics API v2 returned %d metric series collections", len(metricData.Result))}
	}

	return processMetricSeriesCollection(metricData.Result[0])
}

func processMetricSeriesCollection(metricSeriesCollection MetricSeriesCollection) (*MetricsProcessingResultSet, error) {
	if len(metricSeriesCollection.Data) == 0 {
		return nil, &MetricsQueryProcessingError{Message: "Metrics API v2 returned zero metric series", Warnings: metricSeriesCollection.Warnings}
	}

	results := make([]MetricsProcessingResult, 0, len(metricSeriesCollection.Data))
	for _, metricSeries := range metricSeriesCollection.Data {
		value, err := processValues(metricSeries.Values, metricSeriesCollection.Warnings)
		if err != nil {
			return nil, err
		}
		results = append(results, newMetricsProcessingResult(generateResultName(metricSeries.DimensionMap), value))
	}
	return newMetricsProcessingResultsSet(results, metricSeriesCollection.Warnings), nil

}

func processValues(values []*float64, warnings []string) (float64, error) {
	if len(values) == 0 {
		return 0, &MetricsQueryProcessingError{Message: "Metrics API v2 returned zero values", Warnings: warnings}
	}

	if len(values) > 1 {
		return 0, &MetricsQueryReturnedMultipleValuesError{
			ValueCount: len(values),
			Warnings:   warnings,
		}
	}

	if values[0] == nil {
		return 0, &MetricsQueryProcessingError{Message: "Metrics API v2 returned 'null' as value", Warnings: warnings}
	}

	return *values[0], nil
}

// generateResultName generates a result name based on all dimensions.
// As this is used for both indicator and display names, it must then be cleaned before use in indicator names.
func generateResultName(dimensionMap map[string]string) string {
	const nameSuffix = ".name"

	// take all dimension values except where both names and IDs are available, in that case only take the names
	suffixComponents := map[string]string{}
	for key, value := range dimensionMap {
		if value == "" {
			continue
		}

		if strings.HasSuffix(key, nameSuffix) {
			keyWithoutNameSuffix := strings.TrimSuffix(key, nameSuffix)
			suffixComponents[keyWithoutNameSuffix] = value
			continue
		}

		_, found := suffixComponents[key]
		if !found {
			suffixComponents[key] = value
		}
	}

	// ensure suffix component values are ordered by key alphabetically
	keys := maps.Keys(suffixComponents)
	sort.Strings(keys)
	sortedSuffixComponentValues := make([]string, 0, len(keys))
	for _, k := range keys {
		sortedSuffixComponentValues = append(sortedSuffixComponentValues, suffixComponents[k])
	}

	return strings.Join(sortedSuffixComponentValues, " ")
}

package dynatrace

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

// MetricsQueryFailedError represents an error for a metrics query that could not be retrieved because of an error.
type MetricsQueryFailedError struct {
	cause error
}

// Error returns a string representation of the MetricsQueryFailedError.
func (e *MetricsQueryFailedError) Error() string {
	return fmt.Sprintf("error querying Metrics API v2: %v", e.cause)
}

// Unwrap returns the cause of the MetricsQueryFailedError.
func (e *MetricsQueryFailedError) Unwrap() error {
	return e.cause
}

// MetricsQueryProcessingError represents an error that occurred while processing metrics query results.
type MetricsQueryProcessingError struct {
	cause error
}

// Unwrap returns the cause of the MetricsQueryProcessingError.
func (e *MetricsQueryProcessingError) Unwrap() error {
	return e.cause
}

// Error returns a string representation of the MetricsQueryProcessingError.
func (e *MetricsQueryProcessingError) Error() string {
	return fmt.Sprintf("could not process metrics query: %v", e.cause)
}

// MetricsQueryReturnedWrongNumberOfMetricSeriesCollectionsError represents the specific error that a metrics query returned the wrong number of (not exactly one) metric series collection.
type MetricsQueryReturnedWrongNumberOfMetricSeriesCollectionsError struct {
	metricSeriesCollectionCount int
}

// Error returns a string representation of the MetricsQueryReturnedWrongNumberOfMetricSeriesCollectionsError.
func (e *MetricsQueryReturnedWrongNumberOfMetricSeriesCollectionsError) Error() string {
	if e.metricSeriesCollectionCount == 0 {
		return "Metrics API v2 returned zero metric series collections"
	}

	return fmt.Sprintf("Metrics API v2 returned %d metric series collections", e.metricSeriesCollectionCount)
}

// MetricsQueryReturnedZeroMetricSeriesError represents the specific error that a metrics query returned zero metric series.
type MetricsQueryReturnedZeroMetricSeriesError struct {
	Warnings []string
}

// Error returns a string representation of the MetricsQueryReturnedZeroMetricSeriesError.
func (e *MetricsQueryReturnedZeroMetricSeriesError) Error() string {
	return appendOptionalWarningsToMessage("Metrics API v2 returned zero metric series", e.Warnings)
}

// MetricsQueryReturnedMultipleMetricSeriesError represents the specific error that a metrics query returned multiple metric series when only one was allowed.
type MetricsQueryReturnedMultipleMetricSeriesError struct {
	SeriesCount int
	Warnings    []string
}

// Error returns a string representation of the MetricsQueryReturnedMultipleMetricSeriesError.
func (e *MetricsQueryReturnedMultipleMetricSeriesError) Error() string {
	return appendOptionalWarningsToMessage(fmt.Sprintf("Metrics API v2 returned %d metric series but only one is supported", e.SeriesCount), e.Warnings)
}

// MetricsQueryReturnedZeroValuesError represents the specific error that a metrics query returned zero values.
type MetricsQueryReturnedZeroValuesError struct {
	Warnings []string
}

// Error returns a string representation of this error.
func (e *MetricsQueryReturnedZeroValuesError) Error() string {
	return appendOptionalWarningsToMessage("Metrics API v2 returned zero values", e.Warnings)
}

// MetricsQueryReturnedMultipleValuesError represents the specific error that a metrics query returned multiple values.
type MetricsQueryReturnedMultipleValuesError struct {
	ValueCount int
	Warnings   []string
}

// Error returns a string representation of the MetricsQueryReturnedMultipleValuesError.
func (e *MetricsQueryReturnedMultipleValuesError) Error() string {
	return appendOptionalWarningsToMessage(fmt.Sprintf("Metrics API v2 returned %d values but only one is supported", e.ValueCount), e.Warnings)
}

// MetricsQueryReturnedNullValueError represents the specific error that a metrics query returned null values.
type MetricsQueryReturnedNullValueError struct {
	Warnings []string
}

// Error returns a string representation of the MetricsQueryReturnedNullValueError.
func (e *MetricsQueryReturnedNullValueError) Error() string {
	return appendOptionalWarningsToMessage("Metrics API v2 returned 'null' as value", e.Warnings)
}

func appendOptionalWarningsToMessage(message string, warnings []string) string {
	if len(warnings) > 0 {
		return fmt.Sprintf("%s. Warnings: %s", message, strings.Join(warnings, ", "))
	}
	return message
}

// MetricsProcessingResults associates processing results with any warnings that occurred.
type MetricsProcessingResults struct {
	request  MetricsClientQueryRequest
	results  []MetricsProcessingResult
	warnings []string
}

func newMetricsProcessingResults(request MetricsClientQueryRequest, results []MetricsProcessingResult, warnings []string) *MetricsProcessingResults {
	return &MetricsProcessingResults{
		request:  request,
		results:  results,
		warnings: warnings,
	}
}

// Request gets the request of the MetricsProcessingResults
func (r *MetricsProcessingResults) Request() MetricsClientQueryRequest {
	return r.request
}

// Results gets the results of the MetricsProcessingResults.
func (r *MetricsProcessingResults) Results() []MetricsProcessingResult {
	return r.results
}

// FirstResultOrError returns the first result or an error if there are no results.
func (r *MetricsProcessingResults) FirstResultOrError() (*MetricsProcessingResult, error) {
	if len(r.results) == 0 {
		return nil, errors.New("metrics processing yields no result")
	}

	return &r.results[0], nil
}

// Warnings gets any warnings associated with the MetricsProcessingResults.
func (r *MetricsProcessingResults) Warnings() []string {
	return r.warnings
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
	// ProcessRequest processes a request into results or returns an error.
	ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResults, error)
}

// MetricsProcessing is an implementation of MetricsProcessingInterface.
type MetricsProcessing struct {
	metricsClient      MetricsClientInterface
	allowOnlyOneResult bool
}

// NewMetricsProcessingThatAllowsMultipleResults creates a new MetricsProcessing that allows multiple results using the specified client interface.
func NewMetricsProcessingThatAllowsMultipleResults(metricsClient MetricsClientInterface) *MetricsProcessing {
	return &MetricsProcessing{
		metricsClient:      metricsClient,
		allowOnlyOneResult: false,
	}
}

// NewMetricsProcessing creates a new MetricsProcessing that only returns a single result using the specified client interface.
// If the query processed returns more than one metric series, i.e. results, an error is returned.
func NewMetricsProcessingThatAllowsOnlyOneResult(metricsClient MetricsClientInterface) *MetricsProcessing {
	return &MetricsProcessing{
		metricsClient:      metricsClient,
		allowOnlyOneResult: true,
	}
}

// ProcessRequest queries and processes metrics using the specified request. It checks for a single metric series collection, and transforms each metric series into a result with a name derived from its dimension values. Each metric series must have exactly one value.
func (p *MetricsProcessing) ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResults, error) {
	metricData, err := p.metricsClient.GetMetricDataByQuery(ctx, request)
	if err != nil {
		return nil, &MetricsQueryFailedError{cause: err}
	}

	results, err := p.processMetricSeriesCollections(request, metricData.Result)
	if err != nil {
		return nil, &MetricsQueryProcessingError{cause: err}
	}

	return results, nil
}

func (p *MetricsProcessing) processMetricSeriesCollections(request MetricsClientQueryRequest, collections []MetricSeriesCollection) (*MetricsProcessingResults, error) {
	if len(collections) != 1 {
		return nil, &MetricsQueryReturnedWrongNumberOfMetricSeriesCollectionsError{metricSeriesCollectionCount: len(collections)}
	}

	return p.processMetricSeriesCollection(request, collections[0])
}

func (p *MetricsProcessing) processMetricSeriesCollection(request MetricsClientQueryRequest, metricSeriesCollection MetricSeriesCollection) (*MetricsProcessingResults, error) {
	if len(metricSeriesCollection.Data) == 0 {
		return nil, &MetricsQueryReturnedZeroMetricSeriesError{Warnings: metricSeriesCollection.Warnings}
	}

	if p.allowOnlyOneResult && len(metricSeriesCollection.Data) > 1 {
		return nil, &MetricsQueryReturnedMultipleMetricSeriesError{SeriesCount: len(metricSeriesCollection.Data), Warnings: metricSeriesCollection.Warnings}
	}

	results := make([]MetricsProcessingResult, 0, len(metricSeriesCollection.Data))
	for _, metricSeries := range metricSeriesCollection.Data {
		value, err := processValues(metricSeries.Values, metricSeriesCollection.Warnings)
		if err != nil {
			return nil, err
		}
		results = append(results, newMetricsProcessingResult(generateResultName(metricSeries.DimensionMap), value))
	}
	return newMetricsProcessingResults(request, results, metricSeriesCollection.Warnings), nil
}

func processValues(values []*float64, warnings []string) (float64, error) {
	if len(values) == 0 {
		return 0, &MetricsQueryReturnedZeroValuesError{Warnings: warnings}
	}

	if len(values) > 1 {
		return 0, &MetricsQueryReturnedMultipleValuesError{
			ValueCount: len(values),
			Warnings:   warnings,
		}
	}

	if values[0] == nil {
		return 0, &MetricsQueryReturnedNullValueError{Warnings: warnings}
	}

	return *values[0], nil
}

// generateResultName generates a result name based on all dimensions.
// No cleaning is performed here, so it must be cleaned before use e.g. in indicator names.
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

// MetricsProcessingDecorator is an implementation of MetricsProcessingInterface.
type MetricsProcessingDecorator struct {
	metricsClient     MetricsClientInterface
	targetUnitID      string
	metricsProcessing MetricsProcessingInterface
}

// NewRetryForSingleValueMetricsProcessingDecorator creates a MetricsProcessingDecorator that modifies and retries the query to try to obtain a single value for each metric series.
func NewRetryForSingleValueMetricsProcessingDecorator(metricsClient MetricsClientInterface, metricsProcessing MetricsProcessingInterface) *MetricsProcessingDecorator {
	return &MetricsProcessingDecorator{
		metricsClient:     metricsClient,
		targetUnitID:      noneUnitID,
		metricsProcessing: metricsProcessing,
	}
}

// NewConvertUnitsAndRetryForSingleValueMetricsProcessingDecorator creates a MetricsProcessingDecorator that modifies the query to obtain a single value for each metric series converted to the specified unit.
func NewConvertUnitsAndRetryForSingleValueMetricsProcessingDecorator(metricsClient MetricsClientInterface, targetUnitID string, metricsProcessing MetricsProcessingInterface) *MetricsProcessingDecorator {
	return &MetricsProcessingDecorator{
		metricsClient:     metricsClient,
		targetUnitID:      targetUnitID,
		metricsProcessing: metricsProcessing,
	}
}

func (p *MetricsProcessingDecorator) ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResults, error) {
	metricsQueryModifier := newMetricsQueryModifier(p.metricsClient, request.query)

	err := metricsQueryModifier.applyUnitConversion(ctx, p.targetUnitID)
	if err != nil {
		return nil, err
	}

	modifiedQuery, err := metricsQueryModifier.getModifiedQuery()
	if err != nil {
		return nil, err
	}

	results, err := p.metricsProcessing.ProcessRequest(ctx, NewMetricsClientQueryRequest(*modifiedQuery, request.timeframe))
	if err == nil {
		return results, nil
	}

	var qrmvErrorType *MetricsQueryReturnedMultipleValuesError
	if !errors.As(err, &qrmvErrorType) {
		return nil, err
	}

	err = metricsQueryModifier.applyFoldOrResolutionInf(ctx)
	if err != nil {
		return nil, err
	}

	modifiedQuery, err = metricsQueryModifier.getModifiedQuery()
	if err != nil {
		return nil, err
	}

	return p.metricsProcessing.ProcessRequest(ctx, NewMetricsClientQueryRequest(*modifiedQuery, request.timeframe))
}

package dynatrace

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/metrics"
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

func (r *MetricsProcessingResults) Request() MetricsClientQueryRequest {
	return r.request
}

// Results gets the results of the MetricsProcessingResult.
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

// MetricsProcessing is an implementation of MetricsProcessingInterface offers basic retrieval and processing of metrics.
type MetricsProcessing struct {
	client             MetricsClientInterface
	allowOnlyOneResult bool
}

// NewMetricsProcessing creates a new MetricsProcessing using the specified client interface.
func NewMetricsProcessing(client MetricsClientInterface) *MetricsProcessing {
	return &MetricsProcessing{
		client:             client,
		allowOnlyOneResult: false,
	}
}

// NewMetricsProcessing creates a new MetricsProcessing that only returns a single result using the specified client interface.
// If the query processed returns more than one metric series, i.e. results, an error is returned.
func NewMetricsProcessingThatAllowsOnlyOneResult(client MetricsClientInterface) *MetricsProcessing {
	return &MetricsProcessing{
		client:             client,
		allowOnlyOneResult: true,
	}
}

// ProcessRequest queries and processes metrics using the specified request. It checks for a single metric series collection, and transforms each metric series into a result with a name derived from its dimension values. Each metric series must have exactly one value.
func (p *MetricsProcessing) ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResults, error) {
	metricData, err := p.client.GetMetricDataByQuery(ctx, request)
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

// RetryForSingleValueMetricsProcessingDecorator decorates MetricsProcessing by modifying the request in cases where multiple values are returned.
type RetryForSingleValueMetricsProcessingDecorator struct {
	client            MetricsClientInterface
	metricsProcessing MetricsProcessingInterface
}

// NewRetryForSingleValueMetricsProcessingDecorator creates a new RetryForSingleValueMetricsProcessingDecorator using the specified client interface and underlying metrics processing interface.
func NewRetryForSingleValueMetricsProcessingDecorator(client MetricsClientInterface, metricsProcessing MetricsProcessingInterface) *RetryForSingleValueMetricsProcessingDecorator {
	return &RetryForSingleValueMetricsProcessingDecorator{
		client:            client,
		metricsProcessing: metricsProcessing,
	}
}

// ProcessRequest queries and processes metrics using the specified request.
func (p *RetryForSingleValueMetricsProcessingDecorator) ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResults, error) {
	resultSet, err := p.metricsProcessing.ProcessRequest(ctx, request)
	if err == nil {
		return resultSet, nil
	}

	var qrmvErrorType *MetricsQueryReturnedMultipleValuesError
	if !errors.As(err, &qrmvErrorType) {
		return nil, err
	}

	modifiedQuery, err := p.modifyQuery(ctx, request.query)
	if err != nil {
		return nil, fmt.Errorf("could not modify query to produce single value: %w", err)
	}

	return p.metricsProcessing.ProcessRequest(ctx, NewMetricsClientQueryRequest(*modifiedQuery, request.timeframe))
}

// modifyQuery modifies the supplied metrics query such that it should return a single value for each set of dimension values.
// First, it tries to set resolution to Inf if resolution hasn't already been set and it is supported. Otherwise, it tries to do an auto fold if this wouldn't use value.
// Other cases will produce an error, which should be bubbled up to the user to instruct them to fix their tile or query.
func (p *RetryForSingleValueMetricsProcessingDecorator) modifyQuery(ctx context.Context, existingQuery metrics.Query) (*metrics.Query, error) {
	// resolution Inf returning multiple values would indicate a broken API (so unlikely), but check for completeness
	if strings.EqualFold(existingQuery.GetResolution(), metrics.ResolutionInf) {
		return nil, errors.New("not possible to modify query with resolution Inf")
	}

	metricSelector := existingQuery.GetMetricSelector()
	metricDefinition, err := p.client.GetMetricDefinitionByID(ctx, metricSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition for metric: %s", metricSelector)
	}

	if metricDefinition.ResolutionInfSupported && (existingQuery.GetResolution() == "") {
		return metrics.NewQuery(metricSelector, existingQuery.GetEntitySelector(), metrics.ResolutionInf, existingQuery.GetMZSelector())
	}

	if metricDefinition.DefaultAggregation.Type == AggregationTypeValue {
		return nil, errors.New("unable to apply ':fold()' to the metric selector as the default aggregation type is 'value'")
	}

	return metrics.NewQuery("("+metricSelector+"):fold()", existingQuery.GetEntitySelector(), existingQuery.GetResolution(), existingQuery.GetMZSelector())
}

// ConvertUnitMetricsProcessingDecorator decorates MetricsProcessing by converting the unit of the results.
type ConvertUnitMetricsProcessingDecorator struct {
	metricSelectorUnitsModifier *metricSelectorUnitsModifier
	targetUnitID                string
	metricsProcessing           MetricsProcessingInterface
}

// NewConvertUnitMetricsProcessingDecorator creates a new ConvertUnitMetricsProcessingDecorator using the specified client interfaces, target unit ID and underlying metrics processing interface.
func NewConvertUnitMetricsProcessingDecorator(metricsClient MetricsClientInterface,
	targetUnitID string,
	metricsProcessing MetricsProcessingInterface) *ConvertUnitMetricsProcessingDecorator {
	return &ConvertUnitMetricsProcessingDecorator{
		metricSelectorUnitsModifier: newMetricSelectorUnitsModifier(metricsClient),
		targetUnitID:                targetUnitID,
		metricsProcessing:           metricsProcessing,
	}
}

// ProcessRequest queries and processes metrics using the specified request.
func (p *ConvertUnitMetricsProcessingDecorator) ProcessRequest(ctx context.Context, request MetricsClientQueryRequest) (*MetricsProcessingResults, error) {

	request, err := p.magicallyFixRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	return p.metricsProcessing.ProcessRequest(ctx, request)
}

func (p *ConvertUnitMetricsProcessingDecorator) magicallyFixRequest(ctx context.Context, request MetricsClientQueryRequest) (MetricsClientQueryRequest, error) {
	if !doesTargetUnitRequireConversion(p.targetUnitID) {
		return request, nil
	}

	fixedQuery, err := p.modifyQueryForTargetUnit(ctx, request.query)
	if err != nil {
		return request, err
	}

	return NewMetricsClientQueryRequest(*fixedQuery, request.timeframe), nil
}

func (p *ConvertUnitMetricsProcessingDecorator) modifyQueryForTargetUnit(ctx context.Context, query metrics.Query) (*metrics.Query, error) {
	modifiedMetricSelector, err := p.metricSelectorUnitsModifier.applyUnit(ctx, query.GetMetricSelector(), p.targetUnitID)
	if err != nil {
		return nil, err
	}

	return metrics.NewQuery(modifiedMetricSelector, query.GetEntitySelector(), query.GetResolution(), query.GetMZSelector())
}

const (
	emptyUnitID = ""
	autoUnitID  = "auto"
	noneUnitID  = "none"
)

// doesTargetUnitRequireConversion checks if the target unit ID requires conversion or not. Currently, "Auto" (default empty value and explicit `auto` value) and "None" require no conversion.
func doesTargetUnitRequireConversion(targetUnitID string) bool {
	switch targetUnitID {
	case emptyUnitID, autoUnitID, noneUnitID:
		return false
	default:
		return true
	}
}

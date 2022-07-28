package result

// IndicatorResultType represents the type of indicator result, i.e. success, warning or fail.
type IndicatorResultType string

const (
	// IndicatorResultSuccessful represents indicator result of success.
	IndicatorResultSuccessful IndicatorResultType = "success"

	// IndicatorResultWarning represents an indicator result of warning.
	IndicatorResultWarning IndicatorResultType = "warning"

	// IndicatorResultFailed represents indicator result of fail.
	IndicatorResultFailed IndicatorResultType = "fail"
)

// SLIResult encapsulates a Keptn SLIResult with an additional result of success, warning or fail.
type SLIResult struct {
	metric          string
	value           float64
	success         bool
	message         string
	query           string
	indicatorResult IndicatorResultType
}

// NewSuccessfulSLIResult creates a new SLIResult with result of success.
func NewSuccessfulSLIResult(metric string, value float64) SLIResult {
	return NewSuccessfulSLIResultWithMessage(metric, value, "")
}

// NewSuccessfulSLIResult creates a new SLIResult with a message and a result of success.
func NewSuccessfulSLIResultWithMessage(metric string, value float64, message string) SLIResult {
	return SLIResult{
		metric:          metric,
		success:         true,
		value:           value,
		message:         message,
		indicatorResult: IndicatorResultSuccessful,
	}
}

// NewWarningSLIResult creates a new SLIResult with result of warning.
func NewWarningSLIResult(metric string, message string) SLIResult {
	return SLIResult{
		metric:          metric,
		success:         false,
		message:         message,
		indicatorResult: IndicatorResultWarning,
	}
}

// NewFailedSLIResult creates a new SLIResult with result of fail.
func NewFailedSLIResult(metric string, message string) SLIResult {
	return SLIResult{
		metric:          metric,
		success:         false,
		message:         message,
		indicatorResult: IndicatorResultFailed,
	}
}

// Metric gets the metric.
func (r SLIResult) Metric() string {
	return r.metric
}

// Value gets the value.
func (r SLIResult) Value() float64 {
	return r.value
}

// Success gets the success.
func (r SLIResult) Success() bool {
	return r.success
}

// Message gets the message.
func (r SLIResult) Message() string {
	return r.message
}

// Query gets the query.
func (r SLIResult) Query() string {
	return r.query
}

// IndicatorResult gets the indicator result, i.e. pass, warning or fail.
func (r SLIResult) IndicatorResult() IndicatorResultType {
	return r.indicatorResult
}

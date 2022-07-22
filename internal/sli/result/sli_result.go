package result

import keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"

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
	keptnResult     keptnv2.SLIResult
	indicatorResult IndicatorResultType
}

// NewSuccessfulSLIResult creates a new SLIResult with result of success.
func NewSuccessfulSLIResult(metric string, value float64) SLIResult {
	return NewSuccessfulSLIResultWithMessage(metric, value, "")
}

// NewSuccessfulSLIResult creates a new SLIResult with a message and a result of success.
func NewSuccessfulSLIResultWithMessage(metric string, value float64, message string) SLIResult {
	return SLIResult{
		keptnResult: keptnv2.SLIResult{
			Metric:  metric,
			Success: true,
			Value:   value,
			Message: message,
		},
		indicatorResult: IndicatorResultSuccessful,
	}
}

// NewWarningSLIResult creates a new SLIResult with result of warning.
func NewWarningSLIResult(metric string, message string) SLIResult {
	return SLIResult{
		keptnResult: keptnv2.SLIResult{
			Metric:  metric,
			Success: false,
			Message: message,
		},
		indicatorResult: IndicatorResultWarning,
	}
}

// NewFailedSLIResult creates a new SLIResult with result of fail.
func NewFailedSLIResult(metric string, message string) SLIResult {
	return SLIResult{
		keptnResult: keptnv2.SLIResult{
			Metric:  metric,
			Success: false,
			Message: message,
		},
		indicatorResult: IndicatorResultFailed,
	}
}

// Metric gets the metric.
func (r SLIResult) Metric() string {
	return r.keptnResult.Metric
}

// Value gets the value.
func (r SLIResult) Value() float64 {
	return r.keptnResult.Value
}

// Success gets the success.
func (r SLIResult) Success() bool {
	return r.keptnResult.Success
}

// Message gets the message.
func (r SLIResult) Message() string {
	return r.keptnResult.Message
}

// KeptnSLIResult gets the wrapped Keptn SLIResult.
func (r SLIResult) KeptnSLIResult() keptnv2.SLIResult {
	return r.keptnResult
}

// IndicatorResult gets the indicator result, i.e. pass, warning or fail.
func (r SLIResult) IndicatorResult() IndicatorResultType {
	return r.indicatorResult
}

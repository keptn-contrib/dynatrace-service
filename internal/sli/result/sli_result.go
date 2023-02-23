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
	Metric          string
	Value           float64
	Success         bool
	Message         string
	Query           string
	IndicatorResult IndicatorResultType
}

// NewSuccessfulSLIResultWithQuery creates a new SLIResult with a query and a result of success.
func NewSuccessfulSLIResultWithQuery(metric string, value float64, query string) SLIResult {
	return SLIResult{
		Metric:          metric,
		Success:         true,
		Value:           value,
		Query:           query,
		IndicatorResult: IndicatorResultSuccessful,
	}
}

// NewWarningSLIResultWithQuery creates a new SLIResult with a query and a result of warning.
func NewWarningSLIResultWithQuery(metric string, message string, query string) SLIResult {
	return SLIResult{
		Metric:          metric,
		Success:         false,
		Message:         message,
		Query:           query,
		IndicatorResult: IndicatorResultWarning,
	}
}

// NewFailedSLIResult creates a new SLIResult with result of fail.
func NewFailedSLIResult(metric string, message string) SLIResult {
	return NewFailedSLIResultWithQuery(metric, message, "")
}

// NewFailedSLIResultWithQuery creates a new SLIResult with a query and a result of fail.
func NewFailedSLIResultWithQuery(metric string, message string, query string) SLIResult {
	return SLIResult{
		Metric:          metric,
		Success:         false,
		Message:         message,
		Query:           query,
		IndicatorResult: IndicatorResultFailed,
	}
}

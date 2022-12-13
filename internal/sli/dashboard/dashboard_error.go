package dashboard

import "fmt"

// DashboardError represents a base error that happened while getting SLIs and SLOs from a dashboard.
type DashboardError struct {
	cause error
}

// NewDashboardError will create a new DashboardError.
func NewDashboardError(cause error) *DashboardError {
	return &DashboardError{
		cause: cause,
	}
}

func (e *DashboardError) Error() string {
	return fmt.Sprintf("could not get SLIs from dashboard: %v", e.cause)
}

func (e *DashboardError) Unwrap() error {
	return e.cause
}

// ProcessingError represents an error that happened while processing a dashboard.
type ProcessingError struct {
	cause error
}

// NewProcessingError will create a new ProcessingError.
func NewProcessingError(cause error) *ProcessingError {
	return &ProcessingError{
		cause: cause,
	}
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("could not process dashboard: %v", e.cause)
}

func (e *ProcessingError) Unwrap() error {
	return e.cause
}

// RetrievalError represents an error that happened while retrieving a dashboard.
type RetrievalError struct {
	cause error
}

// NewRetrievalError will create a new RetrievalError.
func NewRetrievalError(cause error) *RetrievalError {
	return &RetrievalError{
		cause: cause,
	}
}

func (e *RetrievalError) Error() string {
	return fmt.Sprintf("could not retrieve dashboard: %v", e.cause)
}

func (e *RetrievalError) Unwrap() error {
	return e.cause
}

// UploadSLOsError respresents an error that happened while uploading the SLO file.
type UploadSLOsError struct {
	cause error
}

// NewUploadSLOsError will create a new UploadSLOsError.
func NewUploadSLOsError(cause error) *UploadSLOsError {
	return &UploadSLOsError{
		cause: cause,
	}
}

func (e *UploadSLOsError) Error() string {
	return fmt.Sprintf("could not upload SLO file: %v", e.cause)
}

func (e *UploadSLOsError) Unwrap() error {
	return e.cause
}

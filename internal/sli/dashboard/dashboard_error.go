package dashboard

import "fmt"

// ProcessingError represents a base error that happened while processing a dashboard
type ProcessingError struct {
	cause error
}

func (pe *ProcessingError) Error() string {
	return pe.cause.Error()
}

func (pe *ProcessingError) Unwrap() error {
	return pe.cause
}

// QueryError represents an error that happened while querying a dashboard
type QueryError struct {
	err *ProcessingError
}

// NewQueryError will create a new QueryError
func NewQueryError(cause error) *QueryError {
	return &QueryError{
		err: &ProcessingError{
			cause: cause,
		},
	}
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("could not query Dynatrace dashboard for SLIs: %v", e.err.Error())
}

func (e *QueryError) Unwrap() error {
	return e.err
}

type UploadFileError struct {
	context string
	err     *ProcessingError
}

// NewUploadFileError will create a new UploadFileError
func NewUploadFileError(context string, cause error) *UploadFileError {
	return &UploadFileError{
		context: context,
		err: &ProcessingError{
			cause: cause,
		},
	}
}

func (e *UploadFileError) Error() string {
	return fmt.Sprintf("could not upload %s file: %v", e.context, e.err.Error())
}

func (e *UploadFileError) Unwrap() error {
	return e.err
}

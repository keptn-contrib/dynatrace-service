package dashboard

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQueryError(t *testing.T) {
	var errType *ProcessingError
	assert.ErrorAs(t, NewQueryError(errors.New("Couldnt process dashboard")), &errType)

}

func TestNewUploadFileError(t *testing.T) {
	var errType *ProcessingError
	assert.ErrorAs(t, NewUploadFileError("SLO", errors.New("Couldnt upload slo.yaml file")), &errType)
}

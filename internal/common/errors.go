package common

import "fmt"

const marshalType = "marshal"
const unmarshalType = "unmarshal"
const yamlType = "YAML"
const jsonType = "JSON"

type MarshalError struct {
	marshalType string
	context     string
	dataType    string
	cause       error
}

func NewUnmarshalYAMLError(context string, cause error) *MarshalError {
	return &MarshalError{
		marshalType: unmarshalType,
		context:     context,
		dataType:    yamlType,
		cause:       cause,
	}
}

func NewMarshalYAMLError(context string, cause error) *MarshalError {
	return &MarshalError{
		marshalType: marshalType,
		context:     context,
		dataType:    yamlType,
		cause:       cause,
	}
}

func NewUnmarshalJSONError(context string, cause error) *MarshalError {
	return &MarshalError{
		marshalType: unmarshalType,
		context:     context,
		dataType:    jsonType,
		cause:       cause,
	}
}

func NewMarshalJSONError(context string, cause error) *MarshalError {
	return &MarshalError{
		marshalType: marshalType,
		context:     context,
		dataType:    jsonType,
		cause:       cause,
	}
}

func (e *MarshalError) Error() string {
	return fmt.Sprintf("could not %s %s to %s (%v)", e.marshalType, e.context, e.dataType, e.cause)
}

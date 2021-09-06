// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package credentials_mock

import (
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"sync"
)

// CredentialManagerInterfaceMock is a mock implementation of credentials.CredentialManagerInterface.
//
// 	func TestSomethingThatUsesCredentialManagerInterface(t *testing.T) {
//
// 		// make and configure a mocked credentials.CredentialManagerInterface
// 		mockedCredentialManagerInterface := &CredentialManagerInterfaceMock{
// 			GetDynatraceCredentialsFunc: func(secretName string) (*credentials.DTCredentials, error) {
// 				panic("mock out the GetDynatraceCredentials method")
// 			},
// 			GetKeptnAPICredentialsFunc: func() (*credentials.KeptnAPICredentials, error) {
// 				panic("mock out the GetKeptnAPICredentials method")
// 			},
// 		}
//
// 		// use mockedCredentialManagerInterface in code that requires credentials.CredentialManagerInterface
// 		// and then make assertions.
//
// 	}
type CredentialManagerInterfaceMock struct {
	// GetDynatraceCredentialsFunc mocks the GetDynatraceCredentials method.
	GetDynatraceCredentialsFunc func(secretName string) (*credentials.DTCredentials, error)

	// GetKeptnAPICredentialsFunc mocks the GetKeptnAPICredentials method.
	GetKeptnAPICredentialsFunc func() (*credentials.KeptnAPICredentials, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetDynatraceCredentials holds details about calls to the GetDynatraceCredentials method.
		GetDynatraceCredentials []struct {
			// SecretName is the secretName argument value.
			SecretName string
		}
		// GetKeptnAPICredentials holds details about calls to the GetKeptnAPICredentials method.
		GetKeptnAPICredentials []struct {
		}
	}
	lockGetDynatraceCredentials sync.RWMutex
	lockGetKeptnAPICredentials  sync.RWMutex
}

// GetDynatraceCredentials calls GetDynatraceCredentialsFunc.
func (mock *CredentialManagerInterfaceMock) GetDynatraceCredentials(secretName string) (*credentials.DTCredentials, error) {
	if mock.GetDynatraceCredentialsFunc == nil {
		panic("CredentialManagerInterfaceMock.GetDynatraceCredentialsFunc: method is nil but CredentialManagerInterface.GetDynatraceCredentials was just called")
	}
	callInfo := struct {
		SecretName string
	}{
		SecretName: secretName,
	}
	mock.lockGetDynatraceCredentials.Lock()
	mock.calls.GetDynatraceCredentials = append(mock.calls.GetDynatraceCredentials, callInfo)
	mock.lockGetDynatraceCredentials.Unlock()
	return mock.GetDynatraceCredentialsFunc(secretName)
}

// GetDynatraceCredentialsCalls gets all the calls that were made to GetDynatraceCredentials.
// Check the length with:
//     len(mockedCredentialManagerInterface.GetDynatraceCredentialsCalls())
func (mock *CredentialManagerInterfaceMock) GetDynatraceCredentialsCalls() []struct {
	SecretName string
} {
	var calls []struct {
		SecretName string
	}
	mock.lockGetDynatraceCredentials.RLock()
	calls = mock.calls.GetDynatraceCredentials
	mock.lockGetDynatraceCredentials.RUnlock()
	return calls
}

// GetKeptnAPICredentials calls GetKeptnAPICredentialsFunc.
func (mock *CredentialManagerInterfaceMock) GetKeptnAPICredentials() (*credentials.KeptnAPICredentials, error) {
	if mock.GetKeptnAPICredentialsFunc == nil {
		panic("CredentialManagerInterfaceMock.GetKeptnAPICredentialsFunc: method is nil but CredentialManagerInterface.GetKeptnAPICredentials was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetKeptnAPICredentials.Lock()
	mock.calls.GetKeptnAPICredentials = append(mock.calls.GetKeptnAPICredentials, callInfo)
	mock.lockGetKeptnAPICredentials.Unlock()
	return mock.GetKeptnAPICredentialsFunc()
}

// GetKeptnAPICredentialsCalls gets all the calls that were made to GetKeptnAPICredentials.
// Check the length with:
//     len(mockedCredentialManagerInterface.GetKeptnAPICredentialsCalls())
func (mock *CredentialManagerInterfaceMock) GetKeptnAPICredentialsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetKeptnAPICredentials.RLock()
	calls = mock.calls.GetKeptnAPICredentials
	mock.lockGetKeptnAPICredentials.RUnlock()
	return calls
}

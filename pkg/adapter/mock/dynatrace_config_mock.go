// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package adapter_mock

import (
	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	"sync"
)

// DynatraceConfigGetterInterfaceMock is a mock implementation of adapter.DynatraceConfigGetterInterface.
//
// 	func TestSomethingThatUsesDynatraceConfigGetterInterface(t *testing.T) {
//
// 		// make and configure a mocked adapter.DynatraceConfigGetterInterface
// 		mockedDynatraceConfigGetterInterface := &DynatraceConfigGetterInterfaceMock{
// 			GetDynatraceConfigFunc: func(event adapter.EventContentAdapter, logger keptn.LoggerInterface) (*config.DynatraceConfigFile, error) {
// 				panic("mock out the GetDynatraceConfig method")
// 			},
// 		}
//
// 		// use mockedDynatraceConfigGetterInterface in code that requires adapter.DynatraceConfigGetterInterface
// 		// and then make assertions.
//
// 	}
type DynatraceConfigGetterInterfaceMock struct {
	// GetDynatraceConfigFunc mocks the GetDynatraceConfig method.
	GetDynatraceConfigFunc func(event adapter.EventContentAdapter, logger keptn.LoggerInterface) (*config.DynatraceConfigFile, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetDynatraceConfig holds details about calls to the GetDynatraceConfig method.
		GetDynatraceConfig []struct {
			// Event is the event argument value.
			Event adapter.EventContentAdapter
			// Logger is the logger argument value.
			Logger keptn.LoggerInterface
		}
	}
	lockGetDynatraceConfig sync.RWMutex
}

// GetDynatraceConfig calls GetDynatraceConfigFunc.
func (mock *DynatraceConfigGetterInterfaceMock) GetDynatraceConfig(event adapter.EventContentAdapter, logger keptn.LoggerInterface) (*config.DynatraceConfigFile, error) {
	if mock.GetDynatraceConfigFunc == nil {
		panic("DynatraceConfigGetterInterfaceMock.GetDynatraceConfigFunc: method is nil but DynatraceConfigGetterInterface.GetDynatraceConfig was just called")
	}
	callInfo := struct {
		Event  adapter.EventContentAdapter
		Logger keptn.LoggerInterface
	}{
		Event:  event,
		Logger: logger,
	}
	mock.lockGetDynatraceConfig.Lock()
	mock.calls.GetDynatraceConfig = append(mock.calls.GetDynatraceConfig, callInfo)
	mock.lockGetDynatraceConfig.Unlock()
	return mock.GetDynatraceConfigFunc(event, logger)
}

// GetDynatraceConfigCalls gets all the calls that were made to GetDynatraceConfig.
// Check the length with:
//     len(mockedDynatraceConfigGetterInterface.GetDynatraceConfigCalls())
func (mock *DynatraceConfigGetterInterfaceMock) GetDynatraceConfigCalls() []struct {
	Event  adapter.EventContentAdapter
	Logger keptn.LoggerInterface
} {
	var calls []struct {
		Event  adapter.EventContentAdapter
		Logger keptn.LoggerInterface
	}
	mock.lockGetDynatraceConfig.RLock()
	calls = mock.calls.GetDynatraceConfig
	mock.lockGetDynatraceConfig.RUnlock()
	return calls
}
package keptn

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/stretchr/testify/assert"
)

const testProblemURLLabelName = "Problem URL"

type testEventWithLabels struct {
	labels map[string]string
}

func TestTryGetProblemIDFromLabels(t *testing.T) {

	tests := []struct {
		name              string
		keptnEvent        adapter.EventContentAdapter
		expectedProblemID string
	}{
		{
			name: "Label with single valid problem ID - works",
			keptnEvent: testEventWithLabels{
				labels: map[string]string{
					testProblemURLLabelName: "https://dynatracetenant/#problems/problemdetails;gf=all;pid=2132340906857553706_1649682528607V2",
				},
			},
			expectedProblemID: "2132340906857553706_1649682528607V2",
		},
		{
			name: "Label with two problem IDs - works, returns second",
			keptnEvent: testEventWithLabels{
				labels: map[string]string{
					testProblemURLLabelName: "https://dynatracetenant/#problems/problemdetails;gf=all;pid=2132340906857553706_1649682528607V2;pid=8485558334848276629_1604413609638V2",
				},
			},
			expectedProblemID: "8485558334848276629_1604413609638V2",
		},
		{

			name:              "no labels - returns empty",
			keptnEvent:        testEventWithLabels{},
			expectedProblemID: "",
		},
		{

			name: "Label with with invalid URL - returns empty",
			keptnEvent: testEventWithLabels{
				labels: map[string]string{
					testProblemURLLabelName: "problems/problemdetails?gf=all;pid=2132340906857553706_1649682528607V2",
				},
			},
			expectedProblemID: "",
		},
		{

			name: "Label with with no fragment - returns empty",
			keptnEvent: testEventWithLabels{
				labels: map[string]string{
					testProblemURLLabelName: "https://dynatracetenant/problems/problemdetails?gf=all;pid=2132340906857553706_1649682528607V2",
				},
			},
			expectedProblemID: "",
		},
		{

			name: "some other label - returns empty",
			keptnEvent: testEventWithLabels{
				labels: map[string]string{
					"other_url": "https://dynatracetenant/#problems/problemdetails;gf=all;pid=2132340906857553706_1649682528607V2",
				},
			},
			expectedProblemID: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problemID := TryGetProblemIDFromLabels(tt.keptnEvent)
			assert.EqualValues(t, tt.expectedProblemID, problemID)
		})
	}
}

func (t testEventWithLabels) GetShKeptnContext() string {
	panic("GetShKeptnContext should not be called on mock")
}

func (t testEventWithLabels) GetEvent() string {
	panic("GetEvent should not be called on mock")
}

func (t testEventWithLabels) GetSource() string {
	panic("GetSource should not be called on mock")
}

func (t testEventWithLabels) GetProject() string {
	panic("GetProject should not be called on mock")
}

func (t testEventWithLabels) GetStage() string {
	panic("GetStage should not be called on mock")
}

func (t testEventWithLabels) GetService() string {
	panic("GetService should not be called on mock")
}

func (t testEventWithLabels) GetDeployment() string {
	panic("GetDeployment should not be called on mock")
}

func (t testEventWithLabels) GetTestStrategy() string {
	panic("GetTestStrategy should not be called on mock")
}

func (t testEventWithLabels) GetDeploymentStrategy() string {
	panic("GetDeploymentStrategy should not be called on mock")
}

func (t testEventWithLabels) GetLabels() map[string]string {
	return t.labels
}

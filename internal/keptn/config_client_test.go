package keptn

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testProject = "my-project"
const testStage = "my-stage"
const testService = "my-service"

const testSLIName = "response_time"
const testResponseTime90SLI = "metricSelector=builtin:service.response.time:splitBy():percentile(90)"
const testResponseTime95SLI = "metricSelector=builtin:service.response.time:splitBy():percentile(95)"

func TestConfigClient_GetSLIsNoneDefined(t *testing.T) {
	rc := NewConfigClient(&mockResourceClient{t: t})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Empty(t, slis)
	assert.NoError(t, err)
}

func TestConfigClient_GetSLIsServiceOverridesStage(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			stageResource:   getResponseTime90Resource(t),
			serviceResource: getResponseTime95Resource(t)})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.NoError(t, err)
	if !assert.Len(t, slis, 1) {
		return
	}

	if !assert.Contains(t, slis, testSLIName) {
		return
	}

	assert.EqualValues(t, testResponseTime95SLI, slis[testSLIName])
}

func TestConfigClient_GetSLIsStageOverridesProject(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			projectResource: getResponseTime90Resource(t),
			stageResource:   getResponseTime95Resource(t)})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.NoError(t, err)
	if !assert.Len(t, slis, 1) {
		return
	}

	if !assert.Contains(t, slis, testSLIName) {
		return
	}

	assert.EqualValues(t, testResponseTime95SLI, slis[testSLIName])
}

func TestConfigClient_GetSLIsServiceOverridesProject(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			projectResource: getResponseTime90Resource(t),
			serviceResource: getResponseTime95Resource(t)})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.NoError(t, err)
	if !assert.Len(t, slis, 1) {
		return
	}

	if !assert.Contains(t, slis, testSLIName) {
		return
	}

	assert.EqualValues(t, testResponseTime95SLI, slis[testSLIName])
}

func TestConfigClient_GetSLIsInvalidYAMLCausesError(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			projectResource: getInvalidYAMLResource(t),
			serviceResource: getResponseTime95Resource(t)})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Nil(t, slis)
	assert.Error(t, err)
}

func TestConfigClient_GetSLIsErrorCausesError(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			serviceResource: &mockResource{err: errors.New("Failed to connect")}})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Nil(t, slis)
	assert.Error(t, err)
}

func TestConfigClient_GetSLIsNoIndicatorsCausesError(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			serviceResource: getNoIndicatorsResource(t)})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Nil(t, slis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field")
}

const testDataFolder = "./testdata/config_client/get_slis"

func getNoIndicatorsResource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli_no_indicators.yaml")}
}

func getInvalidYAMLResource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli.invalid_yaml")}
}

func getResponseTime50Resource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli_response_time_50.yaml")}
}

func getResponseTime90Resource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli_response_time_90.yaml")}
}

func getResponseTime95Resource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli_response_time_95.yaml")}
}

func loadResource(t *testing.T, filename string) string {
	content, err := ioutil.ReadFile(filename)
	assert.NoError(t, err)
	return string(content)
}

type mockResource struct {
	resource string
	err      error
}

const testSLIResourceURI = "dynatrace/sli.yaml"

type mockResourceClient struct {
	t               *testing.T
	projectResource *mockResource
	stageResource   *mockResource
	serviceResource *mockResource
}

func (rc *mockResourceClient) GetResource(project string, stage string, service string, resourceURI string) (string, error) {
	rc.t.Fatalf("GetResource() should not be needed in this mock!")
	return "", nil
}

func (rc *mockResourceClient) GetProjectResource(project string, resourceURI string) (string, error) {
	assert.EqualValues(rc.t, testProject, project)
	assert.EqualValues(rc.t, testSLIResourceURI, resourceURI)

	if rc.projectResource == nil {
		return "", &ResourceNotFoundError{project: project, uri: resourceURI}
	}

	return rc.projectResource.resource, rc.projectResource.err
}

func (rc *mockResourceClient) GetStageResource(project string, stage string, resourceURI string) (string, error) {
	assert.EqualValues(rc.t, testProject, project)
	assert.EqualValues(rc.t, testStage, stage)
	assert.EqualValues(rc.t, testSLIResourceURI, resourceURI)

	if rc.stageResource == nil {
		return "", &ResourceNotFoundError{project: project, stage: stage, uri: resourceURI}
	}

	return rc.stageResource.resource, rc.stageResource.err
}

func (rc *mockResourceClient) GetServiceResource(project string, stage string, service string, resourceURI string) (string, error) {
	assert.EqualValues(rc.t, testProject, project)
	assert.EqualValues(rc.t, testStage, stage)
	assert.EqualValues(rc.t, testService, service)
	assert.EqualValues(rc.t, testSLIResourceURI, resourceURI)

	if rc.serviceResource == nil {
		return "", &ResourceNotFoundError{project: project, stage: stage, service: service, uri: resourceURI}
	}

	return rc.serviceResource.resource, rc.serviceResource.err
}

func (rc *mockResourceClient) UploadResource(contentToUpload []byte, remoteResourceURI string, project string, stage string, service string) error {
	rc.t.Fatalf("UploadResource() should not be needed in this mock!")
	return nil
}

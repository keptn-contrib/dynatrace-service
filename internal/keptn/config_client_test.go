package keptn

import (
	"io/ioutil"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/stretchr/testify/assert"
)

const testProject = "my-project"
const testStage = "my-stage"
const testService = "my-service"

// TestConfigClient_GetSLIsNoneDefined tests that getting SLIs when none have been defined returns an empty map but no error.
func TestConfigClient_GetSLIsNoneDefined(t *testing.T) {
	rc := NewConfigClient(&mockResourceClient{t: t})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Empty(t, slis)
	assert.NoError(t, err)
}

// TestConfigClient_GetSLIsWithOverrides tests that service-level SLIs override stage or project-level SLIs and stage-level SLIs override project-level ones.
// In addition any SLIs defined only at a project or stage level should also be returned.
func TestConfigClient_GetSLIsWithOverrides(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			projectResource: getProjectResource(t),
			stageResource:   getStageResource(t),
			serviceResource: getServiceResource(t)})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.NoError(t, err)

	expectedSLIs := map[string]string{
		"sli_a": "metricSelector=builtin:service.response.time:splitBy():percentile(95)&entitySelector=tag(keptn_project:my-project),tag(keptn_stage:my-stage),tag(kept_service:my-service)",
		"sli_b": "metricSelector=builtin:service.response.time:splitBy():percentile(90)&entitySelector=tag(keptn_project:my-project),tag(keptn_stage:my-stage),tag(kept_service:my-service)",
		"sli_c": "metricSelector=builtin:service.response.time:splitBy():percentile(80)&entitySelector=tag(keptn_project:my-project),tag(keptn_stage:my-stage),tag(kept_service:my-service)",
		"sli_d": "metricSelector=builtin:service.response.time:splitBy():percentile(75)&entitySelector=tag(keptn_project:my-project),tag(keptn_stage:my-stage),tag(kept_service:my-service)",
		"sli_e": "metricSelector=builtin:service.response.time:splitBy():percentile(70)&entitySelector=tag(keptn_project:my-project),tag(keptn_stage:my-stage)",
		"sli_f": "metricSelector=builtin:service.response.time:splitBy():percentile(55)&entitySelector=tag(keptn_project:my-project)",
	}

	if !assert.EqualValues(t, len(expectedSLIs), len(slis)) {
		return
	}

	for expectedKey, expectedValue := range expectedSLIs {
		value, ok := slis[expectedKey]
		assert.True(t, ok)
		assert.EqualValues(t, expectedValue, value)
	}
}

// TestConfigClient_GetSLIsInvalidYAMLCausesError tests that an invalid SLI YAML resource produces an error.
func TestConfigClient_GetSLIsInvalidYAMLCausesError(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			projectResource: getInvalidYAMLResource(t)})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Nil(t, slis)
	var marshalErr *common.MarshalError
	assert.ErrorAs(t, err, &marshalErr)
}

// TestConfigClient_GetSLIsRetrievalErrorCausesError tests that resource retrieval errors produce an error.
func TestConfigClient_GetSLIsRetrievalErrorCausesError(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			serviceResource: &mockResource{err: &ResourceRetrievalFailedError{ResourceError{uri: testSLIResourceURI, project: testProject, stage: testStage, service: testService}, "Connection error"}}})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Nil(t, slis)
	assert.Error(t, err)
	var rrfErrorType *ResourceRetrievalFailedError
	assert.ErrorAs(t, err, &rrfErrorType)
}

// TestConfigClient_GetSLIsEmptySLIFileCausesError tests that an empty SLI file produces an error.
func TestConfigClient_GetSLIsEmptySLIFileCausesError(t *testing.T) {
	rc := NewConfigClient(
		&mockResourceClient{
			t:               t,
			serviceResource: &mockResource{err: &ResourceEmptyError{uri: testSLIResourceURI, project: testProject, stage: testStage, service: testService}}})
	slis, err := rc.GetSLIs(testProject, testStage, testService)
	assert.Nil(t, slis)
	assert.Error(t, err)
	var rrfErrorType *ResourceEmptyError
	assert.ErrorAs(t, err, &rrfErrorType)
}

// TestConfigClient_GetSLIsNoIndicatorsCausesError tests that an SLI file containing no indicators produces an error.
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

func getServiceResource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli_service.yaml")}
}

func getStageResource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli_stage.yaml")}
}

func getProjectResource(t *testing.T) *mockResource {
	return &mockResource{resource: loadResource(t, testDataFolder+"/sli_project.yaml")}
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

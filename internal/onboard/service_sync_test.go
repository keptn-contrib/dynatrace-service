package onboard

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	keptnlib "github.com/keptn/go-utils/pkg/lib"
)

const testDynatraceAPIToken = "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"

func Test_doesServiceExist(t *testing.T) {
	type args struct {
		services    []string
		serviceName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "service does exist",
			args: args{
				services:    []string{"service-1", "service-2"},
				serviceName: "service-1",
			},
			want: true,
		},
		{
			name: "service does not exist",
			args: args{
				services:    []string{"service-1", "service-2"},
				serviceName: "service-3",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doesServiceExist(tt.args.services, tt.args.serviceName); got != tt.want {
				t.Errorf("doesServiceExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

type uploadedSLIs struct {
	project string
	stage   string
	service string
	slis    *dynatrace.SLI
}

type uploadedSLOs struct {
	project string
	stage   string
	service string
	slos    *keptnlib.ServiceLevelObjectives
}

type mockSLIAndSLOResourceWriter struct {
	uploadedSLIs []uploadedSLIs
	uploadedSLOs []uploadedSLOs
}

func (w *mockSLIAndSLOResourceWriter) UploadSLIs(project string, stage string, service string, slis *dynatrace.SLI) error {
	w.uploadedSLIs = append(w.uploadedSLIs, uploadedSLIs{project: project, stage: stage, service: service, slis: slis})
	return nil
}

func (w *mockSLIAndSLOResourceWriter) UploadSLOs(project string, stage string, service string, slos *keptnlib.ServiceLevelObjectives) error {
	w.uploadedSLOs = append(w.uploadedSLOs, uploadedSLOs{project: project, stage: stage, service: service, slos: slos})
	return nil
}

const mockSynchronizedProject = "dynatrace"
const mockSynchronizedStage = "quality-gate"

type mockServicesClient struct {
	servicesInDynatraceProject []string
	servicesCreated            []string
}

func (c *mockServicesClient) GetServiceNames(project string, stage string) ([]string, error) {
	if project != mockSynchronizedProject {
		return nil, fmt.Errorf("project %s does not exist", project)
	}

	if stage != mockSynchronizedStage {
		return nil, fmt.Errorf("stage %s does not exist", stage)
	}

	return c.servicesInDynatraceProject, nil
}

func (c *mockServicesClient) CreateServiceInProject(project string, service string) error {
	if project != mockSynchronizedProject {
		return fmt.Errorf("project %s does not exist", project)
	}

	for _, existingService := range c.servicesInDynatraceProject {
		if service == existingService {
			return fmt.Errorf("service %s already exists in project %s", service, project)
		}
	}

	c.servicesInDynatraceProject = append(c.servicesInDynatraceProject, service)
	c.servicesCreated = append(c.servicesCreated, service)
	return nil
}

func newMockServicesClient() *mockServicesClient {
	return &mockServicesClient{
		servicesInDynatraceProject: []string{"my-already-synced-service"},
	}
}

type mockEntitiesClientFactory struct {
	handler    *test.FileBasedURLHandler
	httpClient *http.Client
	url        string
}

func newMockEntitiesClientFactory(t *testing.T) (*mockEntitiesClientFactory, func()) {
	handler := test.NewFileBasedURLHandler(t)
	httpClient, url, teardown := test.CreateHTTPSClient(handler)

	return &mockEntitiesClientFactory{
			handler:    handler,
			httpClient: httpClient,
			url:        url},
		teardown
}

func (f *mockEntitiesClientFactory) CreateEntitiesClient() (*dynatrace.EntitiesClient, error) {
	dynatraceCredentials, err := credentials.NewDynatraceCredentials(f.url, testDynatraceAPIToken)
	if err != nil {
		return nil, err
	}

	dynatraceClient := dynatrace.NewClientWithHTTP(dynatraceCredentials, f.httpClient)
	return dynatrace.NewEntitiesClient(dynatraceClient), nil
}

func Test_ServiceSynchronizer_synchronizeServices(t *testing.T) {

	mockServicesClient := newMockServicesClient()
	mockSLIAndSLOResourceWriter := &mockSLIAndSLOResourceWriter{}

	mockEntitiesClientFactory, teardown := newMockEntitiesClientFactory(t)
	defer teardown()

	const testDataFolder = "./testdata/test_synchronize_services/"
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?entitySelector=type(\"SERVICE\")%20AND%20tag(\"keptn_managed\",\"[Environment]keptn_managed\")%20AND%20tag(\"keptn_service\",\"[Environment]keptn_service\")&fields=+tags&pageSize=50", testDataFolder+"entities_response1.json")
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?nextPageKey=next-page-key", testDataFolder+"entities_response2.json")

	s := &ServiceSynchronizer{
		servicesClient:        mockServicesClient,
		resourcesClient:       mockSLIAndSLOResourceWriter,
		entitiesClientFactory: mockEntitiesClientFactory,
	}
	s.synchronizeServices()

	onboardedService1 := "my-service"
	onboardedService2 := "my-service-2"

	// validate if all service creation requests have been sent
	if assert.EqualValues(t, 2, len(mockServicesClient.servicesCreated)) {
		assert.EqualValues(t, onboardedService1, mockServicesClient.servicesCreated[0])
		assert.EqualValues(t, onboardedService2, mockServicesClient.servicesCreated[1])
	}

	// validate if all SLO uploads have been received
	if assert.EqualValues(t, 2, len(mockSLIAndSLOResourceWriter.uploadedSLOs)) {
		assert.EqualValues(t, onboardedService1, mockSLIAndSLOResourceWriter.uploadedSLOs[0].service)
		assert.EqualValues(t, onboardedService2, mockSLIAndSLOResourceWriter.uploadedSLOs[1].service)
	}

	// validate if all SLI uploads have been received
	if assert.EqualValues(t, 2, len(mockSLIAndSLOResourceWriter.uploadedSLIs)) {
		assert.EqualValues(t, onboardedService1, mockSLIAndSLOResourceWriter.uploadedSLIs[0].service)
		assert.EqualValues(t, onboardedService2, mockSLIAndSLOResourceWriter.uploadedSLIs[1].service)
	}

	// perform a second synchronization run
	s.synchronizeServices()

	// nothing extra should have been created or uploaded
	assert.EqualValues(t, 2, len(mockServicesClient.servicesCreated))
	assert.EqualValues(t, 2, len(mockSLIAndSLOResourceWriter.uploadedSLOs))
	assert.EqualValues(t, 2, len(mockSLIAndSLOResourceWriter.uploadedSLIs))
}

func Test_getServiceFromEntity(t *testing.T) {
	type args struct {
		entity dynatrace.Entity
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "error due to missing tag",
			args: args{
				entity: dynatrace.Entity{
					EntityID:    "entity-id",
					DisplayName: ":10999",
					Tags:        nil,
				},
			},
			wantErr: true,
		},
		{
			name: "use keptn_service tag",
			args: args{
				entity: dynatrace.Entity{
					EntityID:    "entity-id",
					DisplayName: ":10999",
					Tags: []dynatrace.Tag{
						{
							Context:              "CONTEXTLESS",
							Key:                  "keptn_service",
							StringRepresentation: "keptn_service:my-service",
							Value:                "my-service",
						},
					},
				},
			},
			want:    "my-service",
			wantErr: false,
		},
		{
			name: "keptn_service tag with no value",
			args: args{
				entity: dynatrace.Entity{
					EntityID:    "entity-id",
					DisplayName: ":10999",
					Tags: []dynatrace.Tag{
						{
							Context:              "CONTEXTLESS",
							Key:                  "keptn_service",
							StringRepresentation: "keptn_service",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "two keptn_service tags",
			args: args{
				entity: dynatrace.Entity{
					EntityID:    "entity-id",
					DisplayName: ":10999",
					Tags: []dynatrace.Tag{
						{
							Context:              "CONTEXTLESS",
							Key:                  "keptn_service",
							StringRepresentation: "keptn_service:value1",
							Value:                "value1",
						},
						{
							Context:              "CONTEXTLESS",
							Key:                  "keptn_service",
							StringRepresentation: "keptn_service:value2",
							Value:                "value2",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getServiceFromEntity(tt.args.entity)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

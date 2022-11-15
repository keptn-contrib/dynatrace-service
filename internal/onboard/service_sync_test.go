package onboard

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
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
	slis    *keptn.SLI
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

func (w *mockSLIAndSLOResourceWriter) UploadSLIs(_ context.Context, project string, stage string, service string, slis *keptn.SLI) error {
	w.uploadedSLIs = append(w.uploadedSLIs, uploadedSLIs{project: project, stage: stage, service: service, slis: slis})
	return nil
}

func (w *mockSLIAndSLOResourceWriter) UploadSLOs(_ context.Context, project string, stage string, service string, slos *keptnlib.ServiceLevelObjectives) error {
	w.uploadedSLOs = append(w.uploadedSLOs, uploadedSLOs{project: project, stage: stage, service: service, slos: slos})
	return nil
}

const mockSynchronizedProject = "dynatrace"
const mockSynchronizedStage = "quality-gate"

type mockServicesClient struct {
	existingServices []string
	createdServices  []string
}

func (c *mockServicesClient) GetServiceNames(_ context.Context, project string, stage string) ([]string, error) {
	if project != mockSynchronizedProject {
		return nil, fmt.Errorf("project %s does not exist", project)
	}

	if stage != mockSynchronizedStage {
		return nil, fmt.Errorf("stage %s does not exist", stage)
	}

	return c.existingServices, nil
}

func (c *mockServicesClient) CreateServiceInProject(_ context.Context, project string, service string) error {
	if project != mockSynchronizedProject {
		return fmt.Errorf("project %s does not exist", project)
	}

	for _, existingService := range c.existingServices {
		if service == existingService {
			return fmt.Errorf("service %s already exists in project %s", service, project)
		}
	}

	c.existingServices = append(c.existingServices, service)
	c.createdServices = append(c.createdServices, service)
	return nil
}

func newMockServicesClient(existingServices []string) *mockServicesClient {
	return &mockServicesClient{
		existingServices: existingServices,
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

func (f *mockEntitiesClientFactory) CreateEntitiesClient(_ context.Context) (*dynatrace.EntitiesClient, error) {
	dynatraceCredentials, err := credentials.NewDynatraceCredentials(f.url, testDynatraceAPIToken)
	if err != nil {
		return nil, err
	}

	dynatraceClient := dynatrace.NewClientWithHTTP(dynatraceCredentials, f.httpClient)
	return dynatrace.NewEntitiesClient(dynatraceClient), nil
}

// Test_ServiceSynchronizer_synchronizeServices_addNew tests that new services are added.
func Test_ServiceSynchronizer_synchronizeServices_addNew(t *testing.T) {
	mockServicesClient := newMockServicesClient([]string{"my-already-synced-service"})
	mockSLIAndSLOResourceWriter := &mockSLIAndSLOResourceWriter{}

	mockEntitiesClientFactory, teardown := newMockEntitiesClientFactory(t)
	defer teardown()

	const testDataFolder = "./testdata/test_synchronize_services_add_new/"
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?entitySelector=type%28%22SERVICE%22%29+AND+tag%28%22keptn_managed%22%2C%22%5BEnvironment%5Dkeptn_managed%22%29+AND+tag%28%22keptn_service%22%2C%22%5BEnvironment%5Dkeptn_service%22%29&fields=%2Btags&pageSize=50", filepath.Join(testDataFolder, "entities_response1.json"))
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?nextPageKey=next-page-key", filepath.Join(testDataFolder, "entities_response2.json"))

	s := &ServiceSynchronizer{
		servicesClient:        mockServicesClient,
		resourcesClient:       mockSLIAndSLOResourceWriter,
		entitiesClientFactory: mockEntitiesClientFactory,
	}
	s.synchronizeServices(context.Background())

	onboardedService1 := "my-service"
	onboardedService2 := "my-service-2"

	// validate if all service creation requests have been sent
	if assert.EqualValues(t, 2, len(mockServicesClient.createdServices)) {
		assert.EqualValues(t, onboardedService1, mockServicesClient.createdServices[0])
		assert.EqualValues(t, onboardedService2, mockServicesClient.createdServices[1])
	}

	// validate if all SLO uploads have been received
	if assert.EqualValues(t, 2, len(mockSLIAndSLOResourceWriter.uploadedSLOs)) {
		expectedUploadedSLOs := []uploadedSLOs{
			{
				project: mockSynchronizedProject,
				stage:   mockSynchronizedStage,
				service: onboardedService1,
				slos: &keptnlib.ServiceLevelObjectives{
					SpecVersion: "1.0",
					Filter:      nil,
					Comparison: &keptnlib.SLOComparison{
						AggregateFunction:         "avg",
						CompareWith:               "single_result",
						IncludeResultWithScore:    "pass",
						NumberOfComparisonResults: 1,
					},
					Objectives: []*keptnlib.SLO{
						{
							SLI:     "response_time_p95",
							KeySLI:  false,
							Pass:    []*keptnlib.SLOCriteria{{Criteria: []string{"<600"}}},
							Warning: []*keptnlib.SLOCriteria{{Criteria: []string{"<=800"}}},
							Weight:  1,
						},
						{
							SLI:    "error_rate",
							KeySLI: false,
							Pass:   []*keptnlib.SLOCriteria{{Criteria: []string{"<5"}}},
							Weight: 1,
						},
						{
							SLI: "throughput",
						},
					},
					TotalScore: &keptnlib.SLOScore{
						Pass:    "90%",
						Warning: "75%",
					},
				},
			},
			{
				project: mockSynchronizedProject,
				stage:   mockSynchronizedStage,
				service: onboardedService2,
				slos: &keptnlib.ServiceLevelObjectives{
					SpecVersion: "1.0",
					Filter:      nil,
					Comparison: &keptnlib.SLOComparison{
						AggregateFunction:         "avg",
						CompareWith:               "single_result",
						IncludeResultWithScore:    "pass",
						NumberOfComparisonResults: 1,
					},
					Objectives: []*keptnlib.SLO{
						{
							SLI:     "response_time_p95",
							KeySLI:  false,
							Pass:    []*keptnlib.SLOCriteria{{Criteria: []string{"<600"}}},
							Warning: []*keptnlib.SLOCriteria{{Criteria: []string{"<=800"}}},
							Weight:  1,
						},
						{
							SLI:    "error_rate",
							KeySLI: false,
							Pass:   []*keptnlib.SLOCriteria{{Criteria: []string{"<5"}}},
							Weight: 1,
						},
						{
							SLI: "throughput",
						},
					},
					TotalScore: &keptnlib.SLOScore{
						Pass:    "90%",
						Warning: "75%",
					},
				},
			},
		}
		assert.EqualValues(t, expectedUploadedSLOs, mockSLIAndSLOResourceWriter.uploadedSLOs)
	}

	// validate if all SLI uploads have been received
	if assert.EqualValues(t, 2, len(mockSLIAndSLOResourceWriter.uploadedSLIs)) {
		expectedUploadedSLIs := []uploadedSLIs{
			{
				project: mockSynchronizedProject,
				stage:   mockSynchronizedStage,
				service: onboardedService1,
				slis: &keptn.SLI{
					SpecVersion: "1.0",
					Indicators: map[string]string{
						"throughput":        fmt.Sprintf("metricSelector=builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService1),
						"error_rate":        fmt.Sprintf("metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService1),
						"response_time_p50": fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService1),
						"response_time_p90": fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(90)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService1),
						"response_time_p95": fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService1),
					},
				},
			},
			{
				project: mockSynchronizedProject,
				stage:   mockSynchronizedStage,
				service: onboardedService2,
				slis: &keptn.SLI{
					SpecVersion: "1.0",
					Indicators: map[string]string{
						"throughput":        fmt.Sprintf("metricSelector=builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService2),
						"error_rate":        fmt.Sprintf("metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService2),
						"response_time_p50": fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService2),
						"response_time_p90": fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(90)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService2),
						"response_time_p95": fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", onboardedService2),
					},
				},
			},
		}
		assert.EqualValues(t, expectedUploadedSLIs, mockSLIAndSLOResourceWriter.uploadedSLIs)
	}
}

// Test_ServiceSynchronizer_synchronizeServices_skipExisting tests that services that have already been added are not added twice.
func Test_ServiceSynchronizer_synchronizeServices_skipExisting(t *testing.T) {
	mockServicesClient := newMockServicesClient([]string{"my-already-synced-service", "my-service", "my-service-2"})
	mockSLIAndSLOResourceWriter := &mockSLIAndSLOResourceWriter{}

	mockEntitiesClientFactory, teardown := newMockEntitiesClientFactory(t)
	defer teardown()

	const testDataFolder = "./testdata/test_synchronize_services_add_new/"
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?entitySelector=type%28%22SERVICE%22%29+AND+tag%28%22keptn_managed%22%2C%22%5BEnvironment%5Dkeptn_managed%22%29+AND+tag%28%22keptn_service%22%2C%22%5BEnvironment%5Dkeptn_service%22%29&fields=%2Btags&pageSize=50", filepath.Join(testDataFolder, "entities_response1.json"))
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?nextPageKey=next-page-key", filepath.Join(testDataFolder, "entities_response2.json"))

	s := &ServiceSynchronizer{
		servicesClient:        mockServicesClient,
		resourcesClient:       mockSLIAndSLOResourceWriter,
		entitiesClientFactory: mockEntitiesClientFactory,
	}
	s.synchronizeServices(context.Background())

	// no services should have been created
	assert.EqualValues(t, 0, len(mockServicesClient.createdServices))
	assert.EqualValues(t, 0, len(mockSLIAndSLOResourceWriter.uploadedSLOs))
	assert.EqualValues(t, 0, len(mockSLIAndSLOResourceWriter.uploadedSLIs))
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

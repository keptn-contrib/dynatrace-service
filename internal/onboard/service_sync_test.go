package onboard

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	adapter_mock "github.com/keptn-contrib/dynatrace-service/internal/config/mock"
	credentials_mock "github.com/keptn-contrib/dynatrace-service/internal/credentials/mock"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-test/deep"
	"github.com/google/uuid"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const testDTEntityQueryResponse = `{
    "totalCount": 1,
    "pageSize": 50,
    "entities": [
        {
            "entityId": "SERVICE-B0254D5C9720662A",
            "displayName": "bridge",
            "tags": [
                {
                    "context": "CONTEXTLESS",
                    "key": "keptn_managed",
                    "stringRepresentation": "keptn_managed"
                },
                {
                    "context": "CONTEXTLESS",
                    "key": "keptn_service",
                    "value": "bridge",
                    "stringRepresentation": "keptn_service:bridge"
                }
            ]
        }
    ]
}`

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

func getTestServicesAPI() *httptest.Server {
	servicesMockAPI := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		svc := &models.Service{
			ServiceName: "my-service",
		}
		marshal, _ := json.Marshal(svc)

		writer.WriteHeader(http.StatusOK)
		writer.Write(marshal)
	}))
	return servicesMockAPI
}

func getTestProjectsAPI() *httptest.Server {
	projectsMockAPI := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		svc := &models.Project{
			ProjectName: "dynatrace",
		}
		marshal, _ := json.Marshal(svc)

		writer.WriteHeader(http.StatusOK)
		writer.Write(marshal)
	}))
	return projectsMockAPI
}

func getTestMockEventBroker() (chan string, *httptest.Server) {
	receivedEvent := make(chan string)
	mockEventBroker := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		bytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return
		}
		event := &models.KeptnContextExtendedCE{}
		err = json.Unmarshal(bytes, event)
		if err != nil {
			return
		}
		if *event.Type == keptnv2.GetFinishedEventType(keptnv2.ServiceCreateTaskName) {
			bytes, err = json.Marshal(event.Data)
			if err != nil {
				return
			}
			serviceCreateData := &keptnv2.ServiceCreateFinishedEventData{}
			err = json.Unmarshal(bytes, serviceCreateData)
			if err != nil {
				return
			}
			if serviceCreateData.Project == defaultDTProjectName {
				go func() {
					receivedEvent <- serviceCreateData.Service
				}()
				return
			}
		}

	}))
	return receivedEvent, mockEventBroker
}

type createServiceParams struct {
	ServiceName string `json:"serviceName"`
}

func getTestConfigService() (chan string, chan string, chan string, *httptest.Server) {
	receivedSLO := make(chan string)
	receivedSLI := make(chan string)
	receivedServiceCreate := make(chan string)
	mockCS := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		bytes, err := ioutil.ReadAll(request.Body)

		if strings.HasSuffix(request.URL.String(), "dynatrace/service") {
			createSvcParam := &createServiceParams{}
			err = json.Unmarshal(bytes, createSvcParam)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			if createSvcParam.ServiceName != "" {
				go func() {
					receivedServiceCreate <- createSvcParam.ServiceName
				}()
			}
			writer.WriteHeader(http.StatusOK)
			return
		}

		var serviceName string
		split := strings.Split(request.URL.String(), "/")
		serviceNameIndex := 0
		for i, s := range split {
			if s == "service" {
				serviceNameIndex = i + 1
				break
			}
		}
		if serviceNameIndex > 0 {
			serviceName = split[serviceNameIndex]
		} else {
			writer.WriteHeader(http.StatusOK)
			return
		}

		rec := &models.Resources{}

		err = json.Unmarshal(bytes, rec)
		if err != nil {
			writer.WriteHeader(http.StatusOK)
			return
		}
		if rec.Resources[0].ResourceURI != nil && *rec.Resources[0].ResourceURI == "slo.yaml" {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("{}"))
			go func() {
				receivedSLO <- serviceName
			}()
		} else if rec.Resources[0].ResourceURI != nil && *rec.Resources[0].ResourceURI == "dynatrace/sli.yaml" {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("{}"))
			go func() {
				receivedSLI <- serviceName
			}()
		}

	}))
	return receivedServiceCreate, receivedSLO, receivedSLI, mockCS
}

func getTestKeptnHandler(mockCS *httptest.Server, mockEventBroker *httptest.Server) *keptnv2.Keptn {
	source, _ := url.Parse("dynatrace-service")
	keptnContext := uuid.New().String()
	createServiceData := keptnv2.ServiceCreateFinishedEventData{
		EventData: keptnv2.EventData{
			Project: defaultDTProjectName,
			Service: "my-service",
		},
	}
	ce := cloudevents.NewEvent()
	ce.SetType(keptnv2.GetFinishedEventType(keptnv2.ServiceCreateTaskName))
	ce.SetSource(source.String())
	ce.SetExtension("shkeptncontext", keptnContext)
	ce.SetDataContentType(cloudevents.ApplicationJSON)
	ce.SetData(cloudevents.ApplicationJSON, createServiceData)

	k, _ := keptnv2.NewKeptn(&ce, keptncommon.KeptnOpts{
		ConfigurationServiceURL: mockCS.URL,
		EventBrokerURL:          mockEventBroker.URL,
	})
	return k
}

func Test_ServiceSynchronizer_synchronizeServices(t *testing.T) {

	firstDTResponse := dynatrace.EntitiesResponse{
		TotalCount:  3,
		PageSize:    2,
		NextPageKey: "next-page-key",
		Entities: []dynatrace.Entity{
			{
				EntityID:    "1",
				DisplayName: "name",
				Tags: []dynatrace.Tag{
					{
						Context:              "CONTEXTLESS",
						Key:                  "keptn_managed",
						StringRepresentation: "keptn_managed",
						Value:                "",
					},
					{
						Context:              "CONTEXTLESS",
						Key:                  "keptn_service",
						StringRepresentation: "keptn_service:my-service",
						Value:                "my-service",
					},
				},
			},
			{
				EntityID:    "1-2",
				DisplayName: "name",
				Tags: []dynatrace.Tag{
					{
						Context:              "CONTEXTLESS",
						Key:                  "keptn_managed",
						StringRepresentation: "keptn_managed",
						Value:                "",
					},
					{
						Context:              "CONTEXTLESS",
						Key:                  "keptn_service",
						StringRepresentation: "keptn_service:my-already-synced-service",
						Value:                "my-already-synced-service",
					},
				},
			},
		},
	}

	secondDTResponse := dynatrace.EntitiesResponse{
		TotalCount:  2,
		PageSize:    1,
		NextPageKey: "",
		Entities: []dynatrace.Entity{
			{
				EntityID:    "2",
				DisplayName: "name",
				Tags: []dynatrace.Tag{
					{
						Context:              "CONTEXTLESS",
						Key:                  "keptn_managed",
						StringRepresentation: "keptn_managed",
						Value:                "",
					},
					{
						Context:              "CONTEXTLESS",
						Key:                  "keptn_service",
						StringRepresentation: "keptn_service:my-service-2",
						Value:                "my-service-2",
					},
				},
			},
		},
	}
	isFirstRequest := true
	dtMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isFirstRequest {
			if request.FormValue("nextPageKey") != "" {
				t.Errorf("call to Dynatrace API received unexpected nextPageKey parameter")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			isFirstRequest = false
			marshal, _ := json.Marshal(firstDTResponse)
			writer.WriteHeader(http.StatusOK)
			writer.Write(marshal)
			return
		}
		if request.FormValue("nextPageKey") != firstDTResponse.NextPageKey {
			t.Errorf("call to Dynatrace API received unexpected nextPageKey parameter %s. Expected %s", request.FormValue("nextPageKey"), firstDTResponse.NextPageKey)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.FormValue("entitySelector") != "" {
			t.Errorf("entitySelector parameter must not be used in combination with nextPageKey")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.FormValue("fields") != "" {
			t.Errorf("fields parameter must not be used in combination with nextPageKey")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		marshal, _ := json.Marshal(secondDTResponse)
		writer.WriteHeader(http.StatusOK)
		writer.Write(marshal)
	}))
	defer dtMockServer.Close()

	projectsMockAPI := getTestProjectsAPI()
	defer projectsMockAPI.Close()

	firstRequest := true
	servicesMockAPI := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// for the first request, return a list of the services already available in Keptn
		if firstRequest {
			svcList := &models.Services{
				Services: []*models.Service{
					{
						ServiceName: "my-already-synced-service",
					},
				},
			}
			marshal, _ := json.Marshal(svcList)

			writer.WriteHeader(http.StatusOK)
			writer.Write(marshal)
			return
		}
		svc := &models.Service{
			ServiceName: "my-service",
		}
		marshal, _ := json.Marshal(svc)

		writer.WriteHeader(http.StatusOK)
		writer.Write(marshal)
	}))
	defer servicesMockAPI.Close()

	_, mockEventBroker := getTestMockEventBroker()
	defer mockEventBroker.Close()

	receivedServiceCreate, receivedSLO, receivedSLI, mockCS := getTestConfigService()
	defer mockCS.Close()

	os.Setenv(common.ShipyardControllerURLEnvironmentVariableName, mockCS.URL)

	mockCredentials, err := credentials.NewDynatraceCredentials(dtMockServer.URL, testDynatraceAPIToken)
	assert.NoError(t, err)

	s := &ServiceSynchronizer{
		projectClient:   keptn.NewProjectClient(keptnapi.NewProjectHandler(projectsMockAPI.URL)),
		servicesClient:  keptn.NewServiceClient(keptnapi.NewServiceHandler(servicesMockAPI.URL), mockCS.Client()),
		resourcesClient: keptn.NewResourceClient(keptn.NewConfigResourceClient(keptnapi.NewResourceHandler(mockCS.URL))),
		EntitiesClientFunc: func(creds *credentials.DynatraceCredentials) *dynatrace.EntitiesClient {
			return dynatrace.NewEntitiesClient(
				dynatrace.NewClient(mockCredentials))
		},
		servicesInKeptn: []string{},
		credentialsProvider: &credentials_mock.DynatraceCredentialsProviderMock{
			GetDynatraceCredentialsFunc: func(secretName string) (*credentials.DynatraceCredentials, error) {
				return mockCredentials, nil
			},
		},
		configProvider: &adapter_mock.DynatraceConfigProviderMock{
			GetDynatraceConfigFunc: func(event adapter.EventContentAdapter) (*config.DynatraceConfig, error) {
				return config.NewDynatraceConfigWithDefaults(), nil
			}},
	}
	s.synchronizeServices()

	// validate if all service creation requests have been sent
	if done := checkReceivedEntities(t, receivedServiceCreate, []string{"my-service", "my-service-2"}); done {
		t.Error("did not receive expected service creation requests")
	}

	// validate if all SLO uploads have been received
	if done := checkReceivedEntities(t, receivedSLO, []string{"my-service", "my-service-2"}); done {
		t.Error("did not receive expected service creation requests")
	}

	// validate if all SLI uploads have been received
	if done := checkReceivedEntities(t, receivedSLI, []string{"my-service", "my-service-2"}); done {
		t.Error("did not receive expected service creation requests")
	}
}

func checkReceivedEntities(t *testing.T, channel chan string, expected []string) bool {
	received := []string{}
	for {
		select {
		case rec := <-channel:
			received = append(received, rec)
			if len(received) == 2 {
				if diff := deep.Equal(received, expected); len(diff) > 0 {
					t.Error("expected did not match received:")
					for _, d := range diff {
						t.Log(d)
					}
					return true
				}
				return false
			}
		case <-time.After(5 * time.Second):
			t.Error("synchronizeDTEntityWithKeptn(): did not receive expected event")
			return true
		}
	}
}

func Test_getKeptnServiceName(t *testing.T) {
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
			got, err := getKeptnServiceName(tt.args.entity)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func Test_ServiceSynchronizer_addServiceToKeptn(t *testing.T) {

	servicesMockAPI := getTestServicesAPI()
	defer servicesMockAPI.Close()

	_, mockEventBroker := getTestMockEventBroker()
	defer mockEventBroker.Close()

	receivedServiceCreate, receivedSLO, receivedSLI, mockCS := getTestConfigService()
	defer mockCS.Close()
	os.Setenv(common.ShipyardControllerURLEnvironmentVariableName, mockCS.URL)

	type fields struct {
		logger              keptncommon.LoggerInterface
		projectsAPI         keptn.ProjectClientInterface
		servicesAPI         keptn.ServiceClientInterface
		resourcesAPI        keptn.SLIAndSLOResourceWriterInterface
		credentialsProvider credentials.DynatraceCredentialsProvider
		apiMutex            sync.Mutex
		EntitiesClient      func(*credentials.DynatraceCredentials) *dynatrace.EntitiesClient
		servicesInKeptn     []string
		configProvider      config.DynatraceConfigProvider
	}
	type args struct {
		serviceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "create service",
			fields: fields{
				logger:          keptncommon.NewLogger("", "", ""),
				projectsAPI:     nil,
				servicesAPI:     keptn.NewServiceClient(keptnapi.NewServiceHandler(servicesMockAPI.URL), mockCS.Client()),
				resourcesAPI:    keptn.NewResourceClient(keptn.NewConfigResourceClient(keptnapi.NewResourceHandler(mockCS.URL))),
				apiMutex:        sync.Mutex{},
				EntitiesClient:  nil,
				servicesInKeptn: []string{},
			},
			args: args{
				serviceName: "my-service",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ServiceSynchronizer{
				projectClient:       tt.fields.projectsAPI,
				servicesClient:      tt.fields.servicesAPI,
				resourcesClient:     tt.fields.resourcesAPI,
				credentialsProvider: tt.fields.credentialsProvider,
				EntitiesClientFunc:  tt.fields.EntitiesClient,
				servicesInKeptn:     tt.fields.servicesInKeptn,
				configProvider:      tt.fields.configProvider,
			}
			if err := s.addServiceToKeptn(tt.args.serviceName); (err != nil) != tt.wantErr {
				t.Errorf("ServiceSynchronizer.addServiceToKeptn() error = %v, wantErr %v", err, tt.wantErr)
			}

			select {
			case rec := <-receivedServiceCreate:
				if rec != tt.args.serviceName {
					t.Error("synchronizeDTEntityWithKeptn(): did not receive expected event")
				}
			case <-time.After(5 * time.Second):
				t.Error("synchronizeDTEntityWithKeptn(): did not receive expected event")
			}

			select {
			case rec := <-receivedSLO:
				if rec != tt.args.serviceName {
					t.Error("synchronizeDTEntityWithKeptn(): did not receive SLO file")
				}
			case <-time.After(5 * time.Second):
				t.Error("synchronizeDTEntityWithKeptn(): did not receive expected event")
			}

			select {
			case rec := <-receivedSLI:
				if rec != tt.args.serviceName {
					t.Error("synchronizeDTEntityWithKeptn(): did not receive SLI file")
				}
			case <-time.After(5 * time.Second):
				t.Error("synchronizeDTEntityWithKeptn(): did not receive expected event")
			}
		})
	}
}

package lib

import (
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-test/deep"
	"github.com/google/uuid"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
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

func Test_getKeptnServiceNameOfEntity(t *testing.T) {
	type args struct {
		entity entity
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "use service entity ID",
			args: args{
				entity: entity{
					EntityID:    "entity-id",
					DisplayName: ":10999",
					Tags:        nil,
				},
			},
			want: "entity-id",
		},
		{
			name: "use service entity ID because of invalid tag value",
			args: args{
				entity: entity{
					EntityID:    "entity-id",
					DisplayName: ":10999",
					Tags: []tags{
						{
							Context:              "CONTEXTLESS",
							Key:                  "keptn-service",
							StringRepresentation: "keptn_service:/invalid/service",
							Value:                "/invalid/service",
						},
					},
				},
			},
			want: "entity-id",
		},
		{
			name: "use keptn_service tag",
			args: args{
				entity: entity{
					EntityID:    "entity-id",
					DisplayName: ":10999",
					Tags: []tags{
						{
							Context:              "CONTEXTLESS",
							Key:                  "keptn_service",
							StringRepresentation: "keptn_service:my-service",
							Value:                "my-service",
						},
					},
				},
			},
			want: "my-service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getKeptnServiceNameOfEntity(tt.args.entity); got != tt.want {
				t.Errorf("getKeptnServiceNameOfEntity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serviceSynchronizer_fetchKeptnManagedServicesFromDynatrace(t *testing.T) {

	var returnedEntitiesResponse dtEntityListResponse
	dtMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		marshal, _ := json.Marshal(returnedEntitiesResponse)
		writer.WriteHeader(200)
		writer.Write(marshal)
	}))

	defer dtMockServer.Close()

	type fields struct {
		logger          keptncommon.LoggerInterface
		projectsAPI     *keptnapi.ProjectHandler
		servicesAPI     *keptnapi.ServiceHandler
		resourcesAPI    *keptnapi.ResourceHandler
		apiMutex        sync.Mutex
		DTHelper        *DynatraceHelper
		syncTimer       *time.Ticker
		keptnHandler    *keptnv2.Keptn
		servicesInKeptn []string
	}
	type args struct {
		nextPageKey string
		pageSize    int
	}
	tests := []struct {
		name                         string
		fields                       fields
		args                         args
		want                         *dtEntityListResponse
		returnedEntitiesListResponse dtEntityListResponse
		wantErr                      bool
	}{
		{
			name: "",
			fields: fields{
				logger:       keptncommon.NewLogger("", "", ""),
				projectsAPI:  nil,
				servicesAPI:  nil,
				resourcesAPI: nil,
				apiMutex:     sync.Mutex{},
				DTHelper: NewDynatraceHelper(nil, &credentials.DTCredentials{
					Tenant:   dtMockServer.URL,
					ApiToken: "",
				}, keptncommon.NewLogger("", "", "")),
				syncTimer:       nil,
				keptnHandler:    nil,
				servicesInKeptn: nil,
			},
			args: args{
				nextPageKey: "",
				pageSize:    1,
			},
			want: &dtEntityListResponse{
				TotalCount:  1,
				PageSize:    1,
				NextPageKey: "",
				Entities: []entity{
					{
						EntityID:    "1",
						DisplayName: "name",
						Tags: []tags{
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
				},
			},
			returnedEntitiesListResponse: dtEntityListResponse{
				TotalCount:  1,
				PageSize:    1,
				NextPageKey: "",
				Entities: []entity{
					{
						EntityID:    "1",
						DisplayName: "name",
						Tags: []tags{
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
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serviceSynchronizer{
				logger:          tt.fields.logger,
				projectsAPI:     tt.fields.projectsAPI,
				servicesAPI:     tt.fields.servicesAPI,
				resourcesAPI:    tt.fields.resourcesAPI,
				DTHelper:        tt.fields.DTHelper,
				syncTimer:       tt.fields.syncTimer,
				keptnHandler:    tt.fields.keptnHandler,
				servicesInKeptn: tt.fields.servicesInKeptn,
			}
			returnedEntitiesResponse = tt.returnedEntitiesListResponse
			got, err := s.fetchKeptnManagedServicesFromDynatrace(tt.args.nextPageKey, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchKeptnManagedServicesFromDynatrace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); len(diff) > 0 {
				t.Errorf("fetchKeptnManagedServicesFromDynatrace() got = %v, want %v", got, tt.want)
				for _, d := range diff {
					t.Log(d)
				}
			}
		})
	}
}

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

func Test_serviceSynchronizer_synchronizeDTEntityWithKeptn(t *testing.T) {

	servicesMockAPI := getTestServicesAPI()
	defer servicesMockAPI.Close()

	_, mockEventBroker := getTestMockEventBroker()
	defer mockEventBroker.Close()

	receivedServiceCreate, receivedSLO, receivedSLI, mockCS := getTestConfigService()
	defer mockCS.Close()

	k := getTestKeptnHandler(mockCS, mockEventBroker)

	type fields struct {
		logger          keptncommon.LoggerInterface
		projectsAPI     *keptnapi.ProjectHandler
		servicesAPI     *keptnapi.ServiceHandler
		resourcesAPI    *keptnapi.ResourceHandler
		apiMutex        sync.Mutex
		DTHelper        *DynatraceHelper
		syncTimer       *time.Ticker
		keptnHandler    *keptnv2.Keptn
		servicesInKeptn []string
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
				servicesAPI:     keptnapi.NewServiceHandler(servicesMockAPI.URL),
				resourcesAPI:    keptnapi.NewResourceHandler(mockCS.URL),
				apiMutex:        sync.Mutex{},
				DTHelper:        nil,
				syncTimer:       nil,
				keptnHandler:    k,
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
			s := &serviceSynchronizer{
				logger:          tt.fields.logger,
				projectsAPI:     tt.fields.projectsAPI,
				servicesAPI:     tt.fields.servicesAPI,
				resourcesAPI:    tt.fields.resourcesAPI,
				DTHelper:        tt.fields.DTHelper,
				syncTimer:       tt.fields.syncTimer,
				keptnHandler:    tt.fields.keptnHandler,
				servicesInKeptn: tt.fields.servicesInKeptn,
			}
			if err := s.synchronizeDTEntityWithKeptn(tt.args.serviceName); (err != nil) != tt.wantErr {
				t.Errorf("synchronizeDTEntityWithKeptn() error = %v, wantErr %v", err, tt.wantErr)
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
		if *event.Type == keptn.InternalServiceCreateEventType {
			bytes, err = json.Marshal(event.Data)
			if err != nil {
				return
			}
			serviceCreateData := &keptn.ServiceCreateEventData{}
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
	createServiceData := keptn.ServiceCreateEventData{
		Project: defaultDTProjectName,
		Service: "my-service",
	}
	ce := cloudevents.NewEvent()
	ce.SetType(keptn.InternalServiceCreateEventType)
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

func Test_serviceSynchronizer_synchronizeServices(t *testing.T) {

	firstDTResponse := dtEntityListResponse{
		TotalCount:  3,
		PageSize:    2,
		NextPageKey: "next-page-key",
		Entities: []entity{
			{
				EntityID:    "1",
				DisplayName: "name",
				Tags: []tags{
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
				Tags: []tags{
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

	secondDTResponse := dtEntityListResponse{
		TotalCount:  2,
		PageSize:    1,
		NextPageKey: "",
		Entities: []entity{
			{
				EntityID:    "2",
				DisplayName: "name",
				Tags: []tags{
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

	os.Setenv(shipyardController, mockCS.URL)

	k := getTestKeptnHandler(mockCS, mockEventBroker)
	s := &serviceSynchronizer{
		logger:       keptncommon.NewLogger("", "", ""),
		projectsAPI:  keptnapi.NewProjectHandler(projectsMockAPI.URL),
		servicesAPI:  keptnapi.NewServiceHandler(servicesMockAPI.URL),
		resourcesAPI: keptnapi.NewResourceHandler(mockCS.URL),
		DTHelper: NewDynatraceHelper(nil, &credentials.DTCredentials{
			Tenant:   dtMockServer.URL,
			ApiToken: "",
		}, keptncommon.NewLogger("", "", "")),
		syncTimer:       nil,
		keptnHandler:    k,
		servicesInKeptn: []string{},
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

package onboard

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-test/deep"
	"github.com/google/uuid"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
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
			Project: synchronizedProject,
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

func (f *mockEntitiesClientFactory) GetEntitiesClient() (*dynatrace.EntitiesClient, error) {
	dynatraceCredentials, err := credentials.NewDynatraceCredentials(f.url, testDynatraceAPIToken)
	if err != nil {
		return nil, err
	}

	dynatraceClient := dynatrace.NewClientWithHTTP(dynatraceCredentials, f.httpClient)
	return dynatrace.NewEntitiesClient(dynatraceClient), nil
}

func Test_ServiceSynchronizer_synchronizeServices(t *testing.T) {

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

	receivedServiceCreate, receivedSLO, receivedSLI, mockCS := getTestConfigService()
	defer mockCS.Close()

	os.Setenv(common.ShipyardControllerURLEnvironmentVariableName, mockCS.URL)

	mockEntitiesClientFactory, teardown := newMockEntitiesClientFactory(t)
	defer teardown()

	const testDataFolder = "./testdata/test_synchronize_services/"
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?entitySelector=type(\"SERVICE\")%20AND%20tag(\"keptn_managed\",\"[Environment]keptn_managed\")%20AND%20tag(\"keptn_service\",\"[Environment]keptn_service\")&fields=+tags&pageSize=50", testDataFolder+"entities_response1.json")
	mockEntitiesClientFactory.handler.AddExact("/api/v2/entities?nextPageKey=next-page-key", testDataFolder+"entities_response2.json")

	s := &ServiceSynchronizer{
		servicesClient:        keptn.NewServiceClient(keptnapi.NewServiceHandler(servicesMockAPI.URL), mockCS.Client()),
		resourcesClient:       keptn.NewResourceClient(keptn.NewConfigResourceClient(keptnapi.NewResourceHandler(mockCS.URL))),
		entitiesClientFactory: mockEntitiesClientFactory,
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

func Test_ServiceSynchronizer_addServiceToKeptn(t *testing.T) {

	servicesMockAPI := getTestServicesAPI()
	defer servicesMockAPI.Close()

	receivedServiceCreate, receivedSLO, receivedSLI, mockCS := getTestConfigService()
	defer mockCS.Close()
	os.Setenv(common.ShipyardControllerURLEnvironmentVariableName, mockCS.URL)

	serviceName := "my-service"

	s := &ServiceSynchronizer{
		servicesClient:  keptn.NewServiceClient(keptnapi.NewServiceHandler(servicesMockAPI.URL), mockCS.Client()),
		resourcesClient: keptn.NewResourceClient(keptn.NewConfigResourceClient(keptnapi.NewResourceHandler(mockCS.URL))),
	}

	err := s.addServiceToKeptn(serviceName)
	assert.NoError(t, err)

	select {
	case rec := <-receivedServiceCreate:
		assert.EqualValues(t, serviceName, rec, "did not receive expected event")
	case <-time.After(5 * time.Second):
		t.Error("did not receive expected event")
	}

	select {
	case rec := <-receivedSLO:
		assert.EqualValues(t, serviceName, rec, "did not receive SLO file")
	case <-time.After(5 * time.Second):
		t.Error("did not receive expected event")
	}

	select {
	case rec := <-receivedSLI:
		assert.EqualValues(t, serviceName, rec, "did not receive SLI file")
	case <-time.After(5 * time.Second):
		t.Error("did not receive expected event")
	}
}

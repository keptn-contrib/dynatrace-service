package lib

import (
	"encoding/json"
	"github.com/go-test/deep"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"net/http"
	"net/http/httptest"
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
		logger          keptn.LoggerInterface
		projectsAPI     *keptnapi.ProjectHandler
		servicesAPI     *keptnapi.ServiceHandler
		resourcesAPI    *keptnapi.ResourceHandler
		apiMutex        sync.Mutex
		DTHelper        *DynatraceHelper
		syncTimer       *time.Ticker
		keptnHandler    *keptn.Keptn
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
				logger:       keptn.NewLogger("", "", ""),
				projectsAPI:  nil,
				servicesAPI:  nil,
				resourcesAPI: nil,
				apiMutex:     sync.Mutex{},
				DTHelper: NewDynatraceHelper(nil, &credentials.DTCredentials{
					Tenant:   dtMockServer.URL,
					ApiToken: "",
				}, keptn.NewLogger("", "", "")),
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
				apiMutex:        tt.fields.apiMutex,
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

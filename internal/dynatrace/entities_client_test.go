package dynatrace

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

func TestEntitiesClient_GetKeptnManagedServices(t *testing.T) {
	var returnedEntitiesResponse EntitiesResponse
	dtMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		marshal, _ := json.Marshal(returnedEntitiesResponse)
		writer.WriteHeader(200)
		writer.Write(marshal)
	}))

	defer dtMockServer.Close()

	type fields struct {
		client ClientInterface
	}
	type args struct {
		nextPageKey string
		pageSize    int
	}
	tests := []struct {
		name                         string
		fields                       fields
		args                         args
		want                         []Entity
		returnedEntitiesListResponse EntitiesResponse
		wantErr                      bool
	}{
		{
			name: "",
			fields: fields{
				client: NewClient(
					&credentials.DynatraceCredentials{
						Tenant:   dtMockServer.URL,
						ApiToken: "",
					}),
			},
			args: args{
				nextPageKey: "",
				pageSize:    1,
			},
			want: []Entity{
				{
					EntityID:    "1",
					DisplayName: "name",
					Tags: []Tag{
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
			returnedEntitiesListResponse: EntitiesResponse{
				TotalCount:  1,
				PageSize:    1,
				NextPageKey: "",
				Entities: []Entity{
					{
						EntityID:    "1",
						DisplayName: "name",
						Tags: []Tag{
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
			ec := &EntitiesClient{
				Client: tt.fields.client,
			}
			returnedEntitiesResponse = tt.returnedEntitiesListResponse
			got, err := ec.GetKeptnManagedServices()
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

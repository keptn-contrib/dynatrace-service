package dynatrace

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestEntitiesClient_GetKeptnManagedServices(t *testing.T) {
	var returnedEntitiesResponse EntitiesResponse
	dtMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		marshal, _ := json.Marshal(returnedEntitiesResponse)
		writer.WriteHeader(200)
		writer.Write(marshal)
	}))

	mockCredentials := createDynatraceCredentials(t, dtMockServer.URL)

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
				client: NewClient(mockCredentials),
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
			got, err := ec.GetKeptnManagedServices(context.TODO())
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

func TestEntitiesClient_GetAllPGIsForKeptnServices(t *testing.T) {
	const testdataFolder = "./testdata/entities_client/"
	const url = "a/api/v2/entities?entitySelector=type%28%22process_group_instance%22%29%2CtoRelationship.runsOnProcessGroupInstance%28type%28SERVICE%29%2Ctag%28%22keptn_project%3Apod-tato-head%22%29%2Ctag%28%22keptn_stage%3Ahardening%22%29%2Ctag%28%22keptn_service%3Ahelloservice%22%29%29&from=1654000200000&to=1654000320000"

	cfg := PGIQueryConfig{
		project: "pod-tato-head",
		stage:   "hardening",
		service: "helloservice",
		from:    time.Date(2022, 5, 31, 12, 30, 0, 0, time.UTC),
		to:      time.Date(2022, 5, 31, 12, 32, 0, 0, time.UTC),
	}

	tests := []struct {
		name         string
		fileName     string
		expectedPGIs []string
	}{
		{
			name:     "multiple entities returned",
			fileName: "multiple_entities.json",
			expectedPGIs: []string{
				"PROCESS_GROUP_INSTANCE-95C5FBF859599282",
				"PROCESS_GROUP_INSTANCE-D23E64F62FDC200A",
				"PROCESS_GROUP_INSTANCE-DE323A8B8449D009",
				"PROCESS_GROUP_INSTANCE-F59D42FEA235E5F9",
			},
		},
		{
			name:     "single entity returned",
			fileName: "single_entity.json",
			expectedPGIs: []string{
				"PROCESS_GROUP_INSTANCE-D23E64F62FDC200A",
			},
		},
		{
			name:         "single entity returned",
			fileName:     "no_entity.json",
			expectedPGIs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := test.NewFileBasedURLHandler(t)
			handler.AddExact(url, testdataFolder+tt.fileName)

			client, teardown := createEventsClient(t, handler)
			defer teardown()

			actualPGIs, err := client.GetAllPGIsForKeptnServices(context.Background(), cfg)
			if assert.NoError(t, err) {
				assert.EqualValues(t, actualPGIs, tt.expectedPGIs)
			}
		})
	}
}

func createEventsClient(t *testing.T, handler http.Handler) (*EntitiesClient, func()) {
	dynatraceClient, _, teardown := createDynatraceClient(t, handler)

	ec := NewEntitiesClient(dynatraceClient)

	return ec, teardown
}

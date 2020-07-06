package common

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestCheckKeptnConnection(t *testing.T) {

	var returnedResponse int
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(returnedResponse)
			w.Write([]byte(`{}`))
		}),
	)
	defer ts.Close()

	type args struct {
		keptnCredentials *KeptnCredentials
	}
	tests := []struct {
		name             string
		args             args
		returnedResponse int
		wantErr          bool
	}{
		{
			name: "Successful connection",
			args: args{
				keptnCredentials: &KeptnCredentials{
					ApiURL:   ts.URL,
					ApiToken: "my-test-token",
				},
			},
			returnedResponse: 200,
			wantErr:          false,
		},
		{
			name: "unauthorized connection",
			args: args{
				keptnCredentials: &KeptnCredentials{
					ApiURL:   ts.URL,
					ApiToken: "my-test-token",
				},
			},
			returnedResponse: 401,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedResponse = tt.returnedResponse
			if err := CheckKeptnConnection(tt.args.keptnCredentials); (err != nil) != tt.wantErr {
				t.Errorf("CheckKeptnConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetKeptnCredentials(t *testing.T) {
	tests := []struct {
		name           string
		want           *KeptnCredentials
		wantErr        bool
		apiURLEnvVar   string
		apiTokenEnvVar string
	}{
		{
			name:           "return error if required environment variables are not set",
			want:           nil,
			wantErr:        true,
			apiURLEnvVar:   "",
			apiTokenEnvVar: "",
		},
		{
			name: "return credentials with https://",
			want: &KeptnCredentials{
				ApiURL:   "https://api.keptn.test.com",
				ApiToken: "1234",
			},
			wantErr:        false,
			apiURLEnvVar:   "api.keptn.test.com",
			apiTokenEnvVar: "1234",
		},
		{
			name: "return credentials with https://",
			want: &KeptnCredentials{
				ApiURL:   "https://api.keptn.test.com",
				ApiToken: "1234",
			},
			wantErr:        false,
			apiURLEnvVar:   "https://api.keptn.test.com",
			apiTokenEnvVar: "1234",
		},
		{
			name: "return credentials with http://",
			want: &KeptnCredentials{
				ApiURL:   "http://api.keptn.test.com",
				ApiToken: "1234",
			},
			wantErr:        false,
			apiURLEnvVar:   "http://api.keptn.test.com",
			apiTokenEnvVar: "1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("KEPTN_API_URL", tt.apiURLEnvVar)
			os.Setenv("KEPTN_API_TOKEN", tt.apiTokenEnvVar)
			got, err := GetKeptnCredentials()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeptnCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeptnCredentials() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKeptnBridgeURL(t *testing.T) {
	tests := []struct {
		name            string
		want            string
		wantErr         bool
		bridgeURLEnvVar string
	}{
		{
			name:            "return bridge URL",
			want:            "https://bridge.keptn",
			wantErr:         false,
			bridgeURLEnvVar: "bridge.keptn",
		},
		{
			name:            "return bridge URL",
			want:            "https://bridge.keptn",
			wantErr:         false,
			bridgeURLEnvVar: "https://bridge.keptn",
		},
		{
			name:            "return bridge URL with http",
			want:            "http://bridge.keptn",
			wantErr:         false,
			bridgeURLEnvVar: "http://bridge.keptn",
		},
		{
			name:            "return error if env var not set",
			want:            "",
			wantErr:         true,
			bridgeURLEnvVar: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("KEPTN_BRIDGE_URL", tt.bridgeURLEnvVar)
			got, err := GetKeptnBridgeURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeptnBridgeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetKeptnBridgeURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

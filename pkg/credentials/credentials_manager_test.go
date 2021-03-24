package credentials

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"net/http/httptest"
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
		KeptnAPICredentials *KeptnAPICredentials
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
				KeptnAPICredentials: &KeptnAPICredentials{
					APIURL:   ts.URL,
					APIToken: "my-test-token",
				},
			},
			returnedResponse: 200,
			wantErr:          false,
		},
		{
			name: "unauthorized connection",
			args: args{
				KeptnAPICredentials: &KeptnAPICredentials{
					APIURL:   ts.URL,
					APIToken: "my-test-token",
				},
			},
			returnedResponse: 401,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedResponse = tt.returnedResponse
			if err := CheckKeptnConnection(tt.args.KeptnAPICredentials); (err != nil) != tt.wantErr {
				t.Errorf("CheckKeptnConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetKeptnAPICredentials(t *testing.T) {
	tests := []struct {
		name           string
		want           *KeptnAPICredentials
		wantErr        bool
		APIURLEnvVar   string
		APITokenEnvVar string
	}{
		{
			name:           "return error if required environment variables are not set",
			want:           nil,
			wantErr:        true,
			APIURLEnvVar:   "",
			APITokenEnvVar: "",
		},
		{
			name: "return credentials with https://",
			want: &KeptnAPICredentials{
				APIURL:   "https://api.keptn.test.com",
				APIToken: "1234",
			},
			wantErr:        false,
			APIURLEnvVar:   "api.keptn.test.com",
			APITokenEnvVar: "1234",
		},
		{
			name: "return credentials with https://",
			want: &KeptnAPICredentials{
				APIURL:   "https://api.keptn.test.com",
				APIToken: "1234",
			},
			wantErr:        false,
			APIURLEnvVar:   "https://api.keptn.test.com",
			APITokenEnvVar: "1234",
		},
		{
			name: "return credentials with http://",
			want: &KeptnAPICredentials{
				APIURL:   "http://api.keptn.test.com",
				APIToken: "1234",
			},
			wantErr:        false,
			APIURLEnvVar:   "http://api.keptn.test.com",
			APITokenEnvVar: "1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeClient := fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dynatrace",
					Namespace: "keptn",
				},
				Data: map[string][]byte{
					"KEPTN_API_URL":   []byte(tt.APIURLEnvVar),
					"KEPTN_API_TOKEN": []byte(tt.APITokenEnvVar),
				},
			})

			k8sSecretReader, _ := NewK8sCredentialReader(fakeClient)

			cm, err := NewCredentialManager(k8sSecretReader)
			if err != nil {
				t.Errorf("could not initialize CredentialManager: %s", err.Error())
			}

			got, err := cm.GetKeptnAPICredentials()
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
			fakeClient := fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dynatrace",
					Namespace: "keptn",
				},
				Data: map[string][]byte{
					"KEPTN_BRIDGE_URL": []byte(tt.bridgeURLEnvVar),
				},
			})

			k8sSecretReader, _ := NewK8sCredentialReader(fakeClient)

			cm, err := NewCredentialManager(k8sSecretReader)
			if err != nil {
				t.Errorf("could not initialize CredentialManager: %s", err.Error())
			}
			got, err := cm.GetKeptnBridgeURL()
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

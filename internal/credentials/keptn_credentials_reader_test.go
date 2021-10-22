package credentials

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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
		apiURL   string
		apiToken string
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
				apiURL:   ts.URL,
				apiToken: "my-test-token",
			},
			returnedResponse: 200,
			wantErr:          false,
		},
		{
			name: "unauthorized connection",
			args: args{
				apiURL:   ts.URL,
				apiToken: "my-test-token",
			},
			returnedResponse: 401,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedResponse = tt.returnedResponse
			keptnCredentials, err := NewKeptnCredentials(tt.args.apiURL, tt.args.apiToken)
			assert.NoError(t, err)
			if err := CheckKeptnConnection(keptnCredentials); (err != nil) != tt.wantErr {
				t.Errorf("CheckKeptnConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetKeptnAPICredentials(t *testing.T) {

	wantHTTPSKeptnCredentials, err := NewKeptnCredentials("https://api.keptn.test.com", "1234")
	assert.NoError(t, err)

	wantHTTPKeptnCredentials, err := NewKeptnCredentials("http://api.keptn.test.com", "1234")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		want           *KeptnCredentials
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
			name:           "return credentials with https://",
			want:           wantHTTPSKeptnCredentials,
			wantErr:        false,
			APIURLEnvVar:   "api.keptn.test.com",
			APITokenEnvVar: "1234",
		},
		{
			name:           "return credentials with https://",
			want:           wantHTTPSKeptnCredentials,
			wantErr:        false,
			APIURLEnvVar:   "https://api.keptn.test.com",
			APITokenEnvVar: "1234",
		},
		{
			name:           "return credentials with http://",
			want:           wantHTTPKeptnCredentials,
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

			k8sSecretReader := NewK8sSecretReader(fakeClient)

			cm := NewKeptnCredentialsReader(k8sSecretReader)

			got, err := cm.GetKeptnCredentials()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
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
			wantErr:         false,
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

			k8sSecretReader := NewK8sSecretReader(fakeClient)

			cm := NewKeptnCredentialsReader(k8sSecretReader)
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

// Test keptn api credential behavior: values in dynatrace secret should be used, if not available, fall back to environment variables
// If neither is available, an error should be produced.
func TestCredentialManager_GetKeptnAPICredentials(t *testing.T) {

	wantKeptnCredentials, err := NewKeptnCredentials("https://mySampleEnv.live.dynatrace.com", "abc123")
	assert.NoError(t, err)
	wantOtherKeptnCredentials, err := NewKeptnCredentials("https://otherSampleEnv.live.dynatrace.com", "def456")
	assert.NoError(t, err)

	dynatraceSecret := createDynatraceKeptnSecret("dynatrace", "keptn", "https://mySampleEnv.live.dynatrace.com", "abc123", "https://mySampleEnv.live.dynatrace.com/bridge")
	otherDynatraceSecret := createDynatraceKeptnSecret("dynatrace_other", "keptn", "https://sampleEnv.live.dynatrace.com", "xyz000", "https://sampleEnv.live.dynatrace.com/bridge")

	type envVars struct {
		keptnAPIURL   string
		keptnAPIToken string
	}

	tests := []struct {
		name    string
		secret  *v1.Secret
		envVars envVars
		want    *KeptnCredentials
		wantErr bool
	}{
		{
			name:    "no secret, no env vars",
			secret:  &v1.Secret{},
			wantErr: true,
		},
		{
			name:    "with secret, no env vars",
			secret:  dynatraceSecret,
			want:    wantKeptnCredentials,
			wantErr: false,
		},
		{
			name:    "with secret, with env vars",
			secret:  dynatraceSecret,
			envVars: envVars{keptnAPIURL: "https://otherSampleEnv.live.dynatrace.com", keptnAPIToken: "def456"},
			want:    wantKeptnCredentials,
			wantErr: false,
		},
		{
			name:    "no secret, with env vars",
			secret:  &v1.Secret{},
			envVars: envVars{keptnAPIURL: "https://otherSampleEnv.live.dynatrace.com", keptnAPIToken: "def456"},
			want:    wantOtherKeptnCredentials,
			wantErr: false,
		},
		{
			name:    "with other secret, no env vars",
			secret:  otherDynatraceSecret,
			wantErr: true,
		},
		{
			name:    "with other secret, with env vars",
			secret:  otherDynatraceSecret,
			envVars: envVars{keptnAPIURL: "https://otherSampleEnv.live.dynatrace.com", keptnAPIToken: "def456"},
			want:    wantOtherKeptnCredentials,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretReader := NewK8sSecretReader(fake.NewSimpleClientset(tt.secret))

			os.Setenv("KEPTN_API_URL", tt.envVars.keptnAPIURL)
			os.Setenv("KEPTN_API_TOKEN", tt.envVars.keptnAPIToken)
			defer func() {
				os.Unsetenv("KEPTN_API_URL")
				os.Unsetenv("KEPTN_API_TOKEN")
			}()

			cm := NewKeptnCredentialsReader(secretReader)
			got, err := cm.GetKeptnCredentials()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
		})
	}
}

// Test keptn bridge URL behavior: value in dynatrace secret should be used, if not available, fall back to environment variable
// If neither is available, an error should be produced.
func TestCredentialManager_GetKeptnBridgeURL(t *testing.T) {

	dynatraceSecret := createDynatraceKeptnSecret("dynatrace", "keptn", "https://mySampleEnv.live.dynatrace.com", "abc123", "https://mySampleEnv.live.dynatrace.com/bridge")
	otherDynatraceSecret := createDynatraceKeptnSecret("dynatrace_other", "keptn", "https://sampleEnv.live.dynatrace.com", "xyz000", "https://sampleEnv.live.dynatrace.com/bridge")

	type envVars struct {
		keptnBridgeURL string
	}

	tests := []struct {
		name    string
		secret  *v1.Secret
		envVars envVars
		want    string
		wantErr bool
	}{
		{
			name:    "no secret, no env vars",
			secret:  &v1.Secret{},
			wantErr: false,
		},
		{
			name:    "with secret, no env vars",
			secret:  dynatraceSecret,
			want:    "https://mySampleEnv.live.dynatrace.com/bridge",
			wantErr: false,
		},
		{
			name:    "with secret, with env vars",
			secret:  dynatraceSecret,
			envVars: envVars{keptnBridgeURL: "https://sampleEnv.live.dynatrace.com/bridge"},
			want:    "https://mySampleEnv.live.dynatrace.com/bridge",
			wantErr: false,
		},
		{
			name:    "no secret, with env vars",
			secret:  &v1.Secret{},
			envVars: envVars{keptnBridgeURL: "https://sampleEnv.live.dynatrace.com/bridge"},
			want:    "https://sampleEnv.live.dynatrace.com/bridge",
			wantErr: false,
		},
		{
			name:    "with other secret, no env vars",
			secret:  otherDynatraceSecret,
			wantErr: false,
		},
		{
			name:    "with other secret, with env vars",
			secret:  otherDynatraceSecret,
			envVars: envVars{keptnBridgeURL: "https://sampleEnv.live.dynatrace.com/bridge"},
			want:    "https://sampleEnv.live.dynatrace.com/bridge",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretReader := NewK8sSecretReader(fake.NewSimpleClientset(tt.secret))

			os.Setenv("KEPTN_BRIDGE_URL", tt.envVars.keptnBridgeURL)
			defer func() {
				os.Unsetenv("KEPTN_BRIDGE_URL")
			}()

			cm := NewKeptnCredentialsReader(secretReader)
			got, err := cm.GetKeptnBridgeURL()
			if (err != nil) && tt.wantErr {
				return
			} else if (err != nil) != tt.wantErr {
				t.Fatalf("CredentialManager.GetKeptnBridgeURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("CredentialManager.GetKeptnBridgeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createDynatraceKeptnSecret(name string, namespace string, keptnAPIURL string, keptnAPIToken string, KeptnBridgeURL string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"KEPTN_API_URL":    []byte(keptnAPIURL),
			"KEPTN_API_TOKEN":  []byte(keptnAPIToken),
			"KEPTN_BRIDGE_URL": []byte(KeptnBridgeURL),
		},
	}
}

package credentials

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

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
		KeptnAPICredentials *KeptnCredentials
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
				KeptnAPICredentials: &KeptnCredentials{
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
				KeptnAPICredentials: &KeptnCredentials{
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
			name: "return credentials with https://",
			want: &KeptnCredentials{
				APIURL:   "https://api.keptn.test.com",
				APIToken: "1234",
			},
			wantErr:        false,
			APIURLEnvVar:   "api.keptn.test.com",
			APITokenEnvVar: "1234",
		},
		{
			name: "return credentials with https://",
			want: &KeptnCredentials{
				APIURL:   "https://api.keptn.test.com",
				APIToken: "1234",
			},
			wantErr:        false,
			APIURLEnvVar:   "https://api.keptn.test.com",
			APITokenEnvVar: "1234",
		},
		{
			name: "return credentials with http://",
			want: &KeptnCredentials{
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

// Test dynatrace credential behavior: values should be read from dynatrace secret unless secret name has been overridden in dynatrace config file.
// If neither is available, an error should be produced.
func TestCredentialManager_GetDynatraceCredentials(t *testing.T) {

	dynatraceSecret := createDynatraceDTSecret("dynatrace", "keptn", "https://mySampleEnv.live.dynatrace.com", "abc123")
	dynatraceOtherSecret := createDynatraceDTSecret("dynatrace_other", "keptn", "https://mySampleEnv.live.dynatrace.com", "abc123")

	type args struct {
		secretName string
	}
	tests := []struct {
		name    string
		secret  *v1.Secret
		args    args
		want    *DynatraceCredentials
		wantErr bool
	}{
		{
			name:   "with no secret, no config",
			secret: &v1.Secret{},
			args: args{
				secretName: "",
			},
			wantErr: true,
		},
		{
			name:   "with dynatrace secret, no config",
			secret: dynatraceSecret,
			args: args{
				secretName: "",
			},
			want: &DynatraceCredentials{
				Tenant:   "https://mySampleEnv.live.dynatrace.com",
				ApiToken: "abc123",
			},
			wantErr: false,
		},
		{
			name:   "with dynatrace_other secret, with good config",
			secret: dynatraceOtherSecret,
			args: args{
				secretName: "dynatrace_other",
			},
			want: &DynatraceCredentials{
				Tenant:   "https://mySampleEnv.live.dynatrace.com",
				ApiToken: "abc123",
			},
			wantErr: false,
		},
		{
			name:   "with dynatrace_other secret, with bad config",
			secret: dynatraceOtherSecret,
			args: args{
				secretName: "dynatrace_other2",
			},
			wantErr: true,
		},
		{
			name:   "with dynatrace_other secret, no config",
			secret: dynatraceOtherSecret,
			args: args{
				secretName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			secretReader, err := NewK8sCredentialReader(fake.NewSimpleClientset(tt.secret))
			if err != nil {
				t.Fatalf("NewK8sCredentialReader() error = %v", err)
			}
			cm, err := NewCredentialManager(secretReader)
			if err != nil {
				t.Fatalf("NewCredentialManager() error = %v", err)
			}
			decorator := NewCredentialManagerDefaultFallbackDecorator(cm)

			got, err := decorator.GetDynatraceCredentials(tt.args.secretName)
			if (err != nil) && tt.wantErr {
				return
			} else if (err != nil) != tt.wantErr {
				t.Fatalf("CredentialManager.GetDynatraceCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CredentialManager.GetDynatraceCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createDynatraceDTSecret(name string, namespace string, dtTenant string, dtAPIToken string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"DT_TENANT":    []byte(dtTenant),
			"DT_API_TOKEN": []byte(dtAPIToken),
		},
	}
}

// Test keptn api credential behavior: values in dynatrace secret should be used, if not available, fall back to environment variables
// If neither is available, an error should be produced.
func TestCredentialManager_GetKeptnAPICredentials(t *testing.T) {

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
			want:    &KeptnCredentials{APIURL: "https://mySampleEnv.live.dynatrace.com", APIToken: "abc123"},
			wantErr: false,
		},
		{
			name:    "with secret, with env vars",
			secret:  dynatraceSecret,
			envVars: envVars{keptnAPIURL: "https://otherSampleEnv.live.dynatrace.com", keptnAPIToken: "def456"},
			want:    &KeptnCredentials{APIURL: "https://mySampleEnv.live.dynatrace.com", APIToken: "abc123"},
			wantErr: false,
		},
		{
			name:    "no secret, with env vars",
			secret:  &v1.Secret{},
			envVars: envVars{keptnAPIURL: "https://otherSampleEnv.live.dynatrace.com", keptnAPIToken: "def456"},
			want:    &KeptnCredentials{APIURL: "https://otherSampleEnv.live.dynatrace.com", APIToken: "def456"},
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
			want:    &KeptnCredentials{APIURL: "https://otherSampleEnv.live.dynatrace.com", APIToken: "def456"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretReader, err := NewK8sCredentialReader(fake.NewSimpleClientset(tt.secret))
			if err != nil {
				t.Fatalf("NewK8sCredentialReader() error = %v", err)
			}

			os.Setenv("KEPTN_API_URL", tt.envVars.keptnAPIURL)
			os.Setenv("KEPTN_API_TOKEN", tt.envVars.keptnAPIToken)
			defer func() {
				os.Unsetenv("KEPTN_API_URL")
				os.Unsetenv("KEPTN_API_TOKEN")
			}()

			cm, err := NewCredentialManager(secretReader)
			if err != nil {
				t.Fatalf("NewCredentialManager() error = %v", err)
			}

			got, err := cm.GetKeptnAPICredentials()

			if (err != nil) && tt.wantErr {
				return
			} else if (err != nil) != tt.wantErr {
				t.Fatalf("CredentialManager.GetKeptnAPICredentials() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CredentialManager.GetKeptnAPICredentials() = %v, want %v", got, tt.want)
			}
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
			wantErr: true,
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
			wantErr: true,
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
			secretReader, err := NewK8sCredentialReader(fake.NewSimpleClientset(tt.secret))
			if err != nil {
				t.Fatalf("NewK8sCredentialReader() error = %v", err)
			}

			os.Setenv("KEPTN_BRIDGE_URL", tt.envVars.keptnBridgeURL)
			defer func() {
				os.Unsetenv("KEPTN_BRIDGE_URL")
			}()

			cm, err := NewCredentialManager(secretReader)
			if err != nil {
				t.Fatalf("NewCredentialManager() error = %v", err)
			}
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

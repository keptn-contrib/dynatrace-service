package credentials

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Test keptn api credential behavior: values in dynatrace secret should be used, if not available, fall back to environment variables
// If neither is available, an error should be produced.
func TestKeptnCredentialsReader_GetKeptnCredentials(t *testing.T) {

	wantKeptnCredentials, err := NewKeptnCredentials("https://keptn.test.com/api", "abc123", "https://keptn.test.com/bridge")
	assert.NoError(t, err)
	wantKeptn2Credentials, err := NewKeptnCredentials("http://keptn2.test.com/api", "abc123", "http://keptn2.test.com/bridge")
	assert.NoError(t, err)
	wantOtherKeptnCredentials, err := NewKeptnCredentials("https://keptn.other.com/api", "def456", "https://keptn.other.com/bridge")
	assert.NoError(t, err)
	wantKeptnCredentialsNoBridgeURL, err := NewKeptnCredentials("https://keptn.test.com/api", "abc123", "")
	assert.NoError(t, err)

	tests := []struct {
		name                 string
		secretName           string
		secretData           map[string]string
		environmentVariables map[string]string
		want                 *KeptnCredentials
		wantErr              bool
	}{
		{
			name:    "no secret, no env vars",
			wantErr: true,
		},
		{
			name:       "with secret, no env vars - valid URLs",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL":    "https://keptn.test.com/api",
				"KEPTN_API_TOKEN":  "abc123",
				"KEPTN_BRIDGE_URL": "https://keptn.test.com/bridge",
			},
			want:    wantKeptnCredentials,
			wantErr: false,
		},
		{
			name:       "with secret, no env vars - invalid URL",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL":    "/keptn.test.com/api",
				"KEPTN_API_TOKEN":  "abc123",
				"KEPTN_BRIDGE_URL": "https://keptn.test.com/bridge",
			},
			wantErr: true,
		},
		{
			name:       "with secret, no env vars - invalid scheme",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL":    "https://keptn.test.com/api",
				"KEPTN_API_TOKEN":  "abc123",
				"KEPTN_BRIDGE_URL": "ftp://keptn.test.com/bridge",
			},
			wantErr: true,
		},
		{
			name:       "with secret, no env vars - no API token",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL":    "https://keptn.test.com/api",
				"KEPTN_BRIDGE_URL": "https://keptn.test.com/bridge",
			},
			wantErr: true,
		},
		{
			name:       "with secret, no env vars - assume HTTPS",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL":    "keptn.test.com/api",
				"KEPTN_API_TOKEN":  "abc123",
				"KEPTN_BRIDGE_URL": "keptn.test.com/bridge",
			},
			want:    wantKeptnCredentials,
			wantErr: false,
		},
		{
			name:       "with secret, no env vars - explicit HTTP",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL":    "http://keptn2.test.com/api",
				"KEPTN_API_TOKEN":  "abc123",
				"KEPTN_BRIDGE_URL": "http://keptn2.test.com/bridge",
			},
			want:    wantKeptn2Credentials,
			wantErr: false,
		},
		{
			name:       "with secret, with env vars - secret preferred",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL":    "https://keptn.test.com/api",
				"KEPTN_API_TOKEN":  "abc123",
				"KEPTN_BRIDGE_URL": "https://keptn.test.com/bridge",
			},
			environmentVariables: map[string]string{
				"KEPTN_API_URL":   "https://keptn.other.com/api",
				"KEPTN_API_TOKEN": "def456",
			},
			want:    wantKeptnCredentials,
			wantErr: false,
		},
		{
			name: "no secret, with env vars",
			environmentVariables: map[string]string{
				"KEPTN_API_URL":    "https://keptn.other.com/api",
				"KEPTN_API_TOKEN":  "def456",
				"KEPTN_BRIDGE_URL": "https://keptn.other.com/bridge",
			},
			want:    wantOtherKeptnCredentials,
			wantErr: false,
		},
		{
			name:       "mixed, no bridge URL",
			secretName: "dynatrace",
			secretData: map[string]string{
				"KEPTN_API_URL": "https://keptn.test.com/api",
			},
			environmentVariables: map[string]string{
				"KEPTN_API_TOKEN": "abc123",
			},
			want:    wantKeptnCredentialsNoBridgeURL,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var secret *v1.Secret
			if tt.secretName == "" {
				secret = &v1.Secret{}
			} else {
				secret = createTestSecret(tt.secretName, tt.secretData)
			}

			clientSet := fake.NewSimpleClientset(secret)

			for key, value := range tt.environmentVariables {
				os.Setenv(key, value)
			}
			defer func() {
				for k, _ := range tt.environmentVariables {
					os.Unsetenv(k)
				}
			}()

			secretReader := NewK8sSecretReader(clientSet)
			cm := NewKeptnCredentialsReader(secretReader)
			got, err := cm.GetKeptnCredentials(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
		})
	}
}

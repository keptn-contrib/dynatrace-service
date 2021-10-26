package credentials

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Test dynatrace credential behavior: values should be read from dynatrace secret unless secret name has been overridden in dynatrace config file.
// If neither is available, an error should be produced.
func TestCredentialManager_GetDynatraceCredentials(t *testing.T) {

	wantDynatraceCredentials, err := NewDynatraceCredentials("https://mySampleEnv.live.dynatrace.com", testDynatraceAPIToken)
	assert.NoError(t, err)

	dynatraceSecret := createTestSecret("dynatrace",
		map[string]string{
			"DT_TENANT":    "https://mySampleEnv.live.dynatrace.com",
			"DT_API_TOKEN": testDynatraceAPIToken,
		})
	dynatraceOtherSecret := createTestSecret("dynatrace_other",
		map[string]string{
			"DT_TENANT":    "https://mySampleEnv.live.dynatrace.com",
			"DT_API_TOKEN": testDynatraceAPIToken,
		})

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
			want:    wantDynatraceCredentials,
			wantErr: false,
		},
		{
			name:   "with dynatrace_other secret, with good config",
			secret: dynatraceOtherSecret,
			args: args{
				secretName: "dynatrace_other",
			},
			want:    wantDynatraceCredentials,
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
		// TODO: 2021-10-25: Improve tests to cover DynatraceCredentialsProviderFallbackDecorator
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			secretReader := NewK8sSecretReader(fake.NewSimpleClientset(tt.secret))
			cm := NewDynatraceK8sSecretReader(secretReader)
			decorator := NewCredentialManagerDefaultFallbackDecorator(cm)

			got, err := decorator.GetDynatraceCredentials(tt.args.secretName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
		})
	}
}

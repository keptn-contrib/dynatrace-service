package credentials

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Test dynatrace credential behavior: values should be read from dynatrace secret unless secret name has been overridden in dynatrace config file.
// If neither is available, an error should be produced.
func TestCredentialManager_GetDynatraceCredentials(t *testing.T) {

	wantDynatraceCredentials, err := NewDynatraceCredentials("https://mySampleEnv.live.dynatrace.com", testDynatraceAPIToken)
	assert.NoError(t, err)

	dynatraceSecret := createDynatraceDTSecret("dynatrace", "keptn", "https://mySampleEnv.live.dynatrace.com", testDynatraceAPIToken)
	dynatraceOtherSecret := createDynatraceDTSecret("dynatrace_other", "keptn", "https://mySampleEnv.live.dynatrace.com", testDynatraceAPIToken)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			secretReader := NewK8sSecretReader(fake.NewSimpleClientset(tt.secret))
			cm := NewDynatraceK8sSecretReader(secretReader)
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

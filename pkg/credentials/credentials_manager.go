package credentials

import (
	"errors"
	"os"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DTCredentials is a struct for the tenant and api token information
type DTCredentials struct {
	Tenant   string `json:"DT_TENANT" yaml:"DT_TENANT"`
	ApiToken string `json:"DT_API_TOKEN" yaml:"DT_API_TOKEN"`
}

var namespace = getPodNamespace()

func getPodNamespace() string {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return "keptn"
	}

	return ns
}

// GetDynatraceCredentials reads the Dynatrace credentials from the secret. Therefore, it first checks
// if a secret is specified in the dynatrace.conf.yaml and if not defaults to the secret "dynatrace"
func GetDynatraceCredentials(dynatraceConfig *config.DynatraceConfigFile) (*DTCredentials, error) {

	secretName := "dynatrace"
	if dynatraceConfig != nil && len(dynatraceConfig.DtCreds) > 0 {
		secretName = dynatraceConfig.DtCreds
	}

	if common.RunLocal || common.RunLocalTest {
		dtCreds := &DTCredentials{}

		dtCreds.Tenant = os.Getenv("DT_TENANT")
		dtCreds.ApiToken = os.Getenv("DT_API_TOKEN")
		return dtCreds, nil
	}

	kubeAPI, err := common.GetKubernetesClient()
	if err != nil {
		return nil, err
	}
	secret, err := kubeAPI.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if string(secret.Data["DT_TENANT"]) == "" || string(secret.Data["DT_API_TOKEN"]) == "" {
		return nil, errors.New("invalid or no Dynatrace credentials found. Requires at least DT_TENANT and DT_API_TOKEN in secret!")
	}

	dtCreds := &DTCredentials{}

	dtCreds.Tenant = strings.Trim(string(secret.Data["DT_TENANT"]), "\n")
	// remove trailing slash since this causes errors with the API calls
	dtCreds.Tenant = strings.TrimSuffix(dtCreds.Tenant, "/")
	dtCreds.ApiToken = strings.Trim(string(secret.Data["DT_API_TOKEN"]), "\n")

	return dtCreds, nil
}

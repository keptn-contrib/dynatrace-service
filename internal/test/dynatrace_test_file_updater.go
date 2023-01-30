package test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/rest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

// testDynatraceCredentialsFilenameEnvironmentVariable is an environment variable specifying the name of a file containing credentials (tenant and API token) for requesting test data.
const testDynatraceCredentialsFilenameEnvironmentVariable = "TEST_DYNATRACE_CREDENTIALS_FILENAME"

type dynatraceCredentials struct {
	Tenant   string `yaml:"tenant"`
	APIToken string `yaml:"apiToken"`
}

// dynatraceTestFileUpdater makes GET requests to a Dynatrace tenant to update test data files.
type dynatraceTestFileUpdater struct {
	t      *testing.T
	client rest.Client
}

// tryCreateDynatraceTestFileUpdater creates a dynatraceTestFileUpdater instance if TEST_DYNATRACE_CREDENTIALS_FILENAME is set.
// The test is failed if the credentials file is invalid.
func tryCreateDynatraceTestFileUpdater(t *testing.T) *dynatraceTestFileUpdater {
	credentialsFilename := os.Getenv(testDynatraceCredentialsFilenameEnvironmentVariable)
	if credentialsFilename == "" {
		logrus.Infof("Environment variable '%s' is empty, dynatraceTestFileUpdater will not be created", testDynatraceCredentialsFilenameEnvironmentVariable)
		return nil
	}

	logrus.Infof("Creating dynatraceTestFileUpdater using credentials from '%s'", credentialsFilename)

	creds, err := readTestDynatraceCredentials(credentialsFilename)
	if !assert.NoError(t, err, "Could not read test Dynatrace credentials") {
		return nil
	}

	return &dynatraceTestFileUpdater{
		t:      t,
		client: *rest.NewClient(createClient(), creds.Tenant, createAdditionalHeaders(creds.APIToken)),
	}
}

func readTestDynatraceCredentials(filename string) (*dynatraceCredentials, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read test Dynatrace credentials: %w", err)
	}

	creds := &dynatraceCredentials{}
	err = yaml.Unmarshal(yamlFile, creds)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal test Dynatrace credentials: %w", err)
	}

	if creds.Tenant == "" {
		return nil, errors.New("test Dynatrace credentials 'tenant' should not be empty")
	}

	if creds.APIToken == "" {
		return nil, errors.New("test Dynatrace credentials 'apiToken' should not be empty")
	}

	return creds, nil
}

func (h *dynatraceTestFileUpdater) tryUpdateTestFileUsingGet(url string, filename string) {
	if !shouldUpdateFileForURL(url) {
		return
	}

	response, status, _, err := h.client.Get(context.TODO(), url)
	if !assert.NoError(h.t, err, "REST client get should not produce an error") ||
		!assert.EqualValues(h.t, 200, status, fmt.Sprintf("HTTP response status code should be 200 but was %d with: %s", status, string(response))) {
		return
	}

	err = os.WriteFile(filename, response, 0666)
	assert.NoError(h.t, err)
}

func shouldUpdateFileForURL(url string) bool {
	return strings.HasPrefix(url, "/api/v2/metrics") ||
		strings.HasPrefix(url, "/api/v2/slo") ||
		strings.HasPrefix(url, "/api/v2/problems") ||
		strings.HasPrefix(url, "/api/v1/userSessionQueryLanguage/table") ||
		strings.HasPrefix(url, "/api/v2/securityProblems")
}

func createClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func createAdditionalHeaders(apiToken string) rest.HTTPHeader {
	header := rest.HTTPHeader{}
	header.Add("Authorization", "Api-Token "+apiToken)
	return header
}

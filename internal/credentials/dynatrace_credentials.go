package credentials

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/url"
)

var dynatraceAPITokenRegex = regexp.MustCompile(`^([^\.]+)\.([A-Z0-9]{24})\.([A-Z0-9]{64})$`)

type DynatraceCredentials struct {
	tenant   string
	apiToken string
}

func NewDynatraceCredentials(tenant string, apiToken string) (*DynatraceCredentials, error) {
	tenant, err := url.CleanURL(tenant)
	if err != nil {
		return nil, fmt.Errorf("cannot create Dynatrace credentials: %v", err)
	}

	apiToken, err = cleanDynatraceAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("cannot create Dynatrace credentials: %v", err)
	}

	return &DynatraceCredentials{tenant: tenant, apiToken: apiToken}, nil
}

// GetTenant gets the base URL of Dynatrace tenant. This is always prefixed with "https://" or "http://".
func (c *DynatraceCredentials) GetTenant() string {
	return c.tenant
}

func (c *DynatraceCredentials) GetAPIToken() string {
	return c.apiToken
}

func cleanDynatraceAPIToken(t string) (string, error) {
	t = strings.TrimSpace(t)

	chunks := dynatraceAPITokenRegex.FindStringSubmatch(t)
	if len(chunks) != 4 {
		return "", fmt.Errorf("invalid Dynatrace token")
	}

	return t, nil
}

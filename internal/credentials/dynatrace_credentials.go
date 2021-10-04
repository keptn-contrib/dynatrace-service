package credentials

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/url"
)

type DynatraceCredentials struct {
	tenant   string
	apiToken DynatraceAPIToken
}

func NewDynatraceCredentials(tenant string, apiTokenString string) (*DynatraceCredentials, error) {
	tenant, err := url.MakeCleanURL(tenant)
	if err != nil {
		return nil, fmt.Errorf("cannot create dynatrace credentials: %v", err)
	}

	apiToken, err := NewDynatraceAPIToken(apiTokenString)
	if err != nil {
		return nil, fmt.Errorf("cannot create dynatrace credentials: %v", err)
	}

	return &DynatraceCredentials{tenant: tenant, apiToken: *apiToken}, nil
}

// GetTenant gets the base URL of Dynatrace tenant. This is always prefixed with "https://" or "http://".
func (c DynatraceCredentials) GetTenant() string {
	return c.tenant
}

func (c DynatraceCredentials) GetAPIToken() DynatraceAPIToken {
	return c.apiToken
}

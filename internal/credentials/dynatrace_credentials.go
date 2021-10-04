package credentials

type DynatraceCredentials struct {
	tenant   string
	apiToken string
}

func NewDynatraceCredentials(tenant string, apiToken string) (*DynatraceCredentials, error) {
	return &DynatraceCredentials{tenant: tenant, apiToken: apiToken}, nil
}

// GetTenant gets the base URL of Dynatrace tenant. This is always prefixed with "https://" or "http://".
func (c DynatraceCredentials) GetTenant() string {
	return c.tenant
}

func (c DynatraceCredentials) GetAPIToken() string {
	return c.apiToken
}

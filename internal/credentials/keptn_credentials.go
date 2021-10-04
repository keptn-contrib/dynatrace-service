package credentials

type KeptnCredentials struct {
	apiURL   string
	apiToken string
}

func NewKeptnCredentials(apiURL string, apiToken string) (*KeptnCredentials, error) {
	return &KeptnCredentials{apiURL: apiURL, apiToken: apiToken}, nil
}

func (c KeptnCredentials) GetAPIURL() string {
	return c.apiURL
}

func (c KeptnCredentials) GetAPIToken() string {
	return c.apiToken
}

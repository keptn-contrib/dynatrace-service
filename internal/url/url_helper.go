package url

import (
	"errors"
	"net/url"
	"strings"
)

// CleanURL cleans and validates the URL. It trims whitespace and trailing slashes, and defaults to HTTPS if HTTP not specified.
func CleanURL(u string) (string, error) {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")

	uu, err := url.ParseRequestURI(u)
	if err != nil {
		// try again, assuming HTTPS scheme
		uu, err = url.ParseRequestURI("https://" + u)
		if err != nil {
			return "", err
		}
	}

	if uu.Scheme != "https" && uu.Scheme != "http" {
		return "", errors.New("expected https or http scheme")
	}

	if uu.Host == "" {
		return "", errors.New("URL must have a host")
	}

	return uu.String(), nil
}

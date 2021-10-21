package credentials

import (
	"errors"
	"regexp"
	"strings"
)

// DynatraceAPIToken represents a Dynatrace API token
type DynatraceAPIToken struct {
	token string
}

var tokenRegex = regexp.MustCompile(`^([^\.]+)\.([A-Z0-9]{24})\.([A-Z0-9]{64})$`)

// NewDynatraceAPIToken creates a new DynatraceAPIToken after validating the provided string
func NewDynatraceAPIToken(t string) (*DynatraceAPIToken, error) {
	t = strings.TrimSpace(t)

	chunks := tokenRegex.FindStringSubmatch(t)
	if len(chunks) != 4 {
		return nil, errors.New("Dynatrace token must consist of 3 components")
	}

	return &DynatraceAPIToken{token: t}, nil
}

func (t DynatraceAPIToken) String() string {
	return t.token
}

package credentials

import (
	"errors"
	"regexp"
	"strings"
)

type DynatraceAPIToken struct {
	token string
}

var publicComponentRegex = regexp.MustCompile("^[A-Z0-9]{24}$")
var secretComponentRegex = regexp.MustCompile("^[A-Z0-9]{64}$")

func NewDynatraceAPIToken(t string) (*DynatraceAPIToken, error) {
	if t == "" {
		return nil, errors.New("dynatrace token cannot be empty")
	}

	components := strings.Split(t, ".")
	if len(components) != 3 {
		return nil, errors.New("dynatrace token must consist of 3 components")
	}

	if !publicComponentRegex.MatchString(components[1]) {
		return nil, errors.New("public dynatrace token component must consist of 24 alpha-numeric characters")
	}

	if !secretComponentRegex.MatchString(components[2]) {
		return nil, errors.New("secret dynatrace token component must consist of 64 alpha-numeric characters")
	}

	return &DynatraceAPIToken{token: t}, nil
}

func (t DynatraceAPIToken) String() string {
	return t.token
}

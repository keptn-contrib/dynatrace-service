package dynatrace

import (
	"errors"
	"strings"
)

func CheckForUnexpectedHTMLResponseError(err error) error {
	// TODO 2021-08-19: should this be needed elsewhere as well?

	// in some cases, e.g. when the DT API has a problem, or the request URL is malformed, we do get a 200 response coded, but with an HTML error page instead of JSON
	// this function checks for the resulting error in that case and generates an error message that is more user friendly
	if strings.Contains(err.Error(), "invalid character '<'") {
		err = errors.New("received invalid response from Dynatrace API")
	}
	return err
}

package event

import (
	"net/url"
)

// GetEventSource gets the source to be used for CloudEvents originating from the dynatrace-service
func GetEventSource() string {
	source, _ := url.Parse("dynatrace-service")
	return source.String()
}

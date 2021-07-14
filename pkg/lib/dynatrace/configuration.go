package dynatrace

import (
	"os"
	"strconv"
)

// IsHttpSSLVerificationEnabled returns whether the SSL verification is enabled or disabled
func IsHttpSSLVerificationEnabled() bool {
	return readEnvAsBool("HTTP_SSL_VERIFY", true)
}

func readEnvAsBool(env string, fallbackValue bool) bool {
	if b, err := strconv.ParseBool(os.Getenv(env)); err == nil {
		return b
	}
	return fallbackValue
}

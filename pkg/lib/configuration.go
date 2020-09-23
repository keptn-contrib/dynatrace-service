package lib

import (
	"os"
	"strconv"
)

func GetTaggingRulesConfig() bool {
	return readEnvAsBool("GENERATE_TAGGING_RULES", false)
}

func GetProblemNotificationsConfig() bool {
	return readEnvAsBool("GENERATE_PROBLEM_NOTIFICATIONS", false)
}

func GetManagementZonesConfig() bool {
	return readEnvAsBool("GENERATE_MANAGEMENT_ZONES", false)
}

func GetGenerateDashboardsConfig() bool {
	return readEnvAsBool("GENERATE_DASHBOARDS", false)
}

func GetMetricEventsConfig() bool {
	return readEnvAsBool("GENERATE_METRIC_EVENTS", false)
}

func IsHttpSSLVerificationEnabled() bool {
	return readEnvAsBool("HTTP_SSL_VERIFY", true)
}

func readEnvAsBool(env string, fallbackValue bool) bool {
	if b, err := strconv.ParseBool(os.Getenv(env)); err == nil {
		return b
	}
	return fallbackValue
}

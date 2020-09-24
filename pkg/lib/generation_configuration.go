package lib

import (
	"os"
	"strconv"
)

func GetTaggingRulesConfig() bool {
	return readEnvAsBool("GENERATE_TAGGING_RULES")
}

func GetProblemNotificationsConfig() bool {
	return readEnvAsBool("GENERATE_PROBLEM_NOTIFICATIONS")
}

func GetManagementZonesConfig() bool {
	return readEnvAsBool("GENERATE_MANAGEMENT_ZONES")
}

func GetGenerateDashboardsConfig() bool {
	return readEnvAsBool("GENERATE_DASHBOARDS")
}

func GetMetricEventsConfig() bool {
	return readEnvAsBool("GENERATE_METRIC_EVENTS")
}

func GetServiceSyncConfig() bool {
	return readEnvAsBool("SYNCHRONIZE_DYNATRACE_SERVICES")
}

func readEnvAsBool(env string) bool {
	if b, err := strconv.ParseBool(os.Getenv(env)); err == nil {
		return b
	}
	return false
}

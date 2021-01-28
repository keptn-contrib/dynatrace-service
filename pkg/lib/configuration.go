package lib

import (
	"os"
	"strconv"
)

// IsTaggingRulesGenerationEnabled returns whether tagging rules should be generated when configuring the monitoring
func IsTaggingRulesGenerationEnabled() bool {
	return readEnvAsBool("GENERATE_TAGGING_RULES", false)
}

// IsProblemNotificationsGenerationEnabled returns whether problem notifications should be generated when configuring the monitoring
func IsProblemNotificationsGenerationEnabled() bool {
	return readEnvAsBool("GENERATE_PROBLEM_NOTIFICATIONS", false)
}

// IsManagementZonesGenerationEnabled returns whether management zones should be generated when configuring the monitoring
func IsManagementZonesGenerationEnabled() bool {
	return readEnvAsBool("GENERATE_MANAGEMENT_ZONES", false)
}

// IsDashboardsGenerationEnabled returns whether dashboards should be generated when configuring the monitoring
func IsDashboardsGenerationEnabled() bool {
	return readEnvAsBool("GENERATE_DASHBOARDS", false)
}

// IsMetricEventsGenerationEnabled returns whether metric events should be generated when configuring the monitoring
func IsMetricEventsGenerationEnabled() bool {
	return readEnvAsBool("GENERATE_METRIC_EVENTS", false)
}

// IsHttpSSLVerificationEnabled returns whether the SSL verification is enabled or disabled
func IsHttpSSLVerificationEnabled() bool {
	return readEnvAsBool("HTTP_SSL_VERIFY", true)
}

// IsServiceSyncEnabled returns wether the service synchronization is enabled or disabled
func IsServiceSyncEnabled() bool {
	return readEnvAsBool("SYNCHRONIZE_DYNATRACE_SERVICES", false)
}

// IsDashboardEnvironmentPublic returns wether the dashboard is environment wide public or not
func IsDashboardEnvironmentPublic() bool {
	return readEnvAsBool("DASHBOARD_ENVIRONMENT_PUBLIC", true)
}

func readEnvAsBool(env string, fallbackValue bool) bool {
	if b, err := strconv.ParseBool(os.Getenv(env)); err == nil {
		return b
	}
	return fallbackValue
}

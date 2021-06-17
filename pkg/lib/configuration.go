package lib

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
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

// GetServiceSyncInterval returns the number of seconds the service synchronizer should sleep between synchronization runs
// if the environment variable is empty or cannot be parsed, a default sync interval is used
func GetServiceSyncInterval() int {
	return readEnvAsInt("SYNCHRONIZE_DYNATRACE_SERVICES_INTERVAL_SECONDS", 300)
}

func readEnvAsBool(env string, defaultValue bool) bool {
	envValue := os.Getenv(env)
	if envValue == "" {
		log.WithFields(
			log.Fields{
				"name":    env,
				"default": defaultValue,
			}).Info("Environment variable not set or empty. Using default value.")
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(envValue)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"name":    env,
				"value":   envValue,
				"default": defaultValue,
			}).Error("Unable to parse environment variable. Using default value.")
		return defaultValue
	}

	return boolValue
}

func readEnvAsInt(env string, defaultValue int) int {
	envValue := os.Getenv(env)
	if envValue == "" {
		log.WithFields(
			log.Fields{
				"name":    env,
				"default": defaultValue,
			}).Info("Environment variable not set or empty. Using default value.")
		return defaultValue
	}

	parseInt, err := strconv.ParseInt(envValue, 10, 32)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"name":    env,
				"value":   envValue,
				"default": defaultValue,
			}).Error("Unable to parse environment variable. Using default value.")
		return defaultValue
	}

	return int(parseInt)
}

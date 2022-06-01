package env

import (
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const logLevelEnvironmentVariable = "LOG_LEVEL_DYNATRACE_SERVICE"

// GetAPIService gets the API_SERVICE environment variable.
func GetAPIService() string {
	return os.Getenv("API_SERVICE")
}

// GetShipyardController gets the SHIPYARD_CONTROLLER environment variable.
func GetShipyardController() string {
	return os.Getenv("SHIPYARD_CONTROLLER")
}

// GetConfigurationService gets the CONFIGURATION_SERVICE environment variable.
func GetConfigurationService() string {
	return os.Getenv("CONFIGURATION_SERVICE")
}

// GetDatastore gets the DATASTORE environment variable.
func GetDatastore() string {
	return os.Getenv("DATASTORE")
}

// GetVersion gets the version environment variable.
func GetVersion() string {
	return os.Getenv("version")
}

// GetPodNamespace gets the POD_NAMESPACE environment variable with the default of "keptn".
func GetPodNamespace() string {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return "keptn"
	}
	return ns
}

// GetKubernetesServiceHost gets the KUBERNETES_SERVICE_HOST environment variable
func GetKubernetesServiceHost() string {
	return os.Getenv("KUBERNETES_SERVICE_HOST")
}

// GetWorkGracePeriod returns the expected work period in which an event should be processed.
// This period is only enforced during a graceful shutdown.
// If not set, 20 seconds is assumed.
func GetWorkGracePeriod() time.Duration {
	return time.Duration(readEnvAsInt("WORK_GRACE_PERIOD_SECONDS", 20)) * time.Second
}

// GetReplyGracePeriod returns the expected period in which an event should reply.
// This period is only enforced during a graceful shutdown.
// If not set, 5 seconds is assumed.
func GetReplyGracePeriod() time.Duration {
	return time.Duration(readEnvAsInt("REPLY_GRACE_PERIOD_SECONDS", 5)) * time.Second
}

// GetLogLevel gets the log level specified by the LOG_LEVEL_DYNATRACE_SERVICE environment variable.
// If none is specified, log.InfoLevel is assumed.
func GetLogLevel() log.Level {
	level, err := log.ParseLevel(os.Getenv(logLevelEnvironmentVariable))
	if err != nil {
		log.WithError(err).Error("Couldn't parse " + logLevelEnvironmentVariable + " environment variable")
		return log.InfoLevel
	}

	return level
}

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

// GetServiceSyncInterval returns the number of seconds the service synchronizer should sleep between synchronization runs.
// If the environment variable is empty or cannot be parsed, a default sync interval is used.
func GetServiceSyncInterval() int {
	return readEnvAsInt("SYNCHRONIZE_DYNATRACE_SERVICES_INTERVAL_SECONDS", 60)
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

package monitoring

import (
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

type KeptnAPIConnectionCheck struct {
	APIURL               string
	ConnectionSuccessful bool
	Message              string
}

type ConfigureMonitoringEventHandler struct {
	event          ConfigureMonitoringAdapterInterface
	dtClient       dynatrace.ClientInterface
	kClient        keptn.ClientInterface
	resourceClient keptn.ResourceClientInterface
	serviceClient  keptn.ServiceClientInterface
}

// NewConfigureMonitoringEventHandler returns a new ConfigureMonitoringEventHandler
func NewConfigureMonitoringEventHandler(event ConfigureMonitoringAdapterInterface, dtClient dynatrace.ClientInterface, kClient keptn.ClientInterface, resourceClient keptn.ResourceClientInterface, serviceClient keptn.ServiceClientInterface) ConfigureMonitoringEventHandler {
	return ConfigureMonitoringEventHandler{
		event:          event,
		dtClient:       dtClient,
		kClient:        kClient,
		resourceClient: resourceClient,
		serviceClient:  serviceClient,
	}
}

func (eh ConfigureMonitoringEventHandler) HandleEvent() error {
	err := eh.configureMonitoring()
	if err != nil {
		log.WithError(err).Error("Configure monitoring failed")
	}
	return nil
}

func (eh *ConfigureMonitoringEventHandler) configureMonitoring() error {
	log.Info("Configuring Dynatrace monitoring")
	if eh.event.IsNotForDynatrace() {
		return nil
	}

	keptnAPICheck := &KeptnAPIConnectionCheck{}
	// check the connection to the Keptn API
	keptnCredentials, err := credentials.GetKeptnCredentials()
	if err != nil {
		log.WithError(err).Error("Failed to get Keptn API credentials")
		keptnAPICheck.Message = "Failed to get Keptn API Credentials"
		keptnAPICheck.ConnectionSuccessful = false
		keptnAPICheck.APIURL = "unknown"
	} else {
		keptnAPICheck.APIURL = keptnCredentials.GetAPIURL()
		log.WithField("apiURL", keptnCredentials.GetAPIURL()).Print("Verifying access to Keptn API")

		err = credentials.CheckKeptnConnection(keptnCredentials)
		if err != nil {
			keptnAPICheck.ConnectionSuccessful = false
			keptnAPICheck.Message = "Warning: Keptn API connection cannot be verified. This might be due to a no-loopback policy of your LoadBalancer. The endpoint might still be reachable from outside the cluster."
			log.WithError(err).Warn(keptnAPICheck.Message)
		} else {
			keptnAPICheck.ConnectionSuccessful = true
		}
	}

	var shipyard *keptnv2.Shipyard
	if eh.event.GetProject() != "" {
		shipyard, err = eh.kClient.GetShipyard()
		if err != nil {
			return eh.handleError(err)
		}
	}

	cfg := NewConfiguration(eh.dtClient, eh.kClient, eh.resourceClient, eh.serviceClient)

	configuredEntities, err := cfg.ConfigureMonitoring(eh.event.GetProject(), shipyard)
	if err != nil {
		return eh.handleError(err)
	}

	log.Info("Dynatrace Monitoring setup done")
	return eh.handleSuccess(getConfigureMonitoringResultMessage(keptnAPICheck, configuredEntities))
}

func getConfigureMonitoringResultMessage(apiCheck *KeptnAPIConnectionCheck, entities *ConfiguredEntities) string {
	if entities == nil {
		return ""
	}
	msg := "Dynatrace monitoring setup done.\nThe following entities have been configured:\n\n"

	if entities.ManagementZonesEnabled && len(entities.ManagementZones) > 0 {
		msg = msg + "---Management Zones:--- \n"
		for _, mz := range entities.ManagementZones {
			if mz.Success {
				msg = msg + "  - " + mz.Name + ": Created successfully \n"
			} else {
				msg = msg + "  - " + mz.Name + ": Error: " + mz.Message + "\n"
			}
		}
		msg = msg + "\n\n"
	}

	if entities.TaggingRulesEnabled && len(entities.TaggingRules) > 0 {
		msg = msg + "---Automatic Tagging Rules:--- \n"
		for _, mz := range entities.TaggingRules {
			if mz.Success {
				msg = msg + "  - " + mz.Name + ": Created successfully \n"
			} else {
				msg = msg + "  - " + mz.Name + ": Error: " + mz.Message + "\n"
			}
		}
		msg = msg + "\n\n"
	}

	if entities.ProblemNotificationsEnabled {
		msg = msg + "---Problem Notification:--- \n"
		msg = msg + "  - " + entities.ProblemNotifications.Message
		msg = msg + "\n\n"
	}

	if entities.MetricEventsEnabled && len(entities.MetricEvents) > 0 {
		msg = msg + "---Metric Events:--- \n"
		for _, mz := range entities.MetricEvents {
			if mz.Success {
				msg = msg + "  - " + mz.Name + ": Created successfully \n"
			} else {
				msg = msg + "  - " + mz.Name + ": Error: " + mz.Message + "\n"
			}
		}
		msg = msg + "\n\n"
	}

	if entities.DashboardEnabled && entities.Dashboard.Message != "" {
		msg = msg + "---Dashboard:--- \n"
		msg = msg + "  - " + entities.Dashboard.Message
		msg = msg + "\n\n"
	}

	if apiCheck != nil {
		msg = msg + "---Keptn API Connection Check:--- \n"
		msg = msg + "  - Keptn API URL: " + apiCheck.APIURL + "\n"
		msg = msg + fmt.Sprintf("  - Connection Successful: %v. %s\n", apiCheck.ConnectionSuccessful, apiCheck.Message)
		msg = msg + "\n"
	}

	return msg
}

func (eh *ConfigureMonitoringEventHandler) handleError(err error) error {
	log.Error(err)
	return eh.sendConfigureMonitoringFinishedEvent(NewFailureEventFactory(eh.event, err.Error()))
}

func (eh *ConfigureMonitoringEventHandler) handleSuccess(message string) error {
	return eh.sendConfigureMonitoringFinishedEvent(NewSuccessEventFactory(eh.event, message))
}

func (eh *ConfigureMonitoringEventHandler) sendConfigureMonitoringFinishedEvent(factory adapter.CloudEventFactoryInterface) error {
	if err := eh.kClient.SendCloudEvent(factory); err != nil {
		log.WithError(err).Error("Failed to send configure monitoring finished event")
		return err
	}

	return nil
}

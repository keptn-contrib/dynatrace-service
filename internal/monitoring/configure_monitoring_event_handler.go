package monitoring

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"

	log "github.com/sirupsen/logrus"
)

type keptnCredentialsCheckResult struct {
	APIURL  string
	Success bool
	Message string
}

type ConfigureMonitoringEventHandler struct {
	event              ConfigureMonitoringAdapterInterface
	dtClient           dynatrace.ClientInterface
	eventSenderClient  keptn.EventSenderClientInterface
	shipyardReader     keptn.ShipyardReaderInterface
	sliAndSLOReader    keptn.SLIAndSLOReaderInterface
	serviceClient      keptn.ServiceClientInterface
	credentialsChecker keptn.CredentialsCheckerInterface
}

// NewConfigureMonitoringEventHandler returns a new ConfigureMonitoringEventHandler
func NewConfigureMonitoringEventHandler(event ConfigureMonitoringAdapterInterface, dtClient dynatrace.ClientInterface, eventSenderClient keptn.EventSenderClientInterface, shipyardReader keptn.ShipyardReaderInterface, sliAndSLOReader keptn.SLIAndSLOReaderInterface, serviceClient keptn.ServiceClientInterface, credentialsChecker keptn.CredentialsCheckerInterface) ConfigureMonitoringEventHandler {
	return ConfigureMonitoringEventHandler{
		event:              event,
		dtClient:           dtClient,
		eventSenderClient:  eventSenderClient,
		shipyardReader:     shipyardReader,
		sliAndSLOReader:    sliAndSLOReader,
		serviceClient:      serviceClient,
		credentialsChecker: credentialsChecker,
	}
}

// HandleEvent handles a configure monitoring event.
func (eh ConfigureMonitoringEventHandler) HandleEvent(workCtx context.Context, replyCtx context.Context) error {
	err := eh.configureMonitoring(workCtx)
	if err != nil {
		log.WithError(err).Error("Configure monitoring failed")
	}
	return nil
}

func (eh *ConfigureMonitoringEventHandler) configureMonitoring(ctx context.Context) error {
	log.Info("Configuring Dynatrace monitoring")
	if eh.event.IsNotForDynatrace() {
		return nil
	}

	keptnCredentialsCheckResult := eh.checkKeptnCredentials(ctx)
	log.WithField("result", keptnCredentialsCheckResult).Info("Checked Keptn credentials")

	shipyard, err := eh.shipyardReader.GetShipyard(ctx, eh.event.GetProject())
	if err != nil {
		return eh.handleError(err)
	}

	cfg := newConfiguration(eh.dtClient, eh.eventSenderClient, eh.sliAndSLOReader, eh.serviceClient)

	configuredEntities, err := cfg.configureMonitoring(ctx, eh.event.GetProject(), *shipyard)
	if err != nil {
		return eh.handleError(err)
	}

	log.Info("Dynatrace Monitoring setup done")
	return eh.handleSuccess(getConfigureMonitoringResultMessage(keptnCredentialsCheckResult, configuredEntities))
}

func (eh *ConfigureMonitoringEventHandler) checkKeptnCredentials(ctx context.Context) keptnCredentialsCheckResult {
	keptnCredentials, err := credentials.GetKeptnCredentials(ctx)
	if err != nil {
		return keptnCredentialsCheckResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get Keptn API credentials: %s", err.Error()),
			APIURL:  "unknown",
		}
	}

	err = eh.credentialsChecker.CheckCredentials(*keptnCredentials)
	if err != nil {
		return keptnCredentialsCheckResult{
			Success: false,
			Message: fmt.Sprintf("Failed to verify to Keptn API credentials: %s", err.Error()),
			APIURL:  keptnCredentials.GetAPIURL(),
		}
	}

	return keptnCredentialsCheckResult{
		Success: true,
		APIURL:  keptnCredentials.GetAPIURL(),
	}

}

func getConfigureMonitoringResultMessage(keptnCredentialsCheckResult keptnCredentialsCheckResult, entities *configuredEntities) string {
	if entities == nil {
		return ""
	}
	msg := "Dynatrace monitoring setup done.\nThe following entities have been configured:\n\n"

	if len(entities.ManagementZones) > 0 {
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

	if len(entities.TaggingRules) > 0 {
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

	if entities.ProblemNotifications != nil {
		msg = msg + "---Problem Notification:--- \n"
		msg = msg + "  - " + entities.ProblemNotifications.Message
		msg = msg + "\n\n"
	}

	if len(entities.MetricEvents) > 0 {
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

	if entities.Dashboard != nil {
		msg = msg + "---Dashboard:--- \n"
		msg = msg + "  - " + entities.Dashboard.Message
		msg = msg + "\n\n"
	}

	msg = msg + "---Keptn API Connection Check:--- \n"
	msg = msg + "  - Keptn API URL: " + keptnCredentialsCheckResult.APIURL + "\n"
	msg = msg + fmt.Sprintf("  - Connection Successful: %v. %s\n", keptnCredentialsCheckResult.Success, keptnCredentialsCheckResult.Message)
	msg = msg + "\n"

	return msg
}

func (eh *ConfigureMonitoringEventHandler) handleError(err error) error {
	log.WithError(err).Error("Error handling configure monitoring event")
	return eh.sendConfigureMonitoringFinishedEvent(NewErroredConfigureMonitoringFinishedEventFactory(eh.event, err))
}

func (eh *ConfigureMonitoringEventHandler) handleSuccess(message string) error {
	return eh.sendConfigureMonitoringFinishedEvent(NewSucceededConfigureMonitoringFinishedEventFactory(eh.event, message))
}

func (eh *ConfigureMonitoringEventHandler) sendConfigureMonitoringFinishedEvent(factory adapter.CloudEventFactoryInterface) error {
	if err := eh.eventSenderClient.SendCloudEvent(factory); err != nil {
		log.WithError(err).Error("Failed to send configure monitoring finished event")
		return err
	}

	return nil
}

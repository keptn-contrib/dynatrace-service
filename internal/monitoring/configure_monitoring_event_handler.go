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
	apiURL  string
	success bool
	message string
}

type ConfigureMonitoringEventHandler struct {
	event              ConfigureMonitoringAdapterInterface
	dtClient           dynatrace.ClientInterface
	kClient            keptn.ClientInterface
	sloReader          keptn.SLOReaderInterface
	serviceClient      keptn.ServiceClientInterface
	credentialsChecker keptn.CredentialsCheckerInterface
}

// NewConfigureMonitoringEventHandler returns a new ConfigureMonitoringEventHandler
func NewConfigureMonitoringEventHandler(event ConfigureMonitoringAdapterInterface, dtClient dynatrace.ClientInterface, kClient keptn.ClientInterface, sloReader keptn.SLOReaderInterface, serviceClient keptn.ServiceClientInterface, credentialsChecker keptn.CredentialsCheckerInterface) ConfigureMonitoringEventHandler {
	return ConfigureMonitoringEventHandler{
		event:              event,
		dtClient:           dtClient,
		kClient:            kClient,
		sloReader:          sloReader,
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

	shipyard, err := eh.kClient.GetShipyard()
	if err != nil {
		return eh.handleError(err)
	}

	cfg := NewConfiguration(eh.dtClient, eh.kClient, eh.sloReader, eh.serviceClient)

	configuredEntities, err := cfg.ConfigureMonitoring(ctx, eh.event.GetProject(), *shipyard)
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
			success: false,
			message: fmt.Sprintf("Failed to get Keptn API credentials: %s", err.Error()),
			apiURL:  "unknown",
		}
	}

	err = eh.credentialsChecker.CheckCredentials(*keptnCredentials)
	if err != nil {
		return keptnCredentialsCheckResult{
			success: false,
			message: fmt.Sprintf("Failed to verify to Keptn API credentials: %s", err.Error()),
			apiURL:  keptnCredentials.GetAPIURL(),
		}
	}

	return keptnCredentialsCheckResult{
		success: true,
		apiURL:  keptnCredentials.GetAPIURL(),
	}

}

func getConfigureMonitoringResultMessage(keptnCredentialsCheckResult keptnCredentialsCheckResult, entities *ConfiguredEntities) string {
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
	msg = msg + "  - Keptn API URL: " + keptnCredentialsCheckResult.apiURL + "\n"
	msg = msg + fmt.Sprintf("  - Connection Successful: %v. %s\n", keptnCredentialsCheckResult.success, keptnCredentialsCheckResult.message)
	msg = msg + "\n"

	return msg
}

func (eh *ConfigureMonitoringEventHandler) handleError(err error) error {
	log.Error(err)
	return eh.sendConfigureMonitoringFinishedEvent(NewErroredConfigureMonitoringFinishedEventFactory(eh.event, err))
}

func (eh *ConfigureMonitoringEventHandler) handleSuccess(message string) error {
	return eh.sendConfigureMonitoringFinishedEvent(NewSucceededConfigureMonitoringFinishedEventFactory(eh.event, message))
}

func (eh *ConfigureMonitoringEventHandler) sendConfigureMonitoringFinishedEvent(factory adapter.CloudEventFactoryInterface) error {
	if err := eh.kClient.SendCloudEvent(factory); err != nil {
		log.WithError(err).Error("Failed to send configure monitoring finished event")
		return err
	}

	return nil
}

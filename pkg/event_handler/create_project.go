package event_handler

import (
	"encoding/base64"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/ghodss/yaml"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type CreateProjectEventHandler struct {
	Logger   keptn.LoggerInterface
	Event    cloudevents.Event
	DTHelper *lib.DynatraceHelper
}

func (eh CreateProjectEventHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	e := &keptn.ProjectCreateEventData{}
	err := eh.Event.DataAs(e)
	if err != nil {
		eh.Logger.Error("Could not parse event payload: " + err.Error())
		return err
	}

	shipyard := &keptn.Shipyard{}

	decodedShipyard, err := base64.StdEncoding.DecodeString(e.Shipyard)
	if err != nil {
		eh.Logger.Error("Could not decode shipyard: " + err.Error())
	}
	err = yaml.Unmarshal(decodedShipyard, shipyard)
	if err != nil {
		eh.Logger.Error("Could not parse shipyard: " + err.Error())
	}

	clientSet, err := common.GetKubernetesClient()
	if err != nil {
		eh.Logger.Error("could not create k8s client")
	}

	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{})
	if err != nil {
		eh.Logger.Error("could not create Keptn handler: " + err.Error())
	}

	dtHelper, err := lib.NewDynatraceHelper(keptnHandler)
	if err != nil {
		eh.Logger.Error("Could not create Dynatrace Helper: " + err.Error())
	}
	dtHelper.KubeApi = clientSet
	dtHelper.Logger = eh.Logger
	eh.DTHelper = dtHelper

	err = eh.DTHelper.EnsureDTTaggingRulesAreSetUp()
	if err != nil {
		eh.Logger.Error("Could not set up tagging rules: " + err.Error())
	}

	err = eh.DTHelper.EnsureProblemNotificationsAreSetUp()
	if err != nil {
		eh.Logger.Error("Could not set up problem notification: " + err.Error())
	}

	err = eh.DTHelper.CreateCalculatedMetrics(e.Project)
	if err != nil {
		eh.Logger.Error("Could not create calculated metrics: " + err.Error())
	}

	err = eh.DTHelper.CreateTestStepCalculatedMetrics(e.Project)
	if err != nil {
		eh.Logger.Error("Could not create calculated metrics: " + err.Error())
	}

	err = eh.DTHelper.CreateDashboard(e.Project, *shipyard, nil)
	if err != nil {
		eh.Logger.Error("Could not create Dynatrace dashboard for project " + e.Project + ": " + err.Error())
		// do not return because there are no dependencies to the dashboard
	}

	err = eh.DTHelper.CreateManagementZones(e.Project, *shipyard)
	if err != nil {
		eh.Logger.Error("Could not create Management Zones for project " + e.Project + ": " + err.Error())
		return err
	}

	return nil
}

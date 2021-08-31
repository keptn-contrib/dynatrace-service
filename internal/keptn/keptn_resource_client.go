package keptn

import (
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	api "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ResourceClientInterface interface {
	GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error)
	GetResource(event adapter.EventContentAdapter, resource string) (string, error)
}

type ResourceClient struct {
	handler *api.ResourceHandler
}

func NewConfigResourceClient() *ResourceClient {
	return &ResourceClient{
		handler: api.NewResourceHandler(
			common.GetConfigurationServiceURL()),
	}
}

func (rc *ResourceClient) GetSLOs(project string, stage string, service string) (*keptn.ServiceLevelObjectives, error) {
	resource, err := rc.handler.GetServiceResource(project, stage, service, "slo.yaml")
	if err != nil || resource.ResourceContent == "" {
		return nil, errors.New("No SLO file available for service " + service + " in stage " + stage)
	}

	slos := &keptn.ServiceLevelObjectives{}
	err = yaml.Unmarshal([]byte(resource.ResourceContent), slos)
	if err != nil {
		return nil, errors.New("invalid SLO file format")
	}

	return slos, nil
}

func (rc *ResourceClient) GetResource(event adapter.EventContentAdapter, resource string) (string, error) {

	if len(event.GetProject()) > 0 && len(event.GetStage()) > 0 && len(event.GetService()) > 0 {
		keptnResourceContent, err := rc.handler.GetServiceResource(event.GetProject(), event.GetStage(), event.GetService(), resource)
		if err == api.ResourceNotFoundError {
			log.WithFields(
				log.Fields{
					"project": event.GetProject(),
					"stage":   event.GetStage(),
					"service": event.GetService(),
				}).Debugf("%s not available for service", resource)
		} else if err != nil {
			return "", fmt.Errorf("failed to retrieve %s in project %s at stage %s for service %s: %v", resource, event.GetProject(), event.GetStage(), event.GetService(), err)
		} else {
			log.WithFields(
				log.Fields{
					"project": event.GetProject(),
					"stage":   event.GetStage(),
					"service": event.GetService(),
				}).Infof("Found %s for service", resource)
			return keptnResourceContent.ResourceContent, nil
		}
	}

	if len(event.GetProject()) > 0 && len(event.GetStage()) > 0 {
		keptnResourceContent, err := rc.handler.GetStageResource(event.GetProject(), event.GetStage(), resource)
		if err == api.ResourceNotFoundError {
			log.WithFields(
				log.Fields{
					"project": event.GetProject(),
					"stage":   event.GetStage(),
				}).Debugf("%s not available for stage", resource)
		} else if err != nil {
			return "", fmt.Errorf("failed to retrieve %s in project %s at stage %s: %v", resource, event.GetProject(), event.GetStage(), err)
		} else {
			log.WithFields(
				log.Fields{
					"project": event.GetProject(),
					"stage":   event.GetStage(),
				}).Infof("Found %s for stage", resource)
			return keptnResourceContent.ResourceContent, nil
		}
	}

	if len(event.GetProject()) > 0 {
		keptnResourceContent, err := rc.handler.GetProjectResource(event.GetProject(), resource)
		if err == api.ResourceNotFoundError {
			log.WithField("project", event.GetProject()).Debugf("%s not available for project", resource)
		} else if err != nil {
			return "", fmt.Errorf("failed to retrieve %s in project %s: %v", resource, event.GetProject(), err)
		} else {
			log.WithField("project", event.GetProject()).Infof("Found %s for project", resource)
			return keptnResourceContent.ResourceContent, nil
		}
	}

	log.Infof("%s not found", resource)
	return "", nil
}

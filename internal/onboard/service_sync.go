package onboard

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	apimodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

const defaultDTProjectName = "dynatrace"
const defaultDTProjectStage = "quality-gate"
const defaultSLOFile = `---
spec_version: "1.0"
comparison:
  aggregate_function: "avg"
  compare_with: "single_result"
  include_result_with_score: "pass"
  number_of_comparison_results: 1
filter:
objectives:
  - sli: "response_time_p95"
    key_sli: false
    pass:             
      - criteria:
          - "<600"    
    warning:        
      - criteria:
          - "<=800"
    weight: 1
  - sli: "error_rate"
    key_sli: false
    pass:
      - criteria:
          - "<5"
  - sli: throughput
total_score:
  pass: "90%"
  warning: "75%"`

const defaultSLIConfigFile = `---
spec_version: '1.0'
indicators:
  throughput: "metricSelector=builtin:service.requestCount.total:merge(0):sum&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  error_rate: "metricSelector=builtin:service.errors.total.rate:merge(0):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p50: "metricSelector=builtin:service.response.time:merge(0):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p90: "metricSelector=builtin:service.response.time:merge(0):percentile(90)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p95: "metricSelector=builtin:service.response.time:merge(0):percentile(95)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"`

var encodedDefaultSLOFile string

type initSyncEventAdapter struct {
}

func (initSyncEventAdapter) GetShKeptnContext() string {
	return ""
}

func (initSyncEventAdapter) GetEvent() string {
	return ""
}

func (initSyncEventAdapter) GetSource() string {
	return ""
}

func (initSyncEventAdapter) GetProject() string {
	return defaultDTProjectName
}

func (initSyncEventAdapter) GetStage() string {
	return defaultDTProjectStage
}

func (initSyncEventAdapter) GetService() string {
	return ""
}

func (initSyncEventAdapter) GetDeployment() string {
	return ""
}

func (initSyncEventAdapter) GetTestStrategy() string {
	return ""
}

func (initSyncEventAdapter) GetDeploymentStrategy() string {
	return ""
}

func (initSyncEventAdapter) GetImage() string {
	return ""
}

func (initSyncEventAdapter) GetTag() string {
	return ""
}

func (initSyncEventAdapter) GetLabels() map[string]string {
	return nil
}

type serviceSynchronizer struct {
	projectsAPI       *keptnapi.ProjectHandler
	servicesAPI       *keptnapi.ServiceHandler
	resourcesAPI      *keptnapi.ResourceHandler
	apiHandler        *keptnapi.APIHandler
	credentialManager credentials.CredentialManagerInterface
	EntitiesClient    *dynatrace.EntitiesClient
	syncTimer         *time.Ticker
	keptnHandler      *keptnv2.Keptn
	servicesInKeptn   []string
	dtConfigGetter    config.DynatraceConfigGetterInterface
}

var serviceSynchronizerInstance *serviceSynchronizer

const shipyardController = "SHIPYARD_CONTROLLER"
const defaultShipyardControllerURL = "http://shipyard-controller:8080"

// ActivateServiceSynchronizer godoc
func ActivateServiceSynchronizer(c credentials.CredentialManagerInterface) *serviceSynchronizer {
	if serviceSynchronizerInstance == nil {

		encodedDefaultSLOFile = b64.StdEncoding.EncodeToString([]byte(defaultSLOFile))
		serviceSynchronizerInstance = &serviceSynchronizer{
			credentialManager: c,
		}

		serviceSynchronizerInstance.dtConfigGetter = config.NewDynatraceConfigGetter(keptn.NewConfigResourceClient())
		serviceSynchronizerInstance.EntitiesClient = dynatrace.NewEntitiesClient(dynatrace.NewClient(nil))

		configServiceBaseURL := common.GetConfigurationServiceURL()
		shipyardControllerBaseURL := common.GetShipyardControllerURL()
		log.WithFields(
			log.Fields{
				"configServiceBaseURL":      configServiceBaseURL,
				"shipyardControllerBaseURL": shipyardControllerBaseURL,
			}).Debug("Initializing Service Synchronizer")

		serviceSynchronizerInstance.projectsAPI = keptnapi.NewProjectHandler(shipyardControllerBaseURL)
		serviceSynchronizerInstance.servicesAPI = keptnapi.NewServiceHandler(shipyardControllerBaseURL)
		serviceSynchronizerInstance.resourcesAPI = keptnapi.NewResourceHandler(configServiceBaseURL)

		serviceSynchronizerInstance.initializeSynchronizationTimer()

	}
	return serviceSynchronizerInstance
}

func (s *serviceSynchronizer) initializeSynchronizationTimer() {
	syncInterval := lib.GetServiceSyncInterval()
	log.WithField("syncInterval", syncInterval).Info("Service Synchronizer will sync periodically")
	s.syncTimer = time.NewTicker(time.Duration(syncInterval) * time.Second)
	go func() {
		for {
			s.synchronizeServices()
			<-s.syncTimer.C
			log.WithField("delaySeconds", syncInterval).Info("Synchronizing services")
		}
	}()
}

func (s *serviceSynchronizer) synchronizeServices() {
	if err := s.establishDTAPIConnection(); err != nil {
		log.WithError(err).Error("Could not establish Dynatrace API connection")
		return
	}

	log.WithField("project", defaultDTProjectName).Info("Fetching existing services in project")
	if err := s.fetchExistingServices(); err != nil {
		log.WithError(err).Error("Could not fetch existing services")
		return
	}

	log.Info("Fetching service entities with tags 'keptn_managed' and 'keptn_service'")

	entities, err := s.EntitiesClient.GetKeptnManagedServices()
	if err != nil {
		log.WithError(err).Error("Error fetching keptn managed services from dynatrace")
		return
	}

	for _, entity := range entities {
		s.synchronizeEntity(entity)
	}

}

func (s *serviceSynchronizer) synchronizeEntity(entity dynatrace.Entity) {
	log.WithField("entityId", entity.EntityID).Debug("Synchronizing entity")

	serviceName, err := getKeptnServiceName(entity)
	if err != nil {
		log.WithField("entityId", entity.EntityID).Debug("Skipping entity due to no valid service name")
		return
	}
	log.WithFields(
		log.Fields{
			"serviceName": serviceName,
			"entityId":    entity.EntityID,
		}).Debug("Got service name for entity")

	if doesServiceExist(s.servicesInKeptn, serviceName) {
		log.WithField("service", serviceName).Debug("Service already exists in project, skipping")
		return
	}

	if err := s.addServiceToKeptn(serviceName); err != nil {
		log.WithError(err).WithField("entityId", entity.EntityID).Error("Could not synchronize DT entity")
	}
}

func (s *serviceSynchronizer) establishDTAPIConnection() error {
	dynatraceConfig, err := s.dtConfigGetter.GetDynatraceConfig(initSyncEventAdapter{})
	if err != nil {
		return fmt.Errorf("failed to load Dynatrace config: %s", err.Error())
	}

	creds, err := s.credentialManager.GetDynatraceCredentials(dynatraceConfig.DtCreds)
	if err != nil {
		return fmt.Errorf("failed to load Dynatrace credentials: %s", err.Error())
	}

	s.EntitiesClient.Client.DynatraceCreds = creds
	return nil
}

func (s *serviceSynchronizer) fetchExistingServices() error {
	project, errObj := s.projectsAPI.GetProject(apimodels.Project{
		ProjectName: defaultDTProjectName,
	})
	if errObj != nil {
		if errObj.Code == 404 {
			return fmt.Errorf("project %s does not exist", defaultDTProjectName)
		}
		return fmt.Errorf("could not check if Keptn project %s exists: %s", defaultDTProjectName, *errObj.Message)
	}
	if project == nil {
		return fmt.Errorf("keptn project %s does not exist", defaultDTProjectName)
	}

	// get all services currently in the project
	s.servicesInKeptn = []string{}
	allKeptnServicesInProject, err := s.servicesAPI.GetAllServices(defaultDTProjectName, defaultDTProjectStage)
	if err != nil {
		return fmt.Errorf("could not fetch services of Keptn project %s: %s", defaultDTProjectName, err.Error())
	}
	for _, service := range allKeptnServicesInProject {
		s.servicesInKeptn = append(s.servicesInKeptn, service.ServiceName)
	}

	return nil
}

func getKeptnServiceName(entity dynatrace.Entity) (string, error) {
	if entity.Tags != nil {
		for _, tag := range entity.Tags {
			if tag.Key == "keptn_service" && tag.Value != "" {
				return tag.Value, nil
			}
		}
	}
	return "", fmt.Errorf("entity %v has no 'keptn_service' tag", entity.EntityID)
}

func doesServiceExist(services []string, serviceName string) bool {
	for _, service := range services {
		if service == serviceName {
			return true
		}
	}
	return false
}

func (s *serviceSynchronizer) addServiceToKeptn(serviceName string) error {
	_, err := s.createService(defaultDTProjectName, &apimodels.CreateService{
		ServiceName: &serviceName,
	})
	if err != nil {
		return fmt.Errorf("could not create service %s: %s", serviceName, err)
	}

	log.WithField("service", serviceName).Debug("Service is available. Proceeding with SLO upload.")

	if err := s.createSLOResource(serviceName); err == nil {
		log.WithField("service", serviceName).Info(fmt.Sprintf("Uploaded slo.yaml for service %s", serviceName))
	} else {
		log.WithField("service", serviceName).Info("Could not create SLO resource for service")
	}

	if err := s.createSLIResource(serviceName); err == nil {
		log.WithField("service", serviceName).Info("Uploaded sli.yaml for service")
	} else {
		log.WithField("service", serviceName).Info("Could not create SLI resource for service")
	}

	s.servicesInKeptn = append(s.servicesInKeptn, serviceName)
	return nil
}

func (s *serviceSynchronizer) createService(projectName string, service *apimodels.CreateService) (string, error) {
	bodyStr, err := json.Marshal(service)
	if err != nil {
		return "", fmt.Errorf("could not marshal service payload: %s", err.Error())
	}

	var scBaseURL string
	scEndpoint, err := keptncommon.GetServiceEndpoint(shipyardController)
	if err == nil {
		scBaseURL = scEndpoint.String()
	} else {
		scBaseURL = defaultShipyardControllerURL
	}
	req, err := http.NewRequest("POST", scBaseURL+"/v1/project/"+projectName+"/service", bytes.NewBuffer(bodyStr))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 204 {
		if len(body) > 0 {
			return string(body), nil
		}

		return "", nil
	}

	if len(body) > 0 {
		return "", errors.New(string(body))
	}

	return "", fmt.Errorf("received unexpected response: %d %s", resp.StatusCode, resp.Status)
}

func (s *serviceSynchronizer) createSLOResource(serviceName string) error {
	resourceURI := "slo.yaml"
	sloResource := &apimodels.Resource{
		ResourceContent: defaultSLOFile,
		ResourceURI:     &resourceURI,
	}
	_, err := s.resourcesAPI.CreateServiceResources(
		defaultDTProjectName,
		defaultDTProjectStage,
		serviceName,
		[]*apimodels.Resource{sloResource},
	)

	if err != nil {
		return fmt.Errorf("could not upload slo.yaml to service %s: %s", serviceName, err.Error())
	}

	return nil
}

func (s *serviceSynchronizer) createSLIResource(serviceName string) error {
	resourceURI := "dynatrace/sli.yaml"
	sliFileContent := strings.ReplaceAll(defaultSLIConfigFile, "$SERVICE", serviceName)

	sliResource := &apimodels.Resource{
		ResourceContent: sliFileContent,
		ResourceURI:     &resourceURI,
	}
	_, err := s.resourcesAPI.CreateServiceResources(
		defaultDTProjectName,
		defaultDTProjectStage,
		serviceName,
		[]*apimodels.Resource{sliResource},
	)

	if err != nil {
		return fmt.Errorf("could not upload sli.yaml to service %s: %s", serviceName, err.Error())
	}

	return nil
}

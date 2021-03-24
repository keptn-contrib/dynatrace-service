package lib

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/pkg/adapter"
	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	apimodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
	logger            keptncommon.LoggerInterface
	projectsAPI       *keptnapi.ProjectHandler
	servicesAPI       *keptnapi.ServiceHandler
	resourcesAPI      *keptnapi.ResourceHandler
	apiHandler        *keptnapi.APIHandler
	credentialManager credentials.CredentialManagerInterface
	DTHelper          *DynatraceHelper
	syncTimer         *time.Ticker
	keptnHandler      *keptnv2.Keptn
	servicesInKeptn   []string
	dtConfig          *config.DynatraceConfigFile
}

var serviceSynchronizerInstance *serviceSynchronizer

const shipyardController = "SHIPYARD_CONTROLLER"
const configurationService = "CONFIGURATION_SERVICE"
const defaultShipyardControllerURL = "http://shipyard-controller:8080"
const defaultConfigurationServiceURL = "http://configuration-service:8080"

// ActivateServiceSynchronizer godoc
func ActivateServiceSynchronizer(c *credentials.CredentialManager) *serviceSynchronizer {
	if serviceSynchronizerInstance == nil {

		encodedDefaultSLOFile = b64.StdEncoding.EncodeToString([]byte(defaultSLOFile))
		logger := keptncommon.NewLogger("", "", "dynatrace-service")
		serviceSynchronizerInstance = &serviceSynchronizer{
			logger:            logger,
			credentialManager: c,
		}

		dynatraceConfig, err := adapter.GetDynatraceConfig(initSyncEventAdapter{}, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to load Dynatrace config: %s", err.Error()))
			return nil
		}
		serviceSynchronizerInstance.dtConfig = dynatraceConfig

		serviceSynchronizerInstance.DTHelper = NewDynatraceHelper(nil, nil, logger)

		serviceSynchronizerInstance.logger.Debug("Initializing Service Synchronizer")
		configServiceBaseURL := common.GetConfigurationServiceURL()
		shipyardControllerBaseURL := common.GetShipyardControllerURL()

		serviceSynchronizerInstance.logger.Debug("Service Synchronizer uses configuration service URL: " + configServiceBaseURL)
		serviceSynchronizerInstance.logger.Debug("Service Synchronizer uses shipyard controller URL: " + shipyardControllerBaseURL)

		serviceSynchronizerInstance.projectsAPI = keptnapi.NewProjectHandler(shipyardControllerBaseURL)
		serviceSynchronizerInstance.servicesAPI = keptnapi.NewServiceHandler(shipyardControllerBaseURL)
		serviceSynchronizerInstance.resourcesAPI = keptnapi.NewResourceHandler(configServiceBaseURL)

		serviceSynchronizerInstance.initializeSynchronizationTimer()

	}
	return serviceSynchronizerInstance
}

func (s *serviceSynchronizer) initializeSynchronizationTimer() {
	var syncInterval int
	intervalEnv := os.Getenv("SYNCHRONIZE_DYNATRACE_SERVICES_INTERVAL_SECONDS")
	if intervalEnv == "" {
		syncInterval = 300
	}
	parseInt, err := strconv.ParseInt(intervalEnv, 10, 32)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Could not parse SYNCHRONIZE_DYNATRACE_SERVICES_INTERVAL_SECONDS with value %s, using 300s as default.", intervalEnv))
		syncInterval = 300
	}
	syncInterval = int(parseInt)
	s.logger.Info(fmt.Sprintf("Service Synchronizer will sync every %d seconds", syncInterval))
	s.syncTimer = time.NewTicker(time.Duration(syncInterval) * time.Second)
	go func() {
		for {
			<-s.syncTimer.C
			s.logger.Info(fmt.Sprintf("%d seconds have passed. Synchronizing services", syncInterval))
			s.synchronizeServices()
		}
	}()
	s.synchronizeServices()
}

func (s *serviceSynchronizer) establishDTAPIConnection() error {
	creds, err := s.credentialManager.GetDynatraceCredentials(s.dtConfig)
	if err != nil {
		return fmt.Errorf("failed to load Dynatrace credentials: %s", err.Error())
	}

	s.DTHelper.DynatraceCreds = creds
	return nil
}

func (s *serviceSynchronizer) synchronizeServices() {
	if err := s.establishDTAPIConnection(); err != nil {
		s.logger.Error(fmt.Sprintf("could not synchronize DT services: %s", err.Error()))
		return
	}
	s.logger.Info("checking if project " + defaultDTProjectName + " exists")
	project, errObj := s.projectsAPI.GetProject(apimodels.Project{
		ProjectName: defaultDTProjectName,
	})
	if errObj != nil {
		if errObj.Code == 404 {
			s.logger.Info("Project " + defaultDTProjectName + " does not exist. Stopping synchronization")
			return
		}
		s.logger.Error(fmt.Sprintf("Could not check if Keptn project %s exists: %s", defaultDTProjectName, *errObj.Message))
		return
	}
	if project == nil {
		s.logger.Info("Project " + defaultDTProjectName + " does not exist. Stopping synchronization")
		return
	}
	allKeptnServicesInProject, err := s.servicesAPI.GetAllServices(defaultDTProjectName, defaultDTProjectStage)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Could not fetch services of Keptn project %s: %s", defaultDTProjectName, err.Error()))
		return
	}
	s.servicesInKeptn = []string{}
	for _, service := range allKeptnServicesInProject {
		s.servicesInKeptn = append(s.servicesInKeptn, service.ServiceName)
	}
	s.checkForTaggedDynatraceServiceEntities()
}

func (s *serviceSynchronizer) checkForTaggedDynatraceServiceEntities() {
	s.logger.Info("fetching services with tags 'keptn_managed' and 'keptn_service'")

	nextPageKey := ""
	pageSize := 50

	for {
		entitiesResponse, err := s.fetchKeptnManagedServicesFromDynatrace(nextPageKey, pageSize)
		if err != nil {
			s.logger.Error(fmt.Sprintf("could not get keptn_managed services: %v", err))
			return
		}

		for _, entity := range entitiesResponse.Entities {
			s.logger.Debug("Synchronizing entity " + entity.EntityID)
			serviceName := getKeptnServiceNameOfEntity(entity)
			s.logger.Debug(fmt.Sprintf("Keptn Service name used for entity %s: %s", entity.EntityID, serviceName))
			err := s.synchronizeDTEntityWithKeptn(serviceName)
			if err != nil {
				s.logger.Error(fmt.Sprintf("could not synchronize DT entity with ID %s: %v", entity.EntityID, err))
			}
			s.servicesInKeptn = append(s.servicesInKeptn, serviceName)
		}

		if entitiesResponse.NextPageKey == "" {
			break
		}
		nextPageKey = entitiesResponse.NextPageKey

	}
	return

}

func (s *serviceSynchronizer) synchronizeDTEntityWithKeptn(serviceName string) error {
	s.logger.Debug(fmt.Sprintf("Checking if service %s already exists in Keptn project '%s'", serviceName, defaultDTProjectName))
	serviceExists := doesServiceExist(s.servicesInKeptn, serviceName)

	if serviceExists {
		s.logger.Debug(fmt.Sprintf("Service %s already exists in Keptn project '%s'", serviceName, defaultDTProjectName))
		return nil
	}

	s.logger.Debug(fmt.Sprintf("Service %s does not exist yet in Keptn project '%s'", serviceName, defaultDTProjectName))

	_, err := s.createService(defaultDTProjectName, &apimodels.CreateService{
		ServiceName: &serviceName,
	})
	if err != nil {
		return fmt.Errorf("could not create service %s: %s", serviceName, err)
	}

	s.logger.Error(fmt.Sprintf("Service %s is available. Proceeding with uploading SLO", serviceName))

	resourceURI := "slo.yaml"
	sloResource := &apimodels.Resource{
		ResourceContent: defaultSLOFile,
		ResourceURI:     &resourceURI,
	}
	_, err = s.resourcesAPI.CreateServiceResources(
		defaultDTProjectName,
		defaultDTProjectStage,
		serviceName,
		[]*apimodels.Resource{sloResource},
	)

	if err != nil {
		return fmt.Errorf("could not upload slo.yaml to service %s: %v", serviceName, err)
	}
	s.logger.Info(fmt.Sprintf("uploaded slo.yaml for service %s", serviceName))

	resourceURI = "dynatrace/sli.yaml"
	sliFileContent := strings.ReplaceAll(defaultSLIConfigFile, "$SERVICE", serviceName)

	sliResource := &apimodels.Resource{
		ResourceContent: sliFileContent,
		ResourceURI:     &resourceURI,
	}
	_, err = s.resourcesAPI.CreateServiceResources(
		defaultDTProjectName,
		defaultDTProjectStage,
		serviceName,
		[]*apimodels.Resource{sliResource},
	)

	if err != nil {
		return fmt.Errorf("could not upload sli.yaml to service %s: %v", serviceName, err)
	}
	s.logger.Info(fmt.Sprintf("uploaded sli.yaml for service %s", serviceName))

	return nil
}

func doesServiceExist(services []string, serviceName string) bool {
	for _, service := range services {
		if service == serviceName {

			return true
		}
	}
	return false
}

func getKeptnServiceNameOfEntity(entity entity) string {
	if entity.Tags != nil {
		for _, tag := range entity.Tags {
			if tag.Key == "keptn_service" && tag.Value != "" && keptncommon.ValidateKeptnEntityName(tag.Value) {
				return tag.Value

			}
		}
	}
	return entity.EntityID
}

func (s *serviceSynchronizer) fetchKeptnManagedServicesFromDynatrace(nextPageKey string, pageSize int) (*dtEntityListResponse, error) {
	var query string
	if nextPageKey == "" {
		query = "/api/v2/entities?entitySelector=type(\"SERVICE\")%20AND%20tag(\"keptn_managed\")%20AND%20tag(\"keptn_service\")&fields=+tags&pageSize=" + strconv.FormatInt(int64(pageSize), 10)
	} else {
		query = "/api/v2/entities?nextPageKey=" + nextPageKey
	}
	response, err := s.DTHelper.sendDynatraceAPIRequest(query, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch service entities with 'keptn_managed' and 'keptn_service' tags: %v", err)
	}

	dtEntities := &dtEntityListResponse{}
	err = json.Unmarshal([]byte(response), dtEntities)
	if err != nil {
		return nil, fmt.Errorf("could not decode response from Dynatrace API: %v", err)
	}
	return dtEntities, nil
}

func (s *serviceSynchronizer) createService(projectName string, service *apimodels.CreateService) (interface{}, interface{}) {
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

	return "", fmt.Errorf("Received unexptected response: %d %s", resp.StatusCode, resp.Status)
}

type dtEntityListResponse struct {
	TotalCount  int      `json:"totalCount"`
	PageSize    int      `json:"pageSize"`
	NextPageKey string   `json:"nextPageKey"`
	Entities    []entity `json:"entities"`
}
type tags struct {
	Context              string `json:"context"`
	Key                  string `json:"key"`
	StringRepresentation string `json:"stringRepresentation"`
	Value                string `json:"value,omitempty"`
}
type entity struct {
	EntityID    string `json:"entityId"`
	DisplayName string `json:"displayName"`
	Tags        []tags `json:"tags"`
}

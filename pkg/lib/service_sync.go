package lib

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	"github.com/keptn-contrib/dynatrace-service/pkg/config"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	apimodels "github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
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
	logger          keptn.LoggerInterface
	projectsAPI     *keptnapi.ProjectHandler
	servicesAPI     *keptnapi.ServiceHandler
	resourcesAPI    *keptnapi.ResourceHandler
	apiMutex        sync.Mutex
	DTHelper        *DynatraceHelper
	syncTimer       *time.Ticker
	keptnHandler    *keptn.Keptn
	servicesInKeptn []string
}

var serviceSynchronizerInstance *serviceSynchronizer

// ActivateServiceSynchronizer godoc
func ActivateServiceSynchronizer() *serviceSynchronizer {
	if serviceSynchronizerInstance == nil {

		encodedDefaultSLOFile = b64.StdEncoding.EncodeToString([]byte(defaultSLOFile))
		logger := keptn.NewLogger("", "", "dynatrace-service")
		serviceSynchronizerInstance = &serviceSynchronizer{
			logger: logger,
		}

		dynatraceConfig, err := config.GetDynatraceConfig(initSyncEventAdapter{}, logger)
		if err != nil {
			logger.Error("failed to load Dynatrace config: " + err.Error())
			return nil
		}
		creds, err := credentials.GetDynatraceCredentials(dynatraceConfig)
		if err != nil {
			logger.Error("failed to load Dynatrace credentials: " + err.Error())
			return nil
		}

		serviceSynchronizerInstance.DTHelper = NewDynatraceHelper(nil, creds, logger)

		serviceSynchronizerInstance.logger.Debug("Initializing Service Synchronizer")
		var configServiceBaseURL string
		csURL, err := keptn.GetServiceEndpoint("CONFIGURATION_SERVICE")
		if err == nil {
			configServiceBaseURL = csURL.String()
		} else {
			configServiceBaseURL = "http://configuration-service:8080"
		}

		serviceSynchronizerInstance.logger.Debug("Service Synchronizer uses configuration service URL: " + configServiceBaseURL)

		serviceSynchronizerInstance.projectsAPI = keptnapi.NewProjectHandler(configServiceBaseURL)
		serviceSynchronizerInstance.servicesAPI = keptnapi.NewServiceHandler(configServiceBaseURL)
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

func (s *serviceSynchronizer) synchronizeServices() {
	s.logger.Info("checking if project " + defaultDTProjectName + " exists")
	project, _ := s.projectsAPI.GetProject(apimodels.Project{
		ProjectName: defaultDTProjectName,
	})
	if project == nil {
		s.logger.Info("Project " + defaultDTProjectName + " does not exist. Stopping synchronization")
		return
	}
	allKeptnServicesInProject, err := s.servicesAPI.GetAllServices(defaultDTProjectName, defaultDTProjectStage)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Could not fetch services of Keptn project %s: %v", defaultDTProjectName, err))
	}
	s.servicesInKeptn = []string{}
	for _, service := range allKeptnServicesInProject {
		s.servicesInKeptn = append(s.servicesInKeptn, service.ServiceName)
	}
	s.checkForTaggedDynatraceServiceEntities()
}

func (s *serviceSynchronizer) checkForTaggedDynatraceServiceEntities() error {
	s.logger.Info("fetching services with tags 'keptn_managed' and 'keptn_service'")

	nextPageKey := ""
	pageSize := 10

	for {
		entitiesResponse, err := s.fetchKeptnManagedServicesFromDynatrace(nextPageKey, pageSize)
		if err != nil {
			return fmt.Errorf("could not get keptn_managed services: %v", err)
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
	return nil

}

func (s *serviceSynchronizer) synchronizeDTEntityWithKeptn(serviceName string) error {
	s.logger.Debug(fmt.Sprintf("Checking if service %s already exists in Keptn project '%s'", serviceName, defaultDTProjectName))
	serviceExists := doesServiceExist(s.servicesInKeptn, serviceName)

	if serviceExists {
		s.logger.Debug(fmt.Sprintf("Service %s already exists in Keptn project '%s'", serviceName, defaultDTProjectName))
		return nil
	}

	s.logger.Debug(fmt.Sprintf("Service %s does not exist yet in Keptn project '%s'", serviceName, defaultDTProjectName))

	s.logger.Debug(fmt.Sprintf("sending %s event for service %s", keptn.InternalServiceCreateEventType, serviceName))
	createServiceData := keptn.ServiceCreateEventData{
		Project: defaultDTProjectName,
		Service: serviceName,
	}

	source, _ := url.Parse("dynatrace-service")
	contentType := "application/json"
	keptnContext := uuid.New().String()
	ce := &cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        keptn.InternalServiceCreateEventType,
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": keptnContext},
		}.AsV02(),
		Data: createServiceData,
	}

	if s.keptnHandler == nil {
		newKeptn, err := keptn.NewKeptn(ce, keptn.KeptnOpts{})
		if err != nil {
			return fmt.Errorf("could not initialize KeptnHandler: %v", err)
		}
		s.keptnHandler = newKeptn
	}

	s.apiMutex.Lock()
	err := s.keptnHandler.SendCloudEvent(*ce)
	s.apiMutex.Unlock()
	if err != nil {
		return fmt.Errorf("could not send %s for service %s: %v", keptn.InternalServiceCreateEventType, serviceName, err)
	}

	s.logger.Debug(fmt.Sprintf("Sent cloud event to create service. waiting until service %s is available", serviceName))
	// wait for the service to be available
	maxRetries := 5
	serviceAvailable := false
	var createdService *apimodels.Service
	for i := 0; i < maxRetries; i++ {
		<-time.After(3 * time.Second)
		s.apiMutex.Lock()
		createdService, _ = s.servicesAPI.GetService(defaultDTProjectName, defaultDTProjectStage, serviceName)
		s.apiMutex.Unlock()
		if createdService != nil {
			serviceAvailable = true
			break
		}
	}

	if !serviceAvailable {
		return fmt.Errorf("Service %s is not available. Cannot proceed with uploading SLO", serviceName)
	}
	s.logger.Error(fmt.Sprintf("Service %s is available. Proceeding with uploading SLO", serviceName))

	resourceURI := "slo.yaml"
	sloResource := &apimodels.Resource{
		ResourceContent: encodedDefaultSLOFile,
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
			if tag.Key == "keptn_service" && tag.Value != "" && keptn.ValidateKeptnEntityName(tag.Value) {
				return tag.Value

			}
		}
	}
	return entity.EntityID
}

func (s *serviceSynchronizer) fetchKeptnManagedServicesFromDynatrace(nextPageKey string, pageSize int) (*dtEntityListResponse, error) {
	var query string
	if nextPageKey == "" {
		query = "/api/v2/entities?entitySelector=type(\"SERVICE\"),tag(\"keptn_managed\"),tag(\"keptn_service\")&fields=+tags&pageSize=" + strconv.FormatInt(int64(pageSize), 10)
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

package onboard

import (
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnlib "github.com/keptn/go-utils/pkg/lib"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

const defaultDTProjectName = "dynatrace"
const defaultDTProjectStage = "quality-gate"

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

func (initSyncEventAdapter) GetLabels() map[string]string {
	return nil
}

type serviceSynchronizer struct {
	projectClient       keptn.ProjectClientInterface
	servicesClient      keptn.ServiceClientInterface
	resourcesClient     keptn.SLIAndSLOResourceWriterInterface
	apiHandler          *keptnapi.APIHandler
	credentialsProvider credentials.DynatraceCredentialsProvider
	EntitiesClientFunc  func(dtCredentials *credentials.DynatraceCredentials) *dynatrace.EntitiesClient
	syncTimer           *time.Ticker
	keptnHandler        *keptnv2.Keptn
	servicesInKeptn     []string
	dtConfigGetter      config.DynatraceConfigGetterInterface
}

var serviceSynchronizerInstance *serviceSynchronizer

const shipyardController = "SHIPYARD_CONTROLLER"
const defaultShipyardControllerURL = "http://shipyard-controller:8080"

// ActivateServiceSynchronizer godoc
func ActivateServiceSynchronizer(c credentials.DynatraceCredentialsProvider) {
	if serviceSynchronizerInstance == nil {

		serviceSynchronizerInstance = &serviceSynchronizer{
			credentialsProvider: c,
		}

		resourceClient := keptn.NewDefaultResourceClient()

		serviceSynchronizerInstance.dtConfigGetter = config.NewDynatraceConfigGetter(resourceClient)
		serviceSynchronizerInstance.EntitiesClientFunc =
			func(credentials *credentials.DynatraceCredentials) *dynatrace.EntitiesClient {
				dtClient := dynatrace.NewClient(credentials)
				return dynatrace.NewEntitiesClient(dtClient)
			}

		configServiceBaseURL := common.GetConfigurationServiceURL()
		shipyardControllerBaseURL := common.GetShipyardControllerURL()
		log.WithFields(
			log.Fields{
				"configServiceBaseURL":      configServiceBaseURL,
				"shipyardControllerBaseURL": shipyardControllerBaseURL,
			}).Debug("Initializing Service Synchronizer")

		serviceSynchronizerInstance.projectClient = keptn.NewDefaultProjectClient()
		serviceSynchronizerInstance.servicesClient = keptn.NewDefaultServiceClient()
		serviceSynchronizerInstance.resourcesClient = resourceClient

		serviceSynchronizerInstance.initializeSynchronizationTimer()
	}
}

func (s *serviceSynchronizer) initializeSynchronizationTimer() {
	syncInterval := env.GetServiceSyncInterval()
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
	creds, err := s.establishDTAPIConnection()
	if err != nil {
		log.WithError(err).Error("Could not establish Dynatrace API connection")
		return
	}

	log.WithField("project", defaultDTProjectName).Info("Fetching existing services in project")
	if err := s.fetchExistingServices(); err != nil {
		log.WithError(err).Error("Could not fetch existing services")
		return
	}

	log.Info("Fetching service entities with tags 'keptn_managed' and 'keptn_service'")

	entitiesClient := s.EntitiesClientFunc(creds)
	entities, err := entitiesClient.GetKeptnManagedServices()
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

func (s *serviceSynchronizer) establishDTAPIConnection() (*credentials.DynatraceCredentials, error) {
	dynatraceConfig, err := s.dtConfigGetter.GetDynatraceConfig(initSyncEventAdapter{})
	if err != nil {
		return nil, fmt.Errorf("failed to load Dynatrace config: %s", err.Error())
	}

	creds, err := s.credentialsProvider.GetDynatraceCredentials(dynatraceConfig.DtCreds)
	if err != nil {
		return nil, fmt.Errorf("failed to load Dynatrace credentials: %s", err.Error())
	}

	return creds, nil
}

func (s *serviceSynchronizer) fetchExistingServices() error {
	err := s.projectClient.AssertProjectExists(defaultDTProjectName)
	if err != nil {
		return err
	}

	// get all services currently in the project
	s.servicesInKeptn = []string{}
	serviceNames, err := s.servicesClient.GetServiceNames(defaultDTProjectName, defaultDTProjectStage)
	if err != nil {
		return err
	}
	s.servicesInKeptn = serviceNames

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
	err := s.servicesClient.CreateServiceInProject(defaultDTProjectName, serviceName)
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

func (s *serviceSynchronizer) createSLOResource(serviceName string) error {
	defaultSLOs := &keptnlib.ServiceLevelObjectives{
		SpecVersion: "1.0",
		Filter:      nil,
		Comparison: &keptnlib.SLOComparison{
			AggregateFunction:         "avg",
			CompareWith:               "single_result",
			IncludeResultWithScore:    "pass",
			NumberOfComparisonResults: 1,
		},
		Objectives: []*keptnlib.SLO{
			{
				SLI:     "response_time_p95",
				KeySLI:  false,
				Pass:    []*keptnlib.SLOCriteria{{Criteria: []string{"<600"}}},
				Warning: []*keptnlib.SLOCriteria{{Criteria: []string{"<=800"}}},
				Weight:  1,
			},
			{
				SLI:    "error_rate",
				KeySLI: false,
				Pass:   []*keptnlib.SLOCriteria{{Criteria: []string{"<5"}}},
				Weight: 1,
			},
			{
				SLI: "throughput",
			},
		},
		TotalScore: &keptnlib.SLOScore{
			Pass:    "90%",
			Warning: "75%",
		},
	}

	err := s.resourcesClient.UploadSLOs(defaultDTProjectName, defaultDTProjectStage, serviceName, defaultSLOs)
	if err != nil {
		return err
	}

	return nil
}

func (s *serviceSynchronizer) createSLIResource(serviceName string) error {
	indicators := make(map[string]string)
	indicators["throughput"] = fmt.Sprintf("metricSelector=builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", serviceName)
	indicators["error_rate"] = fmt.Sprintf("metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", serviceName)
	indicators["response_time_p50"] = fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", serviceName)
	indicators["response_time_p90"] = fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(90)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", serviceName)
	indicators["response_time_p95"] = fmt.Sprintf("metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:%s)", serviceName)

	defaultSLIs := &dynatrace.SLI{
		SpecVersion: "1.0",
		Indicators:  indicators,
	}

	err := s.resourcesClient.UploadSLI(defaultDTProjectName, defaultDTProjectStage, serviceName, defaultSLIs)
	if err != nil {
		return err
	}

	return nil
}

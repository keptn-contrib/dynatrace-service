package onboard

import (
	"context"
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/config"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	keptnlib "github.com/keptn/go-utils/pkg/lib"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	log "github.com/sirupsen/logrus"
)

const synchronizedProject = "dynatrace"
const synchronizedStage = "quality-gate"

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
	return synchronizedProject
}

func (initSyncEventAdapter) GetStage() string {
	return synchronizedStage
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

// EntitiesClientFactory defines a factory that can get EntitiesClients.
type EntitiesClientFactory interface {
	// CreateEntitiesClient creates a dynatrace.EntitiesClient or returns an error.
	CreateEntitiesClient(ctx context.Context) (*dynatrace.EntitiesClient, error)
}

type defaultEntitiesClientFactory struct {
	configProvider config.DynatraceConfigProvider
}

func newDefaultEntitiesClientFactory(resourceClient keptn.DynatraceConfigReaderInterface) *defaultEntitiesClientFactory {
	return &defaultEntitiesClientFactory{
		configProvider: config.NewDynatraceConfigGetter(resourceClient),
	}
}

// CreateEntitiesClient creates a dynatrace.EntitiesClient or returns an error.
func (f defaultEntitiesClientFactory) CreateEntitiesClient(ctx context.Context) (*dynatrace.EntitiesClient, error) {
	dynatraceConfig, err := f.configProvider.GetDynatraceConfig(initSyncEventAdapter{})
	if err != nil {
		return nil, fmt.Errorf("failed to load Dynatrace config: %w", err)
	}

	credentialsProvider, err := credentials.NewDefaultDynatraceK8sSecretReader()
	if err != nil {
		return nil, err
	}

	credentials, err := credentialsProvider.GetDynatraceCredentials(ctx, dynatraceConfig.DtCreds)
	if err != nil {
		return nil, fmt.Errorf("failed to load Dynatrace credentials: %w", err)
	}

	dynatraceClient := dynatrace.NewClient(credentials)
	return dynatrace.NewEntitiesClient(dynatraceClient), nil
}

// ServiceSynchronizer encapsulates the service onboarder component.
type ServiceSynchronizer struct {
	servicesClient        keptn.ServiceClientInterface
	resourcesClient       keptn.SLIAndSLOWriterInterface
	entitiesClientFactory EntitiesClientFactory
}

// NewDefaultServiceSynchronizer creates a new default ServiceSynchronizer.
func NewDefaultServiceSynchronizer() *ServiceSynchronizer {
	clientSet := keptn.NewClientFactory()
	resourceClient := keptn.NewConfigClient(clientSet.CreateResourceClient())

	serviceSynchronizer := ServiceSynchronizer{
		servicesClient:        clientSet.CreateServiceClient(),
		resourcesClient:       resourceClient,
		entitiesClientFactory: newDefaultEntitiesClientFactory(resourceClient),
	}

	return &serviceSynchronizer
}

// Run runs the service synchronizer which does not return unless cancelled.
// Cancelling runCtx will stop any new synchronization runs, cancelling synchronizationCtx will stop an in progress synchronization.
func (s *ServiceSynchronizer) Run(runCtx context.Context, synchronizationCtx context.Context) {
	syncInterval := env.GetServiceSyncInterval()
	log.WithField("syncInterval", syncInterval).Info("Service Synchronizer will sync periodically")
	for {
		s.synchronizeServices(synchronizationCtx)

		select {
		case <-runCtx.Done():
			log.Info("Service Synchronizer has terminated")
			return

		case <-time.After(time.Duration(syncInterval) * time.Second):
		}

		log.WithField("delaySeconds", syncInterval).Info("Synchronizing services")
	}
}

// synchronizeServices performs a single synchronization run
func (s *ServiceSynchronizer) synchronizeServices(ctx context.Context) {
	existingServices, err := s.getExistingServicesFromKeptn()
	if err != nil {
		log.WithError(err).Error("Could not get existing services from Keptn")
		return
	}

	entities, err := s.getKeptnManagedServicesFromDynatrace(ctx)
	if err != nil {
		log.WithError(err).Error("Could not get Keptn-managed services from Dynatrace")
		return
	}

	for _, entity := range entities {

		service, err := getServiceFromEntity(entity)
		if err != nil {
			log.WithError(err).WithField("entityId", entity.EntityID).Debug("Skipping entity due to no valid service name")
			continue
		}

		if doesServiceExist(existingServices, service) {
			log.WithFields(log.Fields{
				"service":  service,
				"entityId": entity.EntityID,
			}).Debug("Service already exists in project, skipping")
			continue
		}

		if err := s.addServiceToKeptn(service); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"service":  service,
				"entityId": entity.EntityID,
			}).Error("Could not add service")
			continue
		}

		existingServices = append(existingServices, service)
	}
}

func (s *ServiceSynchronizer) getExistingServicesFromKeptn() ([]string, error) {
	return s.servicesClient.GetServiceNames(synchronizedProject, synchronizedStage)
}

func (s *ServiceSynchronizer) getKeptnManagedServicesFromDynatrace(ctx context.Context) ([]dynatrace.Entity, error) {
	entitiesClient, err := s.entitiesClientFactory.CreateEntitiesClient(ctx)
	if err != nil {
		return nil, err
	}

	entities, err := entitiesClient.GetKeptnManagedServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Keptn managed services from Dynatrace: %w", err)
	}

	return entities, nil
}

func getServiceFromEntity(entity dynatrace.Entity) (string, error) {
	serviceTags := make([]string, 0)
	for _, tag := range entity.Tags {
		if tag.Key == "keptn_service" && tag.Value != "" {
			serviceTags = append(serviceTags, tag.Value)
		}
	}
	if len(serviceTags) == 0 {
		return "", fmt.Errorf("entity %v has no valid 'keptn_service' tag", entity.EntityID)
	}
	if len(serviceTags) > 1 {
		return "", fmt.Errorf("entity %v has multiple 'keptn_service' tags", entity.EntityID)
	}

	return serviceTags[0], nil
}

func doesServiceExist(services []string, serviceName string) bool {
	for _, service := range services {
		if service == serviceName {
			return true
		}
	}
	return false
}

func (s *ServiceSynchronizer) addServiceToKeptn(serviceName string) error {
	err := s.servicesClient.CreateServiceInProject(synchronizedProject, serviceName)
	if err != nil {
		return fmt.Errorf("could not create service %s: %s", serviceName, err)
	}

	if err := s.createSLOResource(serviceName); err == nil {
		log.WithField("service", serviceName).Info("Uploaded slo.yaml for service")
	} else {
		log.WithError(err).WithField("service", serviceName).Info("Could not create SLO resource for service")
	}

	if err := s.createSLIResource(serviceName); err == nil {
		log.WithField("service", serviceName).Info("Uploaded sli.yaml for service")
	} else {
		log.WithError(err).WithField("service", serviceName).Info("Could not create SLI resource for service")
	}

	return nil
}

func (s *ServiceSynchronizer) createSLOResource(serviceName string) error {
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

	err := s.resourcesClient.UploadSLOs(synchronizedProject, synchronizedStage, serviceName, defaultSLOs)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceSynchronizer) createSLIResource(serviceName string) error {
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

	err := s.resourcesClient.UploadSLIs(synchronizedProject, synchronizedStage, serviceName, defaultSLIs)
	if err != nil {
		return err
	}

	return nil
}

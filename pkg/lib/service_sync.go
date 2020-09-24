package lib

import (
	"fmt"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"os"
	"strconv"
	"time"
)

const defaultDTProjectName = ""
const defaultDTProjectStage = ""
const defaultSLOFile = ``

type serviceSynchronizer struct {
	logger      keptn.LoggerInterface
	projectsAPI *keptnapi.ProjectHandler
	servicesAPI *keptnapi.ServiceHandler
	DTHelper    *DynatraceHelper
	syncTimer   *time.Ticker
}

var serviceSynchronizerInstance *serviceSynchronizer

// ActivateServiceSynchronizer godoc
func ActivateServiceSynchronizer() *serviceSynchronizer {
	if serviceSynchronizerInstance == nil {

		serviceSynchronizerInstance = &serviceSynchronizer{
			logger: keptn.NewLogger("", "", "dynatrace-service"),
		}
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
	s.logger.Info(fmt.Sprintf("Service Synchronizer will sync every %d seconds", syncInterval))
	syncInterval = int(parseInt)
	s.syncTimer = time.NewTicker(time.Duration(syncInterval) * time.Second)
	go func() {
		for {
			<-s.syncTimer.C
			s.logger.Info(fmt.Sprintf("%d seconds have passed. Synchronizing services", syncInterval))
			s.synchronizeServices()
		}
	}()
}

func (s *serviceSynchronizer) synchronizeServices() {

}

func (s *serviceSynchronizer) checkForTaggedDynatraceServiceEntities() ([]Entities, error) {
	return nil, nil
}

type DTEntityListResponse struct {
	TotalCount int        `json:"totalCount"`
	PageSize   int        `json:"pageSize"`
	Entities   []Entities `json:"entities"`
}
type Tags struct {
	Context              string `json:"context"`
	Key                  string `json:"key"`
	StringRepresentation string `json:"stringRepresentation"`
	Value                string `json:"value,omitempty"`
}
type Entities struct {
	EntityID    string `json:"entityId"`
	DisplayName string `json:"displayName"`
	Tags        []Tags `json:"tags"`
}

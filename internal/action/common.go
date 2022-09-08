package action

import (
	"context"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
)

const eventSource = "Keptn dynatrace-service"
const bridgeURLKey = "Keptns Bridge"

const contextless = "CONTEXTLESS"

type customProperties map[string]string

func newCustomProperties(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, bridgeURL string) customProperties {
	cp := customProperties{
		"Project":       a.GetProject(),
		"Stage":         a.GetStage(),
		"Service":       a.GetService(),
		"TestStrategy":  a.GetTestStrategy(),
		"Image":         imageAndTag.Image(),
		"Tag":           imageAndTag.Tag(),
		"KeptnContext":  a.GetShKeptnContext(),
		"Keptn Service": a.GetSource(),
	}

	for key, value := range a.GetLabels() {
		cp.add(key, value)
	}

	cp.addIfNonEmpty(bridgeURLKey, bridgeURL)

	return cp
}

func (cp customProperties) add(key string, value string) {
	oldValue, isContained := cp[key]
	if isContained {
		log.WithFields(
			log.Fields{
				"key":      key,
				"oldValue": oldValue,
				"value":    value,
			}).Warn("Overwriting value in custom properties")
	}

	cp[key] = value
}

func (cp customProperties) addIfNonEmpty(key string, value string) {
	if key == "" || value == "" {
		return
	}

	cp.add(key, value)
}

func getValueFromLabels(a adapter.EventContentAdapter, key string, defaultValue string) string {
	v := a.GetLabels()[key]
	if v != "" {
		return v
	}
	return defaultValue
}

// KeptnContext is a minimal subset of data needed for creating default attach rules
type KeptnContext interface {
	GetProject() string
	GetStage() string
	GetService() string
}

func createDefaultAttachRules(keptnContext KeptnContext) *dynatrace.AttachRules {
	return &dynatrace.AttachRules{
		TagRule: []dynatrace.TagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []dynatrace.TagEntry{
					{
						Context: contextless,
						Key:     "keptn_project",
						Value:   keptnContext.GetProject(),
					},
					{
						Context: contextless,
						Key:     "keptn_stage",
						Value:   keptnContext.GetStage(),
					},
					{
						Context: contextless,
						Key:     "keptn_service",
						Value:   keptnContext.GetService(),
					},
				},
			},
		},
	}
}

// TimeframeFunc is the signature of a function returning a common.Timeframe or and error
type TimeframeFunc func() (*common.Timeframe, error)

func createOrUpdateAttachRules(ctx context.Context, client dynatrace.ClientInterface, existingAttachRules *dynatrace.AttachRules, imageAndTag common.ImageAndTag, event adapter.EventContentAdapter, timeframe *common.Timeframe) dynatrace.AttachRules {
	version := determineVersionFromTagOrLabel(imageAndTag, event)
	if version == "" || timeframe == nil {
		if existingAttachRules != nil {
			log.WithFields(log.Fields{
				"version":           version,
				"timeframe":         timeframe,
				"customAttachRules": *existingAttachRules,
			}).Debug("no version information available - will use customer provided attach rules")
			return *existingAttachRules
		}

		log.WithFields(log.Fields{
			"version":   version,
			"timeframe": timeframe,
		}).Debug("no version information or time frame available - will use default attach rules")
		return *createDefaultAttachRules(event)
	}

	entityClient := dynatrace.NewEntitiesClient(client)
	pgis, err := entityClient.GetAllPGIsForKeptnServices(ctx, dynatrace.PGIQueryConfig{
		Project: event.GetProject(),
		Stage:   event.GetStage(),
		Service: event.GetService(),
		Version: version,
		From:    timeframe.Start(),
		To:      timeframe.End(),
	})
	if err != nil {
		log.WithError(err).WithField("version", version).Error("could not find PGIs for version")
	}

	if existingAttachRules != nil {
		if len(pgis) == 0 {
			log.WithField("customAttachRules", *existingAttachRules).Debug("no PGIs found - will use customer provided attach rules only")
			return *existingAttachRules
		}

		log.WithFields(log.Fields{
			"customAttachRules": *existingAttachRules,
			"entityIds":         pgis,
		}).Debug("PGIs found and custom attach rules - will combine them")
		existingAttachRules.EntityIds = append(existingAttachRules.EntityIds, pgis...)
		return *existingAttachRules
	}

	if len(pgis) == 0 {
		log.Debug("no PGIs found and no custom attach rules - will use default attach rules")
		return *createDefaultAttachRules(event)
	}

	log.WithField("PGIs", pgis).Debug("PGIs found - will use them only")
	return dynatrace.AttachRules{
		EntityIds: pgis,
	}
}

func determineVersionFromTagOrLabel(imageAndTag common.ImageAndTag, event adapter.EventContentAdapter) string {

	if imageAndTag.HasTag() {
		return imageAndTag.Tag()
	}

	return event.GetLabels()["releasesVersion"]
}

func createAttachRulesForDeploymentTimeFrame(ctx context.Context, dtClient dynatrace.ClientInterface, eClient keptn.EventClientInterface, event adapter.EventContentAdapter, imageAndTag common.ImageAndTag, customAttachRules *dynatrace.AttachRules) dynatrace.AttachRules {

	deploymentTriggeredTime, err := eClient.GetEventTimeStampForType(ctx, event, keptnv2.GetStartedEventType(keptnv2.DeploymentTaskName))
	if err != nil {
		log.WithError(err).Warn("Could not find the corresponding deployment.triggered event")
	}

	deploymentFinishedTime, err := eClient.GetEventTimeStampForType(ctx, event, keptnv2.GetFinishedEventType(keptnv2.DeploymentTaskName))
	if err != nil {
		log.WithError(err).Warn("Could not find the corresponding deployment.finished event")
	}

	var timeframe *common.Timeframe
	if deploymentTriggeredTime != nil && deploymentFinishedTime != nil {
		// ignoring error here, as it should be fine anyway - otherwise attach rules will be set to default / custom
		timeframe, _ = common.NewTimeframe(*deploymentTriggeredTime, *deploymentFinishedTime)
	}

	return createOrUpdateAttachRules(ctx, dtClient, customAttachRules, imageAndTag, event, timeframe)
}

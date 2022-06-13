package action

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

const eventSource = "Keptn dynatrace-service"
const bridgeURLKey = "Keptns Bridge"

const contextless = "CONTEXTLESS"

func createCustomProperties(a adapter.EventContentAdapter, imageAndTag common.ImageAndTag, bridgeURL string) map[string]string {
	customProperties := map[string]string{
		"Project":       a.GetProject(),
		"Stage":         a.GetStage(),
		"Service":       a.GetService(),
		"TestStrategy":  a.GetTestStrategy(),
		"Image":         imageAndTag.Image(),
		"Tag":           imageAndTag.Tag(),
		"KeptnContext":  a.GetShKeptnContext(),
		"Keptn Service": a.GetSource(),
	}

	// now add the rest of the labels into custom properties (changed with #115_116)
	for key, value := range a.GetLabels() {
		customProperties[key] = value
	}

	if bridgeURL != "" {
		customProperties[bridgeURLKey] = bridgeURL
	}

	return customProperties
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

func createOrUpdateAttachRules(ctx context.Context, client dynatrace.ClientInterface, existingAttachRules *dynatrace.AttachRules, imageAndTag common.ImageAndTag, event adapter.EventContentAdapter, timeframeFunc TimeframeFunc) (dynatrace.AttachRules, error) {
	version := determineVersionFromTagOrLabel(imageAndTag, event)
	if version == "" {
		if existingAttachRules != nil {
			log.WithField("customAttachRules", *existingAttachRules).Debug("no version information available - will use customer provided attach rules")
			return *existingAttachRules, nil
		}

		log.Debug("no version information available - will use default attach rules")
		return *createDefaultAttachRules(event), nil
	}

	timeframe, err := timeframeFunc()
	if err != nil {
		return dynatrace.AttachRules{}, err
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
		return dynatrace.AttachRules{}, err
	}

	if existingAttachRules != nil {
		if len(pgis) == 0 {
			log.WithField("customAttachRules", *existingAttachRules).Debug("no PGIs found - will use customer provided attach rules only")
			return *existingAttachRules, nil
		}

		log.WithFields(log.Fields{
			"customAttachRules": *existingAttachRules,
			"entityIds":         pgis,
		}).Debug("PGIs found and custom attach rules - will combine them")
		existingAttachRules.EntityIds = append(existingAttachRules.EntityIds, pgis...)
		return *existingAttachRules, nil
	}

	if len(pgis) == 0 {
		log.Debug("no PGIs found and no custom attach rules - will use default attach rules")
		return *createDefaultAttachRules(event), nil
	}

	log.WithField("PGIs", pgis).Debug("PGIs found - will use them only")
	return dynatrace.AttachRules{
		EntityIds: pgis,
	}, nil
}

func determineVersionFromTagOrLabel(imageAndTag common.ImageAndTag, event adapter.EventContentAdapter) string {

	if imageAndTag.HasTag() {
		return imageAndTag.Tag()
	}

	return event.GetLabels()["releasesVersion"]
}

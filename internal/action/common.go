package action

import (
	"context"

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

type TimeframeFunc func() (*common.Timeframe, error)

func createOrUpdateAttachRules(ctx context.Context, client dynatrace.ClientInterface, existingAttachRules *dynatrace.AttachRules, imageAndTag common.ImageAndTag, keptnContext KeptnContext, timeframeFunc TimeframeFunc) (dynatrace.AttachRules, error) {
	if imageAndTag.DoesNotHaveTag() {
		if existingAttachRules != nil {
			return *existingAttachRules, nil
		}

		return *createDefaultAttachRules(keptnContext), nil
	}

	timeframe, err := timeframeFunc()
	if err != nil {
		return dynatrace.AttachRules{}, err
	}

	entityClient := dynatrace.NewEntitiesClient(client)
	pgis, err := entityClient.GetAllPGIsForKeptnServices(ctx, dynatrace.PGIQueryConfig{
		Project: keptnContext.GetProject(),
		Stage:   keptnContext.GetStage(),
		Service: keptnContext.GetService(),
		Version: imageAndTag.Tag(),
		From:    timeframe.Start(),
		To:      timeframe.End(),
	})
	if err != nil {
		return dynatrace.AttachRules{}, err
	}

	if existingAttachRules != nil {
		if len(pgis) == 0 {
			return *existingAttachRules, nil
		}

		existingAttachRules.EntityIds = append(existingAttachRules.EntityIds, pgis...)
		return *existingAttachRules, nil
	}

	if len(pgis) == 0 {
		return *createDefaultAttachRules(keptnContext), nil
	}

	return dynatrace.AttachRules{
		EntityIds: pgis,
	}, nil
}

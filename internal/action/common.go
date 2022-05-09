package action

import (
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
)

const eventSource = "Keptn dynatrace-service"
const bridgeURLKey = "Keptns Bridge"

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

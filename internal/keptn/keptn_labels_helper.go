package keptn

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

// AddOptionalKeptnBridgeUrlToLabels adds a backlink to the Keptn bridge if the URL has been provided.
// If the provided labels are nil, a new empty map is created.
func AddOptionalKeptnBridgeUrlToLabels(labels map[string]string, shKeptnContext string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	keptnBridgeURL, _ := credentials.GetKeptnBridgeURL()
	if keptnBridgeURL != "" {
		labels[common.KEPTNSBRIDGE_LABEL] = keptnBridgeURL + "/trace/" + shKeptnContext
	}
	return labels
}

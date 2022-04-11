package keptn

import (
	"net/url"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

// AddOptionalKeptnBridgeUrlToLabels adds a backlink to the Keptn bridge if the URL has been provided.
// If the provided labels are nil, a new empty map is created.
func AddOptionalKeptnBridgeUrlToLabels(labels map[string]string, shKeptnContext string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	credentials, err := credentials.GetKeptnCredentials()
	if err != nil {
		return labels
	}

	keptnBridgeURL := credentials.GetBridgeURL()
	if keptnBridgeURL == "" {
		return labels
	}

	labels[common.BridgeLabel] = keptnBridgeURL + "/trace/" + shKeptnContext
	return labels
}

// TryGetProblemIDFromLabels tries to extract the problem ID from a "Problem URL" label or returns "" if it cannot be done.
// The value should be of form https://dynatracetenant/#problems/problemdetails;pid=8485558334848276629_1604413609638V2
func TryGetProblemIDFromLabels(keptnEvent adapter.EventContentAdapter) string {
	for labelName, labelValue := range keptnEvent.GetLabels() {
		if strings.EqualFold(labelName, common.ProblemURLLabel) {
			u, err := url.Parse(labelValue)
			if err != nil {
				return ""
			}

			ix := strings.LastIndex(u.Fragment, ";pid=")
			if ix == -1 {
				return ""
			}

			return u.Fragment[ix+5:]
		}
	}

	return ""
}

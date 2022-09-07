package keptn

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

// BridgeURLCreatorInterface defines a component that can get URLs for Keptn contexts and evaluations.
type BridgeURLCreatorInterface interface {
	TryGetBridgeURLForKeptnContext(ctx context.Context, event adapter.EventContentAdapter) string
	TryGetBridgeURLForEvaluation(ctx context.Context, event adapter.EventContentAdapter) string
}

// BridgeURLCreator is the default implementation of BridgeURLCreatorInterface.
type BridgeURLCreator struct {
	credentialsProvider credentials.KeptnCredentialsProvider
}

// NewBridgeURLCreator creates a new BridgeURLCreator.
func NewBridgeURLCreator(credentialsProvider credentials.KeptnCredentialsProvider) *BridgeURLCreator {
	return &BridgeURLCreator{
		credentialsProvider: credentialsProvider,
	}
}

// TryGetBridgeURLForKeptnContext gets a backlink to the Keptn Bridge if available or returns empty string.
func (c *BridgeURLCreator) TryGetBridgeURLForKeptnContext(ctx context.Context, event adapter.EventContentAdapter) string {
	keptnBridgeURL := c.tryGetBridgeURL(ctx)
	if keptnBridgeURL == "" {
		return ""
	}

	return keptnBridgeURL + "/trace/" + event.GetShKeptnContext()
}

// TryGetBridgeURLForEvaluation gets a backlink to the evaluation in Keptn Bridge if available or returns empty string.
func (c *BridgeURLCreator) TryGetBridgeURLForEvaluation(ctx context.Context, event adapter.EventContentAdapter) string {
	keptnBridgeURL := c.tryGetBridgeURL(ctx)
	if keptnBridgeURL == "" {
		return ""
	}

	return fmt.Sprintf("%s/evaluation/%s/%s", keptnBridgeURL, event.GetShKeptnContext(), event.GetStage())
}

// tryGetBridgeURL gets the Keptn Bridge URL if available or returns empty string.
func (c *BridgeURLCreator) tryGetBridgeURL(ctx context.Context) string {
	creds, err := c.credentialsProvider.GetKeptnCredentials(ctx)
	if err != nil {
		return ""
	}

	return creds.GetBridgeURL()
}

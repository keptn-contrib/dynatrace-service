package dashboard

import (
	"context"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/ff"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/slo"
)

// SLOTileProcessing represents the processing of a SLO dashboard tile.
type SLOTileProcessing struct {
	client       dynatrace.ClientInterface
	timeframe    common.Timeframe
	featureFlags ff.GetSLIFeatureFlags
}

// NewSLOTileProcessing creates a new SLOTileProcessing.
func NewSLOTileProcessing(client dynatrace.ClientInterface, timeframe common.Timeframe, flags ff.GetSLIFeatureFlags) *SLOTileProcessing {
	return &SLOTileProcessing{
		client:       client,
		timeframe:    timeframe,
		featureFlags: flags,
	}
}

// Process processes the specified SLO dashboard tile.
func (p *SLOTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile) []result.SLIWithSLO {
	if len(tile.AssignedEntities) == 0 {
		return []result.SLIWithSLO{result.NewFailedSLIWithSLO(result.CreateInformationalSLO("slo_tile_without_slo"), "SLO tile contains no SLO IDs")}
	}

	var results []result.SLIWithSLO
	for _, sloID := range tile.AssignedEntities {
		results = append(results, p.processSLO(ctx, sloID))
	}
	return results
}

// processSLO processes an SLO by querying the data from the Dynatrace API.
// Returns a TileResult with sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *SLOTileProcessing) processSLO(ctx context.Context, sloID string) result.SLIWithSLO {
	query, err := slo.NewQuery(sloID)
	if err != nil {
		return result.NewFailedSLIWithSLO(result.CreateInformationalSLO("slo_without_id"), err.Error())
	}

	// Step 1: Query the Dynatrace API to get the actual value for this sloID
	request := dynatrace.NewSLOClientGetRequest(query.GetSLOID(), p.timeframe)
	sloResult, err := dynatrace.NewSLOClient(p.client).Get(ctx, request)
	if err != nil {
		return result.NewFailedSLIWithSLO(result.CreateInformationalSLO(cleanIndicatorName("slo_"+sloID)), "error querying Service level objectives API: "+err.Error())
	}

	indicatorName := cleanIndicatorName(sloResult.Name)

	// TODO: 2021-12-20: check: maybe in the future we will allow users to add additional SLO defs via the Tile Name, e.g: weight or KeySli

	// see https://github.com/keptn-contrib/dynatrace-sli-service/issues/97#issuecomment-766110172 for explanation about mappings to pass and warning
	passCriterion := result.SLOCriteria{Criteria: []string{fmt.Sprintf(">=%f", sloResult.Warning)}}
	warningCriterion := result.SLOCriteria{Criteria: []string{fmt.Sprintf(">=%f", sloResult.Target)}}

	return result.NewSuccessfulSLIWithSLOAndQuery(
		result.SLO{
			SLI:     indicatorName,
			Pass:    result.SLOCriteriaList{&passCriterion},
			Warning: result.SLOCriteriaList{&warningCriterion},
			Weight:  1,
			KeySLI:  false,
		},
		sloResult.EvaluatedPercentage,
		request.RequestString(),
	)
}

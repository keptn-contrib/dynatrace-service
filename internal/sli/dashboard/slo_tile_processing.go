package dashboard

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/slo"
	keptn "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
)

// SLOTileProcessing represents the processing of a SLO dashboard tile.
type SLOTileProcessing struct {
	client    dynatrace.ClientInterface
	timeframe common.Timeframe
}

// NewSLOTileProcessing creates a new SLOTileProcessing.
func NewSLOTileProcessing(client dynatrace.ClientInterface, timeframe common.Timeframe) *SLOTileProcessing {
	return &SLOTileProcessing{
		client:    client,
		timeframe: timeframe,
	}
}

// Process processes the specified SLO dashboard tile.
func (p *SLOTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile) []*TileResult {
	if len(tile.AssignedEntities) == 0 {
		failedTileResult := newFailedTileResult("slo_tile_without_slo", "SLO tile contains no SLO IDs")
		return []*TileResult{&failedTileResult}
	}

	var results []*TileResult
	for _, sloID := range tile.AssignedEntities {
		log.WithField("sloEntity", sloID).Debug("Processing SLO Definition")
		results = append(results, p.processSLO(ctx, sloID))
	}
	return results
}

// processSLO processes an SLO by querying the data from the Dynatrace API.
// Returns a TileResult with sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *SLOTileProcessing) processSLO(ctx context.Context, sloID string) *TileResult {
	query, err := slo.NewQuery(sloID)
	if err != nil {
		// TODO: 2021-02-14: Check that this indicator name still aligns with all possible errors.
		failedTileResult := newFailedTileResult("slo_without_id", err.Error())
		return &failedTileResult
	}

	// Step 1: Query the Dynatrace API to get the actual value for this sloID
	sloResult, err := dynatrace.NewSLOClient(p.client).Get(ctx, dynatrace.NewSLOClientGetParameters(query.GetSLOID(), p.timeframe))
	if err != nil {
		failedTileResult := newFailedTileResult(cleanIndicatorName("slo_"+sloID), "error querying Service level objectives API: "+err.Error())
		return &failedTileResult
	}

	indicatorName := cleanIndicatorName(sloResult.Name)

	log.WithFields(
		log.Fields{
			"indicatorName": indicatorName,
			"value":         sloResult.EvaluatedPercentage,
		}).Debug("Adding SLO to sloResult")
	// TODO: 2021-12-20: check: maybe in the future we will allow users to add additional SLO defs via the Tile Name, e.g: weight or KeySli

	// see https://github.com/keptn-contrib/dynatrace-sli-service/issues/97#issuecomment-766110172 for explanation about mappings to pass and warning
	passCriterion := keptn.SLOCriteria{Criteria: []string{fmt.Sprintf(">=%f", sloResult.Warning)}}
	warningCriterion := keptn.SLOCriteria{Criteria: []string{fmt.Sprintf(">=%f", sloResult.Target)}}

	sloDefinition := &keptn.SLO{
		SLI:     indicatorName,
		Pass:    []*keptn.SLOCriteria{&passCriterion},
		Warning: []*keptn.SLOCriteria{&warningCriterion},
		Weight:  1,
		KeySLI:  false,
	}

	return &TileResult{
		sliResult:     result.NewSuccessfulSLIResult(indicatorName, sloResult.EvaluatedPercentage),
		sloDefinition: sloDefinition,
		sliName:       indicatorName,
		sliQuery:      slo.NewQueryProducer(*query).Produce(),
	}
}

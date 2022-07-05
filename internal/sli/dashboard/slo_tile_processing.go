package dashboard

import (
	"context"
	"errors"
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
func (p *SLOTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile) *TileResult {
	if len(tile.AssignedEntities) == 0 {
		failedTileResult := newFailedTileResult("slo_tile_without_slo", "SLO tile contains no SLO IDs")
		return &failedTileResult
	}

	if len(tile.AssignedEntities) > 1 {
		failedTileResult := newFailedTileResult("slo_tile_with_multiple_slos", "SLO tile contains multiple SLO IDs")
		return &failedTileResult
	}

	title := extractCustomTitleFromMetric(tile.Metric)
	sloDefinition, err := parseSLODefinition(title)
	var sloDefError *sloDefinitionError
	if errors.As(err, &sloDefError) {
		failedTileResult := newFailedTileResultFromError(sloDefError.sliNameOrTileTitle(), "SLO tile title parsing error", err)
		return &failedTileResult
	}

	if len(sloDefinition.Pass) > 0 {
		failedTileResult := newFailedTileResult(sloDefError.sliNameOrTileTitle(), "SLO tile title must not include pass criteria")
		return &failedTileResult
	}

	if len(sloDefinition.Warning) > 0 {
		failedTileResult := newFailedTileResult(sloDefError.sliNameOrTileTitle(), "SLO tile title must not include warning criteria")
		return &failedTileResult
	}

	return p.processSLO(ctx, tile.AssignedEntities[0], sloDefinition)
}

const customTitleKey = "customTitle"

func extractCustomTitleFromMetric(metric string) string {
	for _, pair := range newKeyValueParsing(metric).parse() {
		if pair.key == customTitleKey && pair.split {
			return pair.value
		}
	}

	return ""
}

// processSLO processes an SLO by querying the data from the Dynatrace API.
// Returns a TileResult with sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *SLOTileProcessing) processSLO(ctx context.Context, sloID string, sloDefinitionFromTitle *keptn.SLO) *TileResult {
	log.WithField("sloEntity", sloID).Debug("Processing SLO Definition")

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

	// see https://github.com/keptn-contrib/dynatrace-sli-service/issues/97#issuecomment-766110172 for explanation about mappings to pass and warning
	passCriterion := keptn.SLOCriteria{Criteria: []string{fmt.Sprintf(">=%f", sloResult.Warning)}}
	warningCriterion := keptn.SLOCriteria{Criteria: []string{fmt.Sprintf(">=%f", sloResult.Target)}}

	sloDefinition := &keptn.SLO{
		SLI:         sloDefinitionFromTitle.SLI,
		DisplayName: sloDefinitionFromTitle.DisplayName,
		Pass:        []*keptn.SLOCriteria{&passCriterion},
		Warning:     []*keptn.SLOCriteria{&warningCriterion},
		Weight:      sloDefinitionFromTitle.Weight,
		KeySLI:      sloDefinitionFromTitle.KeySLI,
	}

	if sloDefinition.SLI == "" {
		sloDefinition.SLI = cleanIndicatorName(sloResult.Name)
	}

	if sloDefinition.DisplayName == "" {
		sloDefinition.DisplayName = sloResult.Name
	}

	log.WithFields(
		log.Fields{
			"indicatorName": sloDefinition.SLI,
			"value":         sloResult.EvaluatedPercentage,
		}).Debug("Adding SLO to sloResult")

	return &TileResult{
		sliResult: result.NewSuccessfulSLIResult(sloDefinition.SLI, sloResult.EvaluatedPercentage),
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
		sliQuery:  slo.NewQueryProducer(*query).Produce(),
	}
}

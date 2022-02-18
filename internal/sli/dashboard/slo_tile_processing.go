package dashboard

import (
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/slo"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type SLOTileProcessing struct {
	client    dynatrace.ClientInterface
	startUnix time.Time
	endUnix   time.Time
}

func NewSLOTileProcessing(client dynatrace.ClientInterface, startUnix time.Time, endUnix time.Time) *SLOTileProcessing {
	return &SLOTileProcessing{
		client:    client,
		startUnix: startUnix,
		endUnix:   endUnix,
	}
}

func (p *SLOTileProcessing) Process(tile *dynatrace.Tile) []*TileResult {
	if len(tile.AssignedEntities) == 0 {
		unsuccessfulTileResult := newUnsuccessfulTileResult("slo_tile_without_slo", "SLO tile contains no SLO IDs")
		return []*TileResult{&unsuccessfulTileResult}
	}

	var results []*TileResult
	for _, sloID := range tile.AssignedEntities {
		log.WithField("sloEntity", sloID).Debug("Processing SLO Definition")
		results = append(results, p.processSLO(sloID, p.startUnix, p.endUnix))
	}
	return results
}

// processSLO processes an SLO by querying the data from the Dynatrace API.
// Returns a TileResult with sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *SLOTileProcessing) processSLO(sloID string, startUnix time.Time, endUnix time.Time) *TileResult {
	query, err := slo.NewQuery(sloID)
	if err != nil {
		// TODO: 2021-02-14: Check that this indicator name still aligns with all possible errors.
		unsuccessfulTileResult := newUnsuccessfulTileResult("slo_without_id", err.Error())
		return &unsuccessfulTileResult
	}

	// Step 1: Query the Dynatrace API to get the actual value for this sloID
	sloResult, err := dynatrace.NewSLOClient(p.client).Get(dynatrace.NewSLOClientGetParameters(query.GetSLOID(), startUnix, endUnix))
	if err != nil {
		unsuccessfulTileResult := newUnsuccessfulTileResult(common.CleanIndicatorName("slo_"+sloID), err.Error())
		return &unsuccessfulTileResult
	}

	// Step 2: Transform the SLO result into an SLI result and SLO definition
	// IndicatorName is based on the SLO Name
	indicatorName := common.CleanIndicatorName(sloResult.Name)
	sliResult := &keptnv2.SLIResult{
		Metric:  indicatorName,
		Value:   sloResult.EvaluatedPercentage,
		Success: true,
	}

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
		sliResult: sliResult,
		objective: sloDefinition,
		sliName:   indicatorName,
		sliQuery:  slo.NewQueryProducer(*query).Produce(),
	}
}

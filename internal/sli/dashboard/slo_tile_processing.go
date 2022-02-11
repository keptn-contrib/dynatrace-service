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
	// we will take the SLO definition from Dynatrace
	var results []*TileResult

	for _, sloEntity := range tile.AssignedEntities {
		log.WithField("sloEntity", sloEntity).Debug("Processing SLO Definition")

		tileResult, err := p.processSLOTile(sloEntity, p.startUnix, p.endUnix)
		if err != nil {
			log.WithError(err).Error("Error Processing SLO")
			continue
		}

		results = append(results, tileResult)
	}

	return results
}

// processSLOTile Processes an SLO Tile and queries the data from the Dynatrace API.
// If successful returns sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *SLOTileProcessing) processSLOTile(sloID string, startUnix time.Time, endUnix time.Time) (*TileResult, error) {

	// Step 1: Query the Dynatrace API to get the actual value for this sloID
	sloResult, err := dynatrace.NewSLOClient(p.client).Get(dynatrace.NewSLOClientGetParameters(sloID, startUnix, endUnix))
	if err != nil {
		return nil, err
	}

	// Step 2: As we have the SLO Result including SLO Definition we add it to the SLI & SLO objects
	// IndicatorName is based on the slo Name
	// the value defaults to the E
	indicatorName := common.CleanIndicatorName(sloResult.Name)
	value := sloResult.EvaluatedPercentage
	sliResult := &keptnv2.SLIResult{
		Metric:  indicatorName,
		Value:   value,
		Success: true,
	}

	log.WithFields(
		log.Fields{
			"indicatorName": indicatorName,
			"value":         value,
		}).Debug("Adding SLO to sloResult")

	query, err := slo.NewQuery(sloID)
	if err != nil {
		return nil, err
	}

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
	}, nil

}

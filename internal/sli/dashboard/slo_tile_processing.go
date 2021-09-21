package dashboard

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
	"time"
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
		endUnix:   startUnix,
	}
}

func (p *SLOTileProcessing) Process(tile *dynatrace.Tile) []*TileResult {
	// we will take the SLO definition from Dynatrace
	var results []*TileResult

	for _, sloEntity := range tile.AssignedEntities {
		log.WithField("sloEntity", sloEntity).Debug("Processing SLO Definition")

		sliResult, sliIndicator, sliQuery, sloDefinition, err := p.processSLOTile(sloEntity, p.startUnix, p.endUnix)
		if err != nil {
			log.WithError(err).Error("Error Processing SLO")
			continue
		}

		results = append(
			results,
			&TileResult{
				sliResult: sliResult,
				objective: sloDefinition,
				sliName:   sliIndicator,
				sliQuery:  sliQuery,
			})
	}

	return results
}

// processSLOTile Processes an SLO Tile and queries the data from the Dynatrace API.
// If successful returns sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *SLOTileProcessing) processSLOTile(sloID string, startUnix time.Time, endUnix time.Time) (*keptnv2.SLIResult, string, string, *keptncommon.SLO, error) {

	// Step 1: Query the Dynatrace API to get the actual value for this sloID
	sloResult, err := dynatrace.NewSLOClient(p.client).Get(sloID, startUnix, endUnix)
	if err != nil {
		return nil, "", "", nil, err
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

	// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
	// we prepend this with SLO;<SLO-ID>
	sliQuery := fmt.Sprintf("SLO;%s", sloID)

	// lets add the SLO definition in case we need to generate an SLO.yaml
	// we normally parse these values from the tile name. In this case we just build that tile name -> maybe in the future we will allow users to add additional SLO defs via the Tile Name, e.g: weight or KeySli

	// Please see https://github.com/keptn-contrib/dynatrace-sli-service/issues/97 - for more information on that change of Dynatrace SLO API
	// if we still run against an old API we fall back to the old fields
	warning := sloResult.Warning
	if warning <= 0.0 {
		warning = sloResult.TargetWarningOLD
	}
	target := sloResult.Target
	if target <= 0.0 {
		target = sloResult.TargetSuccessOLD
	}
	sloString := fmt.Sprintf("sli=%s;pass=>=%f;warning=>=%f", indicatorName, warning, target)
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(sloString)

	return sliResult, indicatorName, sliQuery, sloDefinition, nil
}

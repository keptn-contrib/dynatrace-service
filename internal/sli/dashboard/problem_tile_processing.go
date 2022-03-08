package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/problemsv2"
	keptn "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
)

const problemsIndicatorName = "problems"

// ProblemTileProcessing represents the processing of a problems dashboard tile.
type ProblemTileProcessing struct {
	client    dynatrace.ClientInterface
	timeframe common.Timeframe
}

// NewProblemTileProcessing creates a new ProblemTileProcessing.
func NewProblemTileProcessing(client dynatrace.ClientInterface, timeframe common.Timeframe) *ProblemTileProcessing {
	return &ProblemTileProcessing{
		client:    client,
		timeframe: timeframe,
	}
}

// Process retrieves the open problem count and returns this as a TileResult.
// An SLO definition with a pass criteria of <= 0 is also included as we don't allow problems.
func (p *ProblemTileProcessing) Process(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) *TileResult {
	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	// query the number of open problems based on the management zone filter of the tile
	problemSelector := "status(\"open\")" + tileManagementZoneFilter.ForProblemSelector()
	return p.processOpenProblemTile(problems.NewQuery(problemSelector, ""))
}

func (p *ProblemTileProcessing) processOpenProblemTile(query problems.Query) *TileResult {

	sliResult := p.getProblemCountAsSLIResult(query)

	log.WithFields(
		log.Fields{
			"indicatorName": problemsIndicatorName,
			"value":         sliResult.Value,
		}).Debug("Adding SLO to sloResult")

	// TODO: 2022-02-14: check: maybe in the future we will allow users to add additional SLO defs via the tile name, e.g. weight or KeySli.
	sloDefinition := &keptn.SLO{
		SLI:    problemsIndicatorName,
		Pass:   []*keptn.SLOCriteria{{Criteria: []string{"<=0"}}},
		Weight: 1,
		KeySLI: true,
	}

	return &TileResult{
		sliResult: sliResult,
		objective: sloDefinition,
		sliName:   problemsIndicatorName,
		sliQuery:  problemsv2.NewQueryProducer(query).Produce(),
	}
}

func (p *ProblemTileProcessing) getProblemCountAsSLIResult(query problems.Query) result.SLIResult {
	totalProblemCount, err := dynatrace.NewProblemsV2Client(p.client).GetTotalCountByQuery(dynatrace.NewProblemsV2ClientQueryParameters(query, p.timeframe))
	if err != nil {
		return result.NewFailedSLIResult(problemsIndicatorName, "error querying Problems API v2:"+err.Error())
	}

	return result.NewSuccessfulSLIResult(problemsIndicatorName, float64(totalProblemCount))
}

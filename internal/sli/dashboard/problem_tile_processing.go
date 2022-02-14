package dashboard

import (
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/v1/problemsv2"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

const problemsIndicatorName = "problems"

type ProblemTileProcessing struct {
	client    dynatrace.ClientInterface
	startUnix time.Time
	endUnix   time.Time
}

func NewProblemTileProcessing(client dynatrace.ClientInterface, startUnix time.Time, endUnix time.Time) *ProblemTileProcessing {
	return &ProblemTileProcessing{
		client:    client,
		startUnix: startUnix,
		endUnix:   endUnix,
	}
}

func (p *ProblemTileProcessing) Process(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) *TileResult {

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	// query the number of open problems based on the management zone filter of the tile
	problemSelector := "status(open)" + tileManagementZoneFilter.ForProblemSelector()
	tileResult, err := p.processOpenProblemTile(problems.NewQuery(problemSelector, ""), p.startUnix, p.endUnix)
	if err != nil {
		log.WithError(err).Error("Error Processing OPEN_PROBLEMS")
		return nil
	}

	return tileResult
}

// processOpenProblemTile Processes an Open Problem Tile and queries the number of open problems. The current default is that there is a pass criteria of <= 0 as we dont allow problems
// If successful returns sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *ProblemTileProcessing) processOpenProblemTile(query problems.Query, startUnix time.Time, endUnix time.Time) (*TileResult, error) {

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
		sliResult: &sliResult,
		objective: sloDefinition,
		sliName:   problemsIndicatorName,
		sliQuery:  problemsv2.NewQueryProducer(query).Produce(),
	}, nil
}

func (p *ProblemTileProcessing) getProblemCountAsSLIResult(query problems.Query) keptnv2.SLIResult {
	totalProblemCount, err := dynatrace.NewProblemsV2Client(p.client).GetTotalCountByQuery(dynatrace.NewProblemsV2ClientQueryParameters(query, p.startUnix, p.endUnix))
	if err != nil {
		return keptnv2.SLIResult{
			Metric:  problemsIndicatorName,
			Success: false,
			Message: err.Error(),
		}
	}

	return keptnv2.SLIResult{
		Metric:  problemsIndicatorName,
		Value:   float64(totalProblemCount),
		Success: true,
	}
}

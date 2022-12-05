package dashboard

import (
	"context"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/problems"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	keptn "github.com/keptn/go-utils/pkg/lib"
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
func (p *ProblemTileProcessing) Process(ctx context.Context, tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) []result.SLIWithSLO {
	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	// query the number of open problems based on the management zone filter of the tile
	problemSelector := "status(\"open\")" + tileManagementZoneFilter.ForProblemSelector()
	return []result.SLIWithSLO{p.processOpenProblemTile(ctx, problems.NewQuery(problemSelector, ""))}
}

func (p *ProblemTileProcessing) processOpenProblemTile(ctx context.Context, query problems.Query) result.SLIWithSLO {
	// TODO: 2022-02-14: check: maybe in the future we will allow users to add additional SLO defs via the tile name, e.g. weight or KeySli.
	sloDefinition := keptn.SLO{
		SLI:    problemsIndicatorName,
		Pass:   []*keptn.SLOCriteria{{Criteria: []string{"<=0"}}},
		Weight: 1,
		KeySLI: true,
	}

	request := dynatrace.NewProblemsV2ClientQueryRequest(query, p.timeframe)
	totalProblemCount, err := dynatrace.NewProblemsV2Client(p.client).GetTotalCountByQuery(ctx, request)
	if err != nil {
		return result.NewFailedSLIWithSLOAndQuery(sloDefinition, request.RequestString(), "error querying Problems API v2: "+err.Error())
	}

	return result.NewSuccessfulSLIWithSLOAndQuery(sloDefinition, float64(totalProblemCount), request.RequestString())
}

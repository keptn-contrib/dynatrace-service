package dashboard

import (
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/secpv2"
	v1secpv2 "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/secpv2"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type SecurityProblemTileProcessing struct {
	client    dynatrace.ClientInterface
	startUnix time.Time
	endUnix   time.Time
}

func NewSecurityProblemTileProcessing(client dynatrace.ClientInterface, startUnix time.Time, endUnix time.Time) *SecurityProblemTileProcessing {
	return &SecurityProblemTileProcessing{
		client:    client,
		startUnix: startUnix,
		endUnix:   endUnix,
	}
}

func (p *SecurityProblemTileProcessing) Process(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) *TileResult {

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	// query the number of open security problems based on the management zone filter of the tile
	securityProblemSelector := "status(OPEN)" + tileManagementZoneFilter.ForProblemSelector()
	tileResult, err := p.processProblemSelector(secpv2.NewQuery(securityProblemSelector), p.startUnix, p.endUnix)
	if err != nil {
		log.WithError(err).Error("Error Processing OPEN_SECURITY_PROBLEMS")
		return nil
	}

	return tileResult
}

// processProblemSelector Processes an Open Problem Tile and queries the number of open problems. The current default is that there is a pass criteria of <= 0 as we dont allow problems
// If successful returns sliResult, sliIndicatorName, sliQuery & sloDefinition
func (p *SecurityProblemTileProcessing) processProblemSelector(query secpv2.Query, startUnix time.Time, endUnix time.Time) (*TileResult, error) {
	sliResult := p.getSecurityProblemCountAsSLIResult(query, startUnix, endUnix)

	log.WithFields(
		log.Fields{
			"indicatorName": sliResult.Metric,
			"value":         sliResult.Value,
		}).Debug("Adding SLO to sloResult")

	sloDefinition := &keptn.SLO{
		SLI:    sliResult.Metric,
		Pass:   []*keptn.SLOCriteria{{Criteria: []string{"<=0"}}},
		Weight: 1,
		KeySLI: true,
	}

	return &TileResult{
		sliResult: &sliResult,
		objective: sloDefinition,
		sliName:   sliResult.Metric,
		sliQuery:  v1secpv2.NewQueryProducer(query).Produce(),
	}, nil
}

func (p *SecurityProblemTileProcessing) getSecurityProblemCountAsSLIResult(query secpv2.Query, startUnix time.Time, endUnix time.Time) keptnv2.SLIResult {
	indicatorName := "security_problems"

	totalSecurityProblemCount, err := dynatrace.NewSecurityProblemsClient(p.client).GetTotalCountByQuery(dynatrace.NewSecurityProblemsV2ClientQueryParameters(query, startUnix, endUnix))
	if err != nil {
		return keptnv2.SLIResult{
			Metric:  indicatorName,
			Success: false,
			Message: err.Error(),
		}
	}

	return keptnv2.SLIResult{
		Metric:  indicatorName,
		Value:   float64(totalSecurityProblemCount),
		Success: true,
	}
}

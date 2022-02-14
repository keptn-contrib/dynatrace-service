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

const securityProblemsIndicatorName = "security_problems"

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

// Process retrieves the open security problem count and returns this as a TileResult.
// An SLO definition with a pass criteria of <= 0 is also included as we don't allow security problems
func (p *SecurityProblemTileProcessing) Process(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) *TileResult {

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	// query the number of open security problems based on the management zone filter of the tile
	securityProblemSelector := "status(OPEN)" + tileManagementZoneFilter.ForProblemSelector()
	return p.processSecurityProblemSelector(secpv2.NewQuery(securityProblemSelector))
}

func (p *SecurityProblemTileProcessing) processSecurityProblemSelector(query secpv2.Query) *TileResult {
	sliResult := p.getSecurityProblemCountAsSLIResult(query)

	log.WithFields(
		log.Fields{
			"indicatorName": sliResult.Metric,
			"value":         sliResult.Value,
		}).Debug("Adding SLO to sloResult")

	sloDefinition := &keptn.SLO{
		SLI:    securityProblemsIndicatorName,
		Pass:   []*keptn.SLOCriteria{{Criteria: []string{"<=0"}}},
		Weight: 1,
		KeySLI: true,
	}

	return &TileResult{
		sliResult: &sliResult,
		objective: sloDefinition,
		sliName:   securityProblemsIndicatorName,
		sliQuery:  v1secpv2.NewQueryProducer(query).Produce(),
	}
}

func (p *SecurityProblemTileProcessing) getSecurityProblemCountAsSLIResult(query secpv2.Query) keptnv2.SLIResult {
	totalSecurityProblemCount, err := dynatrace.NewSecurityProblemsClient(p.client).GetTotalCountByQuery(dynatrace.NewSecurityProblemsV2ClientQueryParameters(query, p.startUnix, p.endUnix))
	if err != nil {
		return keptnv2.SLIResult{
			Metric:  securityProblemsIndicatorName,
			Success: false,
			Message: err.Error(),
		}
	}

	return keptnv2.SLIResult{
		Metric:  securityProblemsIndicatorName,
		Value:   float64(totalSecurityProblemCount),
		Success: true,
	}
}

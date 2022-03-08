package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/secpv2"
	v1secpv2 "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/secpv2"
	keptn "github.com/keptn/go-utils/pkg/lib"
	log "github.com/sirupsen/logrus"
)

const securityProblemsIndicatorName = "security_problems"

// SecurityProblemTileProcessing represents the processing of a problems dashboard tile for security problems .
type SecurityProblemTileProcessing struct {
	client    dynatrace.ClientInterface
	timeframe common.Timeframe
}

// NewSecurityProblemTileProcessing creates a new SecurityProblemTileProcessing.
func NewSecurityProblemTileProcessing(client dynatrace.ClientInterface, timeframe common.Timeframe) *SecurityProblemTileProcessing {
	return &SecurityProblemTileProcessing{
		client:    client,
		timeframe: timeframe,
	}
}

// Process retrieves the open security problem count and returns this as a TileResult.
// An SLO definition with a pass criteria of <= 0 is also included as we don't allow security problems
func (p *SecurityProblemTileProcessing) Process(tile *dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) *TileResult {

	// get the tile specific management zone filter that might be needed by different tile processors
	// Check for tile management zone filter - this would overwrite the dashboardManagementZoneFilter
	tileManagementZoneFilter := NewManagementZoneFilter(dashboardFilter, tile.TileFilter.ManagementZone)

	// query the number of open security problems based on the management zone filter of the tile
	securityProblemSelector := "status(\"open\")" + tileManagementZoneFilter.ForProblemSelector()
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
		sliResult: sliResult,
		objective: sloDefinition,
		sliName:   securityProblemsIndicatorName,
		sliQuery:  v1secpv2.NewQueryProducer(query).Produce(),
	}
}

func (p *SecurityProblemTileProcessing) getSecurityProblemCountAsSLIResult(query secpv2.Query) result.SLIResult {
	totalSecurityProblemCount, err := dynatrace.NewSecurityProblemsClient(p.client).GetTotalCountByQuery(dynatrace.NewSecurityProblemsV2ClientQueryParameters(query, p.timeframe))
	if err != nil {
		return result.NewFailedSLIResult(securityProblemsIndicatorName, "error querying Security problems API: "+err.Error())
	}

	return result.NewSuccessfulSLIResult(securityProblemsIndicatorName, float64(totalSecurityProblemCount))
}

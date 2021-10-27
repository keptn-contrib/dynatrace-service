package dashboard

import (
	"time"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

func createDefaultSLOScore() keptncommon.SLOScore {
	return keptncommon.SLOScore{
		Pass:    "90%",
		Warning: "75%",
	}
}

func createDefaultSLOComparison() keptncommon.SLOComparison {
	return keptncommon.SLOComparison{
		CompareWith:               "single_result",
		IncludeResultWithScore:    "pass",
		NumberOfComparisonResults: 1,
		AggregateFunction:         "avg",
	}
}

// Processing will process a Dynatrace dashboard
type Processing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	startUnix     time.Time
	endUnix       time.Time
}

// NewProcessing will create a new Processing
func NewProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, startUnix time.Time, endUnix time.Time) *Processing {
	return &Processing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		startUnix:     startUnix,
		endUnix:       endUnix,
	}
}

// Process will process a dynatrace.Dashboard
func (p *Processing) Process(dashboard *dynatrace.Dashboard) *QueryResult {

	// lets also generate the dashboard link for that timeframe (gtf=c_START_END) as well as management zone (gf=MZID) to pass back as label to Keptn
	dashboardLinkAsLabel := NewLink(p.client.Credentials().GetTenant(), p.startUnix, p.endUnix, dashboard.ID, dashboard.GetFilter())

	totalScore := createDefaultSLOScore()
	comparison := createDefaultSLOComparison()

	// generate our own SLIResult array based on the dashboard configuration
	result := &QueryResult{
		dashboardLink: dashboardLinkAsLabel,
		dashboard:     dashboard,
		sli: &dynatrace.SLI{
			SpecVersion: "0.1.4",
			Indicators:  make(map[string]string),
		},
		slo: &keptncommon.ServiceLevelObjectives{
			Objectives: []*keptncommon.SLO{},
			TotalScore: &totalScore,
			Comparison: &comparison,
		},
		sliResults: []*keptnv2.SLIResult{},
	}

	log.Debug("Dashboard has changed: reparsing it!")

	// now lets iterate through the dashboard to find our SLIs
	for _, tile := range dashboard.Tiles {
		switch tile.TileType {
		case "MARKDOWN":
			score, comparison := NewMarkdownTileProcessing().Process(&tile, createDefaultSLOScore(), createDefaultSLOComparison())
			if score != nil && comparison != nil {
				result.slo.TotalScore = score
				result.slo.Comparison = comparison
			}
		case "SLO":
			tileResults := NewSLOTileProcessing(p.client, p.startUnix, p.endUnix).Process(&tile)
			result.addTileResults(tileResults)
		case "OPEN_PROBLEMS":
			tileResult := NewProblemTileProcessing(p.client, p.startUnix, p.endUnix).Process(&tile, dashboard.GetFilter())
			result.addTileResult(tileResult)

			// current logic also does security tile processing for open problem tiles
			tileResult = NewSecurityProblemTileProcessing(p.client, p.startUnix, p.endUnix).Process(&tile, dashboard.GetFilter())
			result.addTileResult(tileResult)
		case "DATA_EXPLORER":
			// here we handle the new Metric Data Explorer Tile
			tileResults := NewDataExplorerTileProcessing(p.client, p.eventData, p.customFilters, p.startUnix, p.endUnix).Process(&tile, dashboard.GetFilter())
			result.addTileResults(tileResults)
		case "CUSTOM_CHARTING":
			tileResults := NewCustomChartingTileProcessing(p.client, p.eventData, p.customFilters, p.startUnix, p.endUnix).Process(&tile, dashboard.GetFilter())
			result.addTileResults(tileResults)
		case "DTAQL":
			tileResults := NewUSQLTileProcessing(p.client, p.eventData, p.customFilters, p.startUnix, p.endUnix).Process(&tile)
			result.addTileResults(tileResults)
		default:
			// we do not do markdowns (HEADER) or synthetic tests (SYNTHETIC_TESTS)
			continue
		}
	}

	return result
}

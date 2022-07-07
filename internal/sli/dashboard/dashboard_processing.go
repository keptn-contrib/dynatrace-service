package dashboard

import (
	"context"
	"fmt"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
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
	timeframe     common.Timeframe
}

// NewProcessing will create a new Processing
func NewProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe) *Processing {
	return &Processing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
	}
}

// Process processes a dynatrace.Dashboard.
func (p *Processing) Process(ctx context.Context, dashboard *dynatrace.Dashboard) (*QueryResult, error) {

	// lets also generate the dashboard link for that timeframe (gtf=c_START_END) as well as management zone (gf=MZID) to pass back as label to Keptn
	dashboardLinkAsLabel := NewLink(p.client.Credentials().GetTenant(), p.timeframe, dashboard.ID, dashboard.GetFilter())

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
	}

	log.Debug("Dashboard will be parsed!")

	// now let's iterate through the dashboard to find our SLIs
	markdownAlreadyProcessed := false
	for _, tile := range dashboard.Tiles {
		switch tile.TileType {
		case dynatrace.MarkdownTileType:
			res, err := NewMarkdownTileProcessing().Process(&tile, createDefaultSLOScore(), createDefaultSLOComparison())
			if err != nil {
				return nil, fmt.Errorf("markdown tile parsing error: %w", err)
			}
			if res != nil {
				if markdownAlreadyProcessed {
					return nil, fmt.Errorf("only one markdown tile allowed for KQG configuration")
				}
				result.slo.TotalScore = &res.totalScore
				result.slo.Comparison = &res.comparison
				markdownAlreadyProcessed = true
			}
		case dynatrace.SLOTileType:
			tileResult := NewSLOTileProcessing(p.client, p.timeframe).Process(ctx, &tile)
			result.addTileResult(tileResult)
		case dynatrace.OpenProblemsTileType:
			tileResult := NewProblemTileProcessing(p.client, p.timeframe).Process(ctx, &tile, dashboard.GetFilter())
			result.addTileResult(tileResult)
		case dynatrace.DataExplorerTileType:
			tileResults := NewDataExplorerTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile, dashboard.GetFilter())
			result.addTileResults(tileResults)
		case dynatrace.CustomChartingTileType:
			tileResults := NewCustomChartingTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile, dashboard.GetFilter())
			result.addTileResults(tileResults)
		case dynatrace.USQLTileType:
			tileResults := NewUSQLTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile)
			result.addTileResults(tileResults)
		default:
			// we do not do markdowns (HEADER) or synthetic tests (SYNTHETIC_TESTS)
			continue
		}
	}

	return result, nil
}

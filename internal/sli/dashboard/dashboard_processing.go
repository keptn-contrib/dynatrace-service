package dashboard

import (
	"context"
	"fmt"
	"strings"

	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
)

// sloUploaderInterface can write SLOs.
type sloUploaderInterface interface {

	// UploadSLOs uploads the SLOs for the specified project, stage and service.
	UploadSLOs(ctx context.Context, project string, stage string, service string, slos *keptncommon.ServiceLevelObjectives) error
}

type duplicateSLINameChecker struct {
	nameCounts map[string]int
}

func newDuplicateSLINameChecker(results []result.SLIWithSLO) duplicateSLINameChecker {
	nameCounts := make(map[string]int, len(results))
	for _, r := range results {
		name := r.SLIResult().Metric
		nameCounts[name] = nameCounts[name] + 1
	}

	return duplicateSLINameChecker{
		nameCounts: nameCounts,
	}
}

func (c *duplicateSLINameChecker) hasDuplicateName(r result.SLIWithSLO) bool {
	return c.nameCounts[r.SLIResult().Metric] > 1
}

type duplicateDisplayNameChecker struct {
	displayNameCounts map[string]int
}

func newDuplicateDisplayNameChecker(results []result.SLIWithSLO) duplicateDisplayNameChecker {
	displayNameCounts := make(map[string]int, len(results))
	for _, r := range results {
		displayName := r.SLODefinition().DisplayName
		if displayName == "" {
			continue
		}

		displayNameCounts[displayName] = displayNameCounts[displayName] + 1
	}

	return duplicateDisplayNameChecker{
		displayNameCounts: displayNameCounts,
	}
}

func (c *duplicateDisplayNameChecker) hasDuplicateDisplayName(t result.SLIWithSLO) bool {
	displayName := t.SLODefinition().DisplayName
	if displayName == "" {
		return false
	}

	return c.displayNameCounts[displayName] > 1
}

// Processing will process a Dynatrace dashboard
type Processing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	timeframe     common.Timeframe
	sloUploader   sloUploaderInterface
}

// NewProcessing will create a new Processing
func NewProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, timeframe common.Timeframe, sloUploader sloUploaderInterface) *Processing {
	return &Processing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		timeframe:     timeframe,
		sloUploader:   sloUploader,
	}
}

// Process processes a dynatrace.Dashboard.
func (p *Processing) Process(ctx context.Context, dashboard *dynatrace.Dashboard) ([]result.SLIWithSLO, error) {
	processingResult, err := p.process(ctx, dashboard)
	if err != nil {
		return nil, NewProcessingError(err)
	}

	return processingResult, nil
}

func (p *Processing) process(ctx context.Context, dashboard *dynatrace.Dashboard) ([]result.SLIWithSLO, error) {
	log.Debug("Dashboard will be parsed!")

	totalScore := common.CreateDefaultSLOScore()
	comparison := common.CreateDefaultSLOComparison()
	results := []result.SLIWithSLO{}

	// now let's iterate through the dashboard to find our SLIs
	markdownAlreadyProcessed := false
	for _, tile := range dashboard.Tiles {
		switch tile.TileType {
		case dynatrace.MarkdownTileType:
			res, err := NewMarkdownTileProcessing().TryProcess(&tile)
			if err != nil {
				return nil, fmt.Errorf("markdown tile parsing error: %w", err)
			}
			if res != nil {
				if markdownAlreadyProcessed {
					return nil, fmt.Errorf("only one markdown tile allowed for KQG configuration")
				}

				totalScore = res.totalScore
				comparison = res.comparison
				markdownAlreadyProcessed = true
			}

		default:
			results = append(results, p.processTile(ctx, tile, dashboard.GetFilter())...)
		}
	}

	err := p.createAndUploadSLOs(ctx, results, totalScore, comparison)
	if err != nil {
		return nil, NewUploadSLOsError(err)
	}

	return checkForDuplicatesInResults(results), nil
}

func (p *Processing) processTile(ctx context.Context, tile dynatrace.Tile, dashboardFilter *dynatrace.DashboardFilter) []result.SLIWithSLO {
	switch tile.TileType {
	case dynatrace.SLOTileType:
		return NewSLOTileProcessing(p.client, p.timeframe).Process(ctx, &tile)
	case dynatrace.OpenProblemsTileType:
		return NewProblemTileProcessing(p.client, p.timeframe).Process(ctx, &tile, dashboardFilter)
	case dynatrace.DataExplorerTileType:
		return NewDataExplorerTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile, dashboardFilter)
	case dynatrace.CustomChartingTileType:
		return NewCustomChartingTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile, dashboardFilter)
	case dynatrace.USQLTileType:
		return NewUSQLTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile)
	default:
		// we do not do markdowns (HEADER) or synthetic tests (SYNTHETIC_TESTS)
		return nil
	}
}

func (p *Processing) createAndUploadSLOs(ctx context.Context, results []result.SLIWithSLO, totalScore keptncommon.SLOScore, comparison keptncommon.SLOComparison) error {
	objectives := make([]*keptncommon.SLO, 0, len(results))
	for _, r := range results {
		sloDefinition := r.SLODefinition()
		objectives = append(objectives, &sloDefinition)
	}

	slos := keptncommon.ServiceLevelObjectives{
		Objectives: objectives,
		TotalScore: &totalScore,
		Comparison: &comparison,
	}

	return p.sloUploader.UploadSLOs(ctx, p.eventData.GetProject(), p.eventData.GetStage(), p.eventData.GetService(), &slos)
}

func checkForDuplicatesInResults(results []result.SLIWithSLO) []result.SLIWithSLO {
	sliNameChecker := newDuplicateSLINameChecker(results)
	displayNameChecker := newDuplicateDisplayNameChecker(results)
	checkedResults := make([]result.SLIWithSLO, 0, len(results))
	for _, r := range results {
		if sliNameChecker.hasDuplicateName(r) && displayNameChecker.hasDuplicateDisplayName(r) {
			r = addErrorAndFailResult(r, "duplicate SLI and display name")
		} else if sliNameChecker.hasDuplicateName(r) {
			r = addErrorAndFailResult(r, "duplicate SLI name")
		} else if displayNameChecker.hasDuplicateDisplayName(r) {
			r = addErrorAndFailResult(r, "duplicate display name")
		}

		checkedResults = append(checkedResults, r)
	}
	return checkedResults
}

func addErrorAndFailResult(r result.SLIWithSLO, message string) result.SLIWithSLO {
	return result.NewFailedSLIWithSLOAndQuery(
		r.SLODefinition(),
		r.SLIResult().Query,
		strings.Join([]string{message, r.SLIResult().Message}, "; "),
	)
}

package dashboard

import (
	"context"
	"fmt"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
)

type processingResultBuilder struct {
	slo        *keptnapi.ServiceLevelObjectives
	sliResults []result.SLIResult
}

func newProcessingResultBuilder() *processingResultBuilder {
	totalScore := common.CreateDefaultSLOScore()
	comparison := common.CreateDefaultSLOComparison()
	return &processingResultBuilder{
		slo: &keptncommon.ServiceLevelObjectives{
			Objectives: []*keptncommon.SLO{},
			TotalScore: &totalScore,
			Comparison: &comparison,
		},
	}
}

func (b *processingResultBuilder) applyMarkdownParsingResult(r *markdownParsingResult) {
	b.slo.TotalScore = &r.totalScore
	b.slo.Comparison = &r.comparison
}

// addTileResult adds a TileResult to the processingResultBuilder
func (b *processingResultBuilder) addTileResult(result TileResult) {
	if result.sloDefinition != nil {
		b.slo.Objectives = append(b.slo.Objectives, result.sloDefinition)
	}

	b.sliResults = append(b.sliResults, result.sliResult)
}

// addTileResult adds multiple TileResult to the processingResultBuilder,
func (b *processingResultBuilder) addTileResults(results []TileResult) {
	for _, result := range results {
		b.addTileResult(result)
	}
}

func (b *processingResultBuilder) build() *ProcessingResult {
	return &ProcessingResult{
		slo:        b.slo,
		sliResults: b.sliResults,
	}
}

type ProcessingResult struct {
	slo        *keptnapi.ServiceLevelObjectives
	sliResults []result.SLIResult
}

// SLOs gets the SLOs.
func (r *ProcessingResult) SLOs() *keptnapi.ServiceLevelObjectives {
	return r.slo
}

// HasSLOs checks whether any objectives are available
func (r *ProcessingResult) HasSLOs() bool {
	return r.slo != nil && len(r.slo.Objectives) > 0
}

// SLIResults gets the SLI results.
func (r *ProcessingResult) SLIResults() []result.SLIResult {
	return r.sliResults
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
func (p *Processing) Process(ctx context.Context, dashboard *dynatrace.Dashboard) (*ProcessingResult, error) {

	resultBuilder := newProcessingResultBuilder()
	log.Debug("Dashboard will be parsed!")

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
				resultBuilder.applyMarkdownParsingResult(res)
				markdownAlreadyProcessed = true
			}
		case dynatrace.SLOTileType:
			resultBuilder.addTileResults(NewSLOTileProcessing(p.client, p.timeframe).Process(ctx, &tile))
		case dynatrace.OpenProblemsTileType:
			resultBuilder.addTileResults(NewProblemTileProcessing(p.client, p.timeframe).Process(ctx, &tile, dashboard.GetFilter()))
		case dynatrace.DataExplorerTileType:
			resultBuilder.addTileResults(NewDataExplorerTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile, dashboard.GetFilter()))
		case dynatrace.CustomChartingTileType:
			resultBuilder.addTileResults(NewCustomChartingTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile, dashboard.GetFilter()))
		case dynatrace.USQLTileType:
			resultBuilder.addTileResults(NewUSQLTileProcessing(p.client, p.eventData, p.customFilters, p.timeframe).Process(ctx, &tile))
		default:
			// we do not do markdowns (HEADER) or synthetic tests (SYNTHETIC_TESTS)
			continue
		}
	}

	return resultBuilder.build(), nil
}

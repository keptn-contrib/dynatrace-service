package dashboard

import (
	"context"
	"fmt"
	"strings"

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
	totalScore  keptncommon.SLOScore
	comparison  keptncommon.SLOComparison
	tileResults []TileResult
}

func newProcessingResultBuilder() *processingResultBuilder {
	return &processingResultBuilder{
		totalScore: common.CreateDefaultSLOScore(),
		comparison: common.CreateDefaultSLOComparison(),
	}
}

type duplicateSLINameChecker struct {
	nameCounts map[string]int
}

func newDuplicateSLINameChecker(results []TileResult) duplicateSLINameChecker {
	nameCounts := make(map[string]int, len(results))
	for _, result := range results {
		name := result.sliResult.Metric
		nameCounts[name] = nameCounts[name] + 1
	}

	return duplicateSLINameChecker{
		nameCounts: nameCounts,
	}
}

func (c *duplicateSLINameChecker) hasDuplicateName(sliResult result.SLIResult) bool {
	return c.nameCounts[sliResult.Metric] > 1
}

type duplicateDisplayNameChecker struct {
	displayNameCounts map[string]int
}

func newDuplicateDisplayNameChecker(results []TileResult) duplicateDisplayNameChecker {
	displayNameCounts := make(map[string]int, len(results))
	for _, result := range results {
		if result.sloDefinition == nil {
			continue
		}

		displayName := result.sloDefinition.DisplayName
		if displayName == "" {
			continue
		}

		displayNameCounts[displayName] = displayNameCounts[displayName] + 1
	}

	return duplicateDisplayNameChecker{
		displayNameCounts: displayNameCounts,
	}
}

func (c *duplicateDisplayNameChecker) hasDuplicateDisplayName(t TileResult) bool {
	if t.sloDefinition == nil {
		return false
	}

	displayName := t.sloDefinition.DisplayName
	if displayName == "" {
		return false
	}

	return c.displayNameCounts[displayName] > 1
}

func (b *processingResultBuilder) applyMarkdownParsingResult(r *markdownParsingResult) {
	b.totalScore = r.totalScore
	b.comparison = r.comparison
}

// addTileResult adds multiple TileResult to the processingResultBuilder,
func (b *processingResultBuilder) addTileResults(results []TileResult) {
	for _, result := range results {
		b.tileResults = append(b.tileResults, result)
	}
}

func (b *processingResultBuilder) build() *ProcessingResult {
	objectives := make([]*keptncommon.SLO, 0, len(b.tileResults))
	sliResults := make([]result.SLIResult, 0, len(b.tileResults))

	sliNameChecker := newDuplicateSLINameChecker(b.tileResults)
	displayNameChecker := newDuplicateDisplayNameChecker(b.tileResults)
	for _, tileResult := range b.tileResults {
		sliResult := tileResult.sliResult

		if sliNameChecker.hasDuplicateName(sliResult) && displayNameChecker.hasDuplicateDisplayName(tileResult) {
			sliResult = addErrorAndFailResult(sliResult, "duplicate SLI and display name")
		} else if sliNameChecker.hasDuplicateName(sliResult) {
			sliResult = addErrorAndFailResult(sliResult, "duplicate SLI name")
		} else if displayNameChecker.hasDuplicateDisplayName(tileResult) {
			sliResult = addErrorAndFailResult(sliResult, "duplicate display name")
		}

		if tileResult.sloDefinition != nil {
			objectives = append(objectives, tileResult.sloDefinition)
		}
		sliResults = append(sliResults, sliResult)
	}

	return NewProcessingResult(
		&keptncommon.ServiceLevelObjectives{
			Objectives: objectives,
			TotalScore: &b.totalScore,
			Comparison: &b.comparison,
		},
		sliResults)
}

func addErrorAndFailResult(sliResult result.SLIResult, message string) result.SLIResult {
	sliResult.Success = false
	sliResult.IndicatorResult = result.IndicatorResultFailed
	sliResult.Value = 0
	sliResult.Message = strings.Join([]string{message, sliResult.Message}, "; ")
	return sliResult
}

// ProcessingResult contains the result of processing a dashboard.
type ProcessingResult struct {
	slo        *keptnapi.ServiceLevelObjectives
	sliResults []result.SLIResult
}

// NewProcessingResult creates a new ProcessingResult.
func NewProcessingResult(slo *keptnapi.ServiceLevelObjectives, sliResults []result.SLIResult) *ProcessingResult {
	return &ProcessingResult{
		slo:        slo,
		sliResults: sliResults,
	}
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

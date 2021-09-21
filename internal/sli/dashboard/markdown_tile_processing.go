package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	"strconv"
	"strings"
)

type MarkdownTileProcessing struct {
}

// NewMarkdownTileProcessing will create a new MarkdownTileProcessing
func NewMarkdownTileProcessing() *MarkdownTileProcessing {
	return &MarkdownTileProcessing{}
}

// Process will overwrite the default values for SLOScore and SLOComparison with the contents found in the markdown
func (p *MarkdownTileProcessing) Process(tile *dynatrace.Tile, defaultScore keptncommon.SLOScore, defaultComparison keptncommon.SLOComparison) (*keptncommon.SLOScore, *keptncommon.SLOComparison) {
	// we allow the user to use a markdown to specify SLI/SLO properties, e.g: KQG.Total.Pass
	// if we find KQG. we process the markdown
	return parseMarkdownConfiguration(tile.Markdown, defaultScore, defaultComparison)
}

// parseMarkdownConfiguration parses a text that can be used in a Markdown tile to specify global SLO properties
func parseMarkdownConfiguration(markdown string, totalScore keptncommon.SLOScore, comparison keptncommon.SLOComparison) (*keptncommon.SLOScore, *keptncommon.SLOComparison) {
	if !strings.Contains(markdown, "KQG.") {
		return nil, nil
	}

	markdownSplits := strings.Split(markdown, ";")
	for _, markdownSplitValue := range markdownSplits {
		configValueSplits := strings.Split(markdownSplitValue, "=")
		if len(configValueSplits) != 2 {
			continue
		}

		// lets get configname and value
		configName := strings.ToLower(configValueSplits[0])
		configValue := configValueSplits[1]

		switch configName {
		case "kqg.total.pass":
			totalScore.Pass = configValue
		case "kqg.total.warning":
			totalScore.Warning = configValue
		case "kqg.compare.withscore":
			comparison.IncludeResultWithScore = configValue
			if (configValue == "pass") || (configValue == "pass_or_warn") || (configValue == "all") {
				comparison.IncludeResultWithScore = configValue
			} else {
				comparison.IncludeResultWithScore = "pass"
			}
		case "kqg.compare.results":
			noresults, err := strconv.Atoi(configValue)
			if err != nil {
				comparison.NumberOfComparisonResults = 1
			} else {
				comparison.NumberOfComparisonResults = noresults
			}
			if comparison.NumberOfComparisonResults > 1 {
				comparison.CompareWith = "several_results"
			} else {
				comparison.CompareWith = "single_result"
			}
		case "kqg.compare.function":
			if (configValue == "avg") || (configValue == "p50") || (configValue == "p90") || (configValue == "p95") {
				comparison.AggregateFunction = configValue
			} else {
				comparison.AggregateFunction = "avg"
			}
		}
	}

	return &totalScore, &comparison
}

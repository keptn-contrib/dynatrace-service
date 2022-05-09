package dashboard

import (
	"fmt"
	"strconv"
	"strings"

	keptncommon "github.com/keptn/go-utils/pkg/lib"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
)

type markdownParsingErrors struct {
	errors []error
}

func (err *markdownParsingErrors) Error() string {
	var errStrings = make([]string, len(err.errors))
	for i, e := range err.errors {
		errStrings[i] = e.Error()
	}
	return strings.Join(errStrings, ";")
}

type invalidValueError struct {
	key   string
	value string
}

func (err *invalidValueError) Error() string {
	return fmt.Sprintf("invalid value for '%s': %s", err.key, err.value)
}

type markdownParsingResult struct {
	totalScore keptncommon.SLOScore
	comparison keptncommon.SLOComparison
}

type MarkdownTileProcessing struct {
}

// NewMarkdownTileProcessing will create a new MarkdownTileProcessing
func NewMarkdownTileProcessing() *MarkdownTileProcessing {
	return &MarkdownTileProcessing{}
}

// Process will overwrite the default values for SLOScore and SLOComparison with the contents found in the markdown
func (p *MarkdownTileProcessing) Process(tile *dynatrace.Tile, defaultScore keptncommon.SLOScore, defaultComparison keptncommon.SLOComparison) (*markdownParsingResult, error) {
	// we allow the user to use a markdown to specify SLI/SLO properties, e.g: KQG.Total.Pass
	// if we find KQG. we process the markdown
	return parseMarkdownConfiguration(tile.Markdown, defaultScore, defaultComparison)
}

const (
	totalPass                  = "kqg.total.pass"
	totalWarning               = "kqg.total.warning"
	compareWithScore           = "kqg.compare.withscore"
	compareWithScoreAll        = "all"
	compareWithScorePass       = "pass"
	compareWithScorePassOrWarn = "pass_or_warn"
	compareResults             = "kqg.compare.results" // this is an int! You cannot specify 'single_result' or 'several results' at all -> it is derived from the number of results
	compareResultsSingle       = "single_result"
	compareResultsMultiple     = "several_results"
	compareFunction            = "kqg.compare.function"
	compareFunctionAvg         = "avg"
	compareFunctionP50         = "p50"
	compareFunctionP90         = "p90"
	compareFunctionP95         = "p95"
)

// parseMarkdownConfiguration parses a text that can be used in a Markdown tile to specify global SLO properties
func parseMarkdownConfiguration(markdown string, totalScore keptncommon.SLOScore, comparison keptncommon.SLOComparison) (*markdownParsingResult, error) {
	if !strings.Contains(markdown, "KQG.") {
		return nil, nil
	}

	result := &markdownParsingResult{
		totalScore: totalScore,
		comparison: comparison,
	}

	var errs []error

	markdownSplits := strings.Split(markdown, ";")
	for _, markdownSplitValue := range markdownSplits {
		configValueSplits := strings.Split(markdownSplitValue, "=")
		if len(configValueSplits) != 2 {
			continue
		}

		// lets separate key and value
		key := strings.ToLower(configValueSplits[0])
		value := configValueSplits[1]

		switch key {
		case totalPass:
			result.totalScore.Pass = value
		case totalWarning:
			result.totalScore.Warning = value
		case compareWithScore:
			score, err := parseCompareWithScore(value)
			if err != nil {
				errs = append(errs, err)
			}
			result.comparison.IncludeResultWithScore = score
		case compareResults:
			numberOfResults, err := parseCompareNumberOfResults(value)
			if err != nil {
				errs = append(errs, err)
			}
			result.comparison.NumberOfComparisonResults = numberOfResults
		case compareFunction:
			aggregateFunc, err := parseAggregateFunction(value)
			if err != nil {
				errs = append(errs, err)
			}
			result.comparison.AggregateFunction = aggregateFunc
		}
	}

	if len(errs) > 0 {
		return nil, &markdownParsingErrors{
			errors: errs,
		}
	}

	result.comparison.CompareWith = compareResultsSingle
	if result.comparison.NumberOfComparisonResults > 1 {
		result.comparison.CompareWith = compareResultsMultiple
	}

	return result, nil
}

func parseCompareWithScore(value string) (string, error) {
	switch value {
	case compareWithScorePass, compareWithScoreAll, compareWithScorePassOrWarn:
		return value, nil
	}

	return "", &invalidValueError{key: compareWithScore, value: value}
}

func parseCompareNumberOfResults(value string) (int, error) {
	numberOfResults, err := strconv.Atoi(value)
	if err != nil {
		return 0, &invalidValueError{key: compareResults, value: value}
	}

	if numberOfResults < 1 {
		return 0, &invalidValueError{key: compareResults, value: value}
	}

	return numberOfResults, nil
}

func parseAggregateFunction(value string) (string, error) {
	switch value {
	case compareFunctionAvg, compareFunctionP50, compareFunctionP90, compareFunctionP95:
		return value, nil
	}

	return "", &invalidValueError{key: compareFunction, value: value}
}

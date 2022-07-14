package dashboard

import (
	"fmt"
	"regexp"
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
	TotalPass                  = "kqg.total.pass"
	TotalWarning               = "kqg.total.warning"
	CompareWithScore           = "kqg.compare.withscore"
	CompareWithScoreAll        = "all"
	CompareWithScorePass       = "pass"
	CompareWithScorePassOrWarn = "pass_or_warn"
	CompareResults             = "kqg.compare.results"
	CompareResultsSingle       = "single_result"
	CompareResultsMultiple     = "several_results"
	CompareFunction            = "kqg.compare.function"
	CompareFunctionAvg         = "avg"
	CompareFunctionP50         = "p50"
	CompareFunctionP90         = "p90"
	CompareFunctionP95         = "p95"
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
	keyFound := make(map[string]bool)

	for _, kv := range newKeyValueParsing(markdown).parse() {
		if !kv.split {
			continue
		}

		switch strings.ToLower(kv.key) {
		case TotalPass:
			if keyFound[TotalPass] {
				errs = append(errs, &duplicateKeyError{key: TotalPass})
				break
			}
			if isNotAPercentValue(kv.value) {
				errs = append(errs, &invalidValueError{key: TotalPass, value: kv.value})
			}
			result.totalScore.Pass = kv.value
			keyFound[TotalPass] = true
		case TotalWarning:
			if keyFound[TotalWarning] {
				errs = append(errs, &duplicateKeyError{key: TotalWarning})
				break
			}
			if isNotAPercentValue(kv.value) {
				errs = append(errs, &invalidValueError{key: TotalWarning, value: kv.value})
			}
			result.totalScore.Warning = kv.value
			keyFound[TotalWarning] = true
		case CompareWithScore:
			if keyFound[CompareWithScore] {
				errs = append(errs, &duplicateKeyError{key: CompareWithScore})
				break
			}
			score, err := parseCompareWithScore(kv.value)
			if err != nil {
				errs = append(errs, err)
			}
			result.comparison.IncludeResultWithScore = score
			keyFound[CompareWithScore] = true
		case CompareResults:
			if keyFound[CompareResults] {
				errs = append(errs, &duplicateKeyError{key: CompareResults})
				break
			}
			numberOfResults, err := parseCompareNumberOfResults(kv.value)
			if err != nil {
				errs = append(errs, err)
			}
			result.comparison.NumberOfComparisonResults = numberOfResults
			keyFound[CompareResults] = true
		case CompareFunction:
			if keyFound[CompareFunction] {
				errs = append(errs, &duplicateKeyError{key: CompareFunction})
				break
			}
			aggregateFunc, err := parseAggregateFunction(kv.value)
			if err != nil {
				errs = append(errs, err)
			}
			result.comparison.AggregateFunction = aggregateFunc
			keyFound[CompareFunction] = true
		}
	}

	if len(errs) > 0 {
		return nil, &markdownParsingErrors{
			errors: errs,
		}
	}

	result.comparison.CompareWith = CompareResultsSingle
	if result.comparison.NumberOfComparisonResults > 1 {
		result.comparison.CompareWith = CompareResultsMultiple
	}

	return result, nil
}

func isNotAPercentValue(value string) bool {
	pattern := regexp.MustCompile("^(\\d+|\\d+\\.\\d+)([%]?)$")

	return !pattern.MatchString(value)
}

func parseCompareWithScore(value string) (string, error) {
	switch value {
	case CompareWithScorePass, CompareWithScoreAll, CompareWithScorePassOrWarn:
		return value, nil
	}

	return "", &invalidValueError{key: CompareWithScore, value: value}
}

func parseCompareNumberOfResults(value string) (int, error) {
	numberOfResults, err := strconv.Atoi(value)
	if err != nil {
		return 0, &invalidValueError{key: CompareResults, value: value}
	}

	if numberOfResults < 1 {
		return 0, &invalidValueError{key: CompareResults, value: value}
	}

	return numberOfResults, nil
}

func parseAggregateFunction(value string) (string, error) {
	switch value {
	case CompareFunctionAvg, CompareFunctionP50, CompareFunctionP90, CompareFunctionP95:
		return value, nil
	}

	return "", &invalidValueError{key: CompareFunction, value: value}
}

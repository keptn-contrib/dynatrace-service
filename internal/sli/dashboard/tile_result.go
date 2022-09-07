package dashboard

import (
	"fmt"

	keptnapi "github.com/keptn/go-utils/pkg/lib"

	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
)

// TileResult stores the result of processing a dashboard tile and retrieving the SLIResult.
type TileResult struct {
	sliResult     result.SLIResult
	sloDefinition *keptnapi.SLO
}

func newSuccessfulTileResult(sloDefinition keptnapi.SLO, value float64, query string) TileResult {
	return TileResult{
		sliResult:     result.NewSuccessfulSLIResultWithQuery(sloDefinition.SLI, value, query),
		sloDefinition: &sloDefinition,
	}
}

func newFailedTileResult(indicatorName string, message string) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(indicatorName, message),
	}
}

func newFailedTileResultFromError(indicatorName string, message string, err error) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(indicatorName, fmt.Sprintf("%s: %s", message, err)),
	}
}

func newFailedTileResultFromSLODefinition(sloDefinition keptnapi.SLO, message string) TileResult {
	return TileResult{
		sliResult:     result.NewFailedSLIResult(sloDefinition.SLI, message),
		sloDefinition: &sloDefinition,
	}
}

func newWarningTileResultFromSLODefinition(sloDefinition keptnapi.SLO, message string) TileResult {
	return TileResult{
		sliResult:     result.NewWarningSLIResult(sloDefinition.SLI, message),
		sloDefinition: &sloDefinition,
	}
}

func newFailedTileResultFromSLODefinitionAndQuery(sloDefinition keptnapi.SLO, sliQuery string, message string) TileResult {
	return TileResult{
		sliResult:     result.NewFailedSLIResultWithQuery(sloDefinition.SLI, message, sliQuery),
		sloDefinition: &sloDefinition,
	}
}

func newWarningTileResultFromSLODefinitionAndQuery(sloDefinition keptnapi.SLO, sliQuery string, message string) TileResult {
	return TileResult{
		sliResult:     result.NewWarningSLIResultWithQuery(sloDefinition.SLI, message, sliQuery),
		sloDefinition: &sloDefinition,
	}
}

package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
)

// TileResult stores the result of processing a dashboard tile and retrieving the SLIResult.
type TileResult struct {
	sliResult result.SLIResult
	objective *keptnapi.SLO
	sliName   string
	sliQuery  string
}

func newFailedTileResult(indicatorName string, message string) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(indicatorName, message),
		sliName:   indicatorName,
	}
}

func newFailedTileResultFromSLODefinition(sloDefinition *keptnapi.SLO, message string) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(sloDefinition.SLI, message),
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
	}
}

func newWarningTileResultFromSLODefinition(sloDefinition *keptnapi.SLO, message string) TileResult {
	return TileResult{
		sliResult: result.NewWarningSLIResult(sloDefinition.SLI, message),
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
	}
}

func newFailedTileResultFromSLODefinitionAndSLIQuery(sloDefinition *keptnapi.SLO, sliQuery string, message string) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(sloDefinition.SLI, message),
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
		sliQuery:  sliQuery,
	}
}

func newWarningTileResultFromSLODefinitionAndSLIQuery(sloDefinition *keptnapi.SLO, sliQuery string, message string) TileResult {
	return TileResult{
		sliResult: result.NewWarningSLIResult(sloDefinition.SLI, message),
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
		sliQuery:  sliQuery,
	}
}

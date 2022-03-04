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

func newUnsuccessfulTileResult(indicatorName string, message string) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(indicatorName, message),
		sliName:   indicatorName,
	}
}

func newUnsuccessfulTileResultFromSLODefinition(sloDefinition *keptnapi.SLO, message string) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(sloDefinition.SLI, message),
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
	}
}

func newUnsuccessfulTileResultFromSLODefinitionAndSLIQuery(sloDefinition *keptnapi.SLO, sliQuery string, message string) TileResult {
	return TileResult{
		sliResult: result.NewFailedSLIResult(sloDefinition.SLI, message),
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
		sliQuery:  sliQuery,
	}
}

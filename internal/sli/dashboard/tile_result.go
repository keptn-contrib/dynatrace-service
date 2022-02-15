package dashboard

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type TileResult struct {
	sliResult *keptnv2.SLIResult
	objective *keptnapi.SLO
	sliName   string
	sliQuery  string
}

func newUnsuccessfulTileResult(indicatorName string, message string) TileResult {
	return TileResult{
		sliResult: &keptnv2.SLIResult{
			Metric:  indicatorName,
			Success: false,
			Message: message,
		},
		sliName: indicatorName,
	}
}

func newUnsuccessfulTileResultFromSLODefinition(sloDefinition *keptnapi.SLO, message string) TileResult {
	return TileResult{
		sliResult: &keptnv2.SLIResult{
			Metric:  sloDefinition.SLI,
			Value:   0,
			Success: false,
			Message: message,
		},
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
	}
}

func newUnsuccessfulTileResultFromSLODefinitionAndSLIQuery(sloDefinition *keptnapi.SLO, sliQuery string, message string) TileResult {
	return TileResult{
		sliResult: &keptnv2.SLIResult{
			Metric:  sloDefinition.SLI,
			Value:   0,
			Success: false,
			Message: message,
		},
		objective: sloDefinition,
		sliName:   sloDefinition.SLI,
		sliQuery:  sliQuery,
	}
}

package result

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
)

// SLIWithSLO stores the result of processing a dashboard tile and retrieving the SLIResult.
type SLIWithSLO struct {
	sliResult     SLIResult
	sloDefinition keptnapi.SLO
}

func NewSLIWithSLO(sliResult SLIResult, sloDefinition keptnapi.SLO) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     sliResult,
		sloDefinition: sloDefinition,
	}
}

func NewSuccessfulSLIWithSLO(sloDefinition keptnapi.SLO, value float64) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewSuccessfulSLIResult(sloDefinition.SLI, value),
		sloDefinition: sloDefinition,
	}
}

func NewFailedSLIWithSLO(sloDefinition keptnapi.SLO, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewFailedSLIResult(sloDefinition.SLI, message),
		sloDefinition: sloDefinition,
	}
}

func NewWarningSLIWithSLO(sloDefinition keptnapi.SLO, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewWarningSLIResult(sloDefinition.SLI, message),
		sloDefinition: sloDefinition,
	}
}

func NewSuccessfulSLIWithSLOAndQuery(sloDefinition keptnapi.SLO, value float64, query string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewSuccessfulSLIResultWithQuery(sloDefinition.SLI, value, query),
		sloDefinition: sloDefinition,
	}
}

func NewFailedSLIWithSLOAndQuery(sloDefinition keptnapi.SLO, sliQuery string, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewFailedSLIResultWithQuery(sloDefinition.SLI, message, sliQuery),
		sloDefinition: sloDefinition,
	}
}

func NewWarningSLIWithSLOAndQuery(sloDefinition keptnapi.SLO, sliQuery string, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewWarningSLIResultWithQuery(sloDefinition.SLI, message, sliQuery),
		sloDefinition: sloDefinition,
	}
}

func CreateInformationalSLODefinition(name string) keptnapi.SLO {
	return keptnapi.SLO{
		SLI:    name,
		Weight: 1,
	}
}

func (r *SLIWithSLO) SLIResult() SLIResult {
	return r.sliResult
}

func (r *SLIWithSLO) SLODefinition() keptnapi.SLO {
	return r.sloDefinition
}

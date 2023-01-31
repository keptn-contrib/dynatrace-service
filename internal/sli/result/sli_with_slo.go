package result

// SLIWithSLO stores the result of processing a dashboard tile and retrieving the SLIResult.
type SLIWithSLO struct {
	sliResult     SLIResult
	sloDefinition SLO
}

func NewSLIWithSLO(sliResult SLIResult, sloDefinition SLO) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     sliResult,
		sloDefinition: sloDefinition,
	}
}

func NewFailedSLIWithSLO(sloDefinition SLO, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewFailedSLIResult(sloDefinition.SLI, message),
		sloDefinition: sloDefinition,
	}
}

func NewSuccessfulSLIWithSLOAndQuery(sloDefinition SLO, value float64, query string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewSuccessfulSLIResultWithQuery(sloDefinition.SLI, value, query),
		sloDefinition: sloDefinition,
	}
}

func NewFailedSLIWithSLOAndQuery(sloDefinition SLO, sliQuery string, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewFailedSLIResultWithQuery(sloDefinition.SLI, message, sliQuery),
		sloDefinition: sloDefinition,
	}
}

func NewWarningSLIWithSLOAndQuery(sloDefinition SLO, sliQuery string, message string) SLIWithSLO {
	return SLIWithSLO{
		sliResult:     NewWarningSLIResultWithQuery(sloDefinition.SLI, message, sliQuery),
		sloDefinition: sloDefinition,
	}
}

func (r *SLIWithSLO) SLIResult() SLIResult {
	return r.sliResult
}

func (r *SLIWithSLO) SLODefinition() SLO {
	return r.sloDefinition
}

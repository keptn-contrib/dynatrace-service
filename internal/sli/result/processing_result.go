package result

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
)

// ProcessingResult contains the result of processing a get SLI request.
type ProcessingResult struct {
	slo        *keptnapi.ServiceLevelObjectives
	sliResults []SLIResult
}

// NewProcessingResult creates a new ProcessingResult.
func NewProcessingResult(slo *keptnapi.ServiceLevelObjectives, sliResults []SLIResult) *ProcessingResult {
	return &ProcessingResult{
		slo:        slo,
		sliResults: sliResults,
	}
}

// SLOs gets the SLOs.
func (r *ProcessingResult) SLOs() *keptnapi.ServiceLevelObjectives {
	return r.slo
}

// HasSLOs checks whether any objectives are available.
func (r *ProcessingResult) HasSLOs() bool {
	return r.slo != nil && len(r.slo.Objectives) > 0
}

// SLIResults gets the SLI results.
func (r *ProcessingResult) SLIResults() []SLIResult {
	return r.sliResults
}

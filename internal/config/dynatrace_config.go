package config

import "github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

// DynatraceConfig defines the Dynatrace configuration structure
type DynatraceConfig struct {
	SpecVersion string                 `json:"spec_version" yaml:"spec_version"`
	DtCreds     string                 `json:"dtCreds,omitempty" yaml:"dtCreds,omitempty"`
	Dashboard   string                 `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`
	AttachRules *dynatrace.AttachRules `json:"attachRules,omitempty" yaml:"attachRules,omitempty"`
}

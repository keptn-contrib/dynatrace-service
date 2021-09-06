package config

import "github.com/keptn-contrib/dynatrace-service/internal/dynatrace"

// DynatraceConfigFile defines the Dynatrace configuration structure
type DynatraceConfigFile struct {
	SpecVersion string                   `json:"spec_version" yaml:"spec_version"`
	DtCreds     string                   `json:"dtCreds,omitempty" yaml:"dtCreds,omitempty"`
	Dashboard   string                   `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`
	AttachRules *dynatrace.DtAttachRules `json:"attachRules,omitempty" yaml:"attachRules,omitempty"`
}

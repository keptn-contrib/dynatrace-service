package config

// DynatraceConfigFilename is the resource path for the dynatrace.conf.yaml
const DynatraceConfigFilename = "dynatrace/dynatrace.conf.yaml"

// DynatraceConfigFilenameLOCAL is the local test resource path for the dynatrace.conf.yaml
const DynatraceConfigFilenameLOCAL = "dynatrace/_dynatrace.conf.yaml"

// DtTag defines a Dynatrace configuration structure
type DtTag struct {
	Context string `json:"context" yaml:"context"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
}

// DtTagRule defines a Dynatrace configuration structure
type DtTagRule struct {
	MeTypes []string `json:"meTypes" yaml:"meTypes"`
	Tags    []DtTag  `json:"tags" yaml:"tags"`
}

// DtAttachRules defines a Dynatrace configuration structure
type DtAttachRules struct {
	TagRule []DtTagRule `json:"tagRule" yaml:"tagRule"`
}

// DynatraceConfigFile defines the Dynatrace configuration structure
type DynatraceConfigFile struct {
	SpecVersion string         `json:"spec_version" yaml:"spec_version"`
	DtCreds     string         `json:"dtCreds,omitempty" yaml:"dtCreds,omitempty"`
	AttachRules *DtAttachRules `json:"attachRules,omitempty" yaml:"attachRules,omitempty"`
}

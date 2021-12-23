package config

import (
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	"github.com/stretchr/testify/assert"
)

func Test_parseDynatraceConfigYAML(t *testing.T) {
	tests := []struct {
		name       string
		yamlString string
		want       *DynatraceConfig
		wantErr    bool
	}{
		{
			name:       "empty string",
			yamlString: "",
			want:       NewDynatraceConfigWithDefaults(),
			wantErr:    false,
		},
		{
			name: "valid yaml no dashboard",
			yamlString: `
spec_version: '0.1.0'
dtCreds: dyna`,
			want: &DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dyna",
			},
			wantErr: false,
		},
		{
			name: "valid yaml with dashboard",
			yamlString: `
spec_version: '0.1.0'
dtCreds: dyna
dashboard: dash`,
			want: &DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dyna",
				Dashboard:   "dash",
			},
			wantErr: false,
		},
		{
			name: "invalid yaml",
			yamlString: `
spec_version: '0.1.0'
dtCreds: dyna
dashboard: ****`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "yaml with special characters",
			yamlString: `
spec_version: '0.1.0'
dtCreds: dyna
dashboard: '****'`,
			want: &DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dyna",
				Dashboard:   "****",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDynatraceConfigYAML(tt.yamlString)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
		})
	}
}

// TestDynatraceConfigGetter_GetDynatraceConfig tests that placeholders are replaced correctly using data from an event.
func TestDynatraceConfigGetter_GetDynatraceConfig(t *testing.T) {
	mockEvent := test.EventData{
		Context:            "01234567-0123-0123-0123-012345678901",
		Event:              "sh.keptn.event.get-sli.triggered",
		Source:             "service",
		Project:            "myproject",
		Stage:              "mystage",
		Service:            "myservice",
		Deployment:         "mydeployment",
		TestStrategy:       "myteststrategy",
		DeploymentStrategy: "mydeploymentstrategy",
		Labels: map[string]string{
			"dashboard": "12345678-1111-4444-8888-123456789012",
			"metype":    "SERVICE",
			"context":   "CONTEXT1",
			"key":       "special_tag",
			"value":     "special_value"},
	}

	tests := []struct {
		name         string
		configString string
		wantConfig   DynatraceConfig
		wantErr      bool
	}{
		{
			name: "Test with attach rules",
			configString: `spec_version: '0.1.0'
dtCreds: dynatrace-$PROJECT
dashboard: $LABEL.dashboard
attachRules:
  tagRule:
  - meTypes:
    - $LABEL.metype
    tags:
    - context: CONTEXTLESS
      key: keptn_project
      value: $PROJECT
    - context: CONTEXTLESS
      key: keptn_service
      value: $SERVICE
    - context: CONTEXTLESS
      key: keptn_stage
      value: $STAGE
    - context: $LABEL.context
      key: $LABEL.key
      value: $LABEL.value`,
			wantConfig: DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dynatrace-myproject",
				Dashboard:   "12345678-1111-4444-8888-123456789012",
				AttachRules: &dynatrace.AttachRules{
					TagRule: []dynatrace.TagRule{
						{
							MeTypes: []string{
								"SERVICE",
							},
							Tags: []dynatrace.TagEntry{
								{
									Context: "CONTEXTLESS",
									Key:     "keptn_project",
									Value:   "myproject",
								},
								{
									Context: "CONTEXTLESS",
									Key:     "keptn_service",
									Value:   "myservice",
								},
								{
									Context: "CONTEXTLESS",
									Key:     "keptn_stage",
									Value:   "mystage",
								},
								{
									Context: "CONTEXT1",
									Key:     "special_tag",
									Value:   "special_value",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Test with partial attach rules",
			configString: `spec_version: '0.1.0'
dtCreds: dynatrace-$PROJECT
dashboard: $LABEL.dashboard
attachRules:
  tagRule:
  - meTypes:
    - $LABEL.metype`,
			wantConfig: DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dynatrace-myproject",
				Dashboard:   "12345678-1111-4444-8888-123456789012",
				AttachRules: &dynatrace.AttachRules{
					TagRule: []dynatrace.TagRule{
						{
							MeTypes: []string{
								"SERVICE",
							},
							Tags: []dynatrace.TagEntry{},
						},
					},
				},
			},
		},
		{
			name: "Test without attach rules",
			configString: `spec_version: '0.1.0'
dtCreds: dynatrace-$PROJECT
dashboard: $LABEL.dashboard`,
			wantConfig: DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dynatrace-myproject",
				Dashboard:   "12345678-1111-4444-8888-123456789012",
			},
		},
		{
			name: "Test with label that does not exist",
			configString: `spec_version: '0.1.0'
dtCreds: dynatrace-$PROJECT
dashboard: $LABEL.my_dashboard`,
			wantConfig: DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dynatrace-myproject",
				Dashboard:   "$LABEL.my_dashboard",
			},
		},
		{
			name: "Test with label with prefix that exists",
			configString: `spec_version: '0.1.0'
dtCreds: dynatrace-$PROJECT
dashboard: $LABEL.dashboard_name`,
			wantConfig: DynatraceConfig{
				SpecVersion: "0.1.0",
				DtCreds:     "dynatrace-myproject",
				Dashboard:   "12345678-1111-4444-8888-123456789012_name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configGetter := NewDynatraceConfigGetter(&dynatraceConfigResourceClientMock{configString: tt.configString})
			config, err := configGetter.GetDynatraceConfig(&mockEvent)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, &tt.wantConfig, config)
			}
		})
	}
}

type dynatraceConfigResourceClientMock struct {
	configString string
}

func (c *dynatraceConfigResourceClientMock) GetDynatraceConfig(project string, stage string, service string) (string, error) {
	return c.configString, nil
}

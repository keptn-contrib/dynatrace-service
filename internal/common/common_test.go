package common

import (
	"reflect"
	"testing"
)

func Test_parseDynatraceConfigFile(t *testing.T) {
	tests := []struct {
		name       string
		yamlString string
		want       DynatraceConfigFile
		wantErr    bool
	}{
		{
			name:       "empty string",
			yamlString: "",
			want:       DynatraceConfigFile{},
			wantErr:    false,
		},
		{
			name: "valid yaml no dashboard",
			yamlString: `
spec_version: '0.1.0'
dtCreds: dyna`,
			want: DynatraceConfigFile{
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
			want: DynatraceConfigFile{
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
			want:    DynatraceConfigFile{},
			wantErr: true,
		},
		{
			name: "yaml with special characters",
			yamlString: `
spec_version: '0.1.0'
dtCreds: dyna
dashboard: '****'`,
			want: DynatraceConfigFile{
				SpecVersion: "0.1.0",
				DtCreds:     "dyna",
				Dashboard:   "****",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDynatraceConfigFile(tt.yamlString)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDynatraceConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDynatraceConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

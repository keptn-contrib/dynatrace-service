package common

import (
	"os"
	"testing"
)

func TestGetKeptnBridgeURL(t *testing.T) {
	tests := []struct {
		name            string
		want            string
		wantErr         bool
		bridgeURLEnvVar string
	}{
		{
			name:            "return bridge URL",
			want:            "https://bridge.keptn",
			wantErr:         false,
			bridgeURLEnvVar: "bridge.keptn",
		},
		{
			name:            "return bridge URL",
			want:            "https://bridge.keptn",
			wantErr:         false,
			bridgeURLEnvVar: "https://bridge.keptn",
		},
		{
			name:            "return bridge URL with http",
			want:            "http://bridge.keptn",
			wantErr:         false,
			bridgeURLEnvVar: "http://bridge.keptn",
		},
		{
			name:            "return error if env var not set",
			want:            "",
			wantErr:         true,
			bridgeURLEnvVar: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("KEPTN_BRIDGE_URL", tt.bridgeURLEnvVar)
			got, err := GetKeptnBridgeURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeptnBridgeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetKeptnBridgeURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

package common

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimestampToString(t *testing.T) {
	dt := time.Date(1970, 1, 1, 0, 1, 23, 456, time.UTC)
	expected := "83000" // = (1*60 + 23) * 1000 ms

	got := TimestampToString(dt)

	assert.EqualValues(t, expected, got)
}

// tests the parseUnixTimestamp with invalid params
func TestParseInvalidUnixTimestamp(t *testing.T) {
	_, err := ParseUnixTimestamp("")

	assert.NotNil(t, err)
}

// tests the parseUnixTimestamp with valid params
func TestParseValidUnixTimestamp(t *testing.T) {
	expectedTime := time.Date(2019, 10, 24, 15, 44, 27, 152330783, time.UTC)

	got, _ := ParseUnixTimestamp("2019-10-24T15:44:27.152330783Z")

	assert.EqualValues(t, expectedTime, got)
}

func TestParsePassAndWarningFromString(t *testing.T) {
	type args struct {
		customName string
	}
	tests := []struct {
		name string
		args args
		want keptnapi.SLO
	}{
		{
			name: "simple test",
			args: args{
				customName: "Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true",
			},
			want: keptnapi.SLO{
				SLI:     "teststep_rt",
				Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{"<500ms", "<+10%"}}},
				Warning: []*keptnapi.SLOCriteria{{Criteria: []string{"<1000ms", "<+20%"}}},
				Weight:  1,
				KeySLI:  true,
			},
		},
		{
			name: "test with = in pass/warn expression",
			args: args{
				customName: "Host Disk Queue Length (max);sli=host_disk_queue;pass=<=0;warning=<1;key=false",
			},
			want: keptnapi.SLO{
				SLI:     "host_disk_queue",
				Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{"<=0"}}},
				Warning: []*keptnapi.SLOCriteria{{Criteria: []string{"<1"}}},
				Weight:  1,
				KeySLI:  false,
			},
		},
		{
			name: "test weight",
			args: args{
				customName: "Host CPU %;sli=host_cpu;pass=<20;warning=<50;key=false;weight=2",
			},
			want: keptnapi.SLO{
				SLI:     "host_cpu",
				Pass:    []*keptnapi.SLOCriteria{{Criteria: []string{"<20"}}},
				Warning: []*keptnapi.SLOCriteria{{Criteria: []string{"<50"}}},
				Weight:  2,
				KeySLI:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePassAndWarningWithoutDefaultsFrom(tt.args.customName)

			assert.EqualValues(t, &tt.want, got)
		})
	}
}

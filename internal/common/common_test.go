package common

import (
	"testing"
	"time"

	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
)

func TestTimestampToUnixMillisecondsString(t *testing.T) {
	dt := time.Date(1970, 1, 1, 0, 1, 23, 456, time.UTC)
	expected := "83000" // = (1*60 + 23) * 1000 ms

	got := TimestampToUnixMillisecondsString(dt)

	assert.EqualValues(t, expected, got)
}

func TestParseSLOFromString_SuccessCases(t *testing.T) {
	tests := []struct {
		name      string
		sloString string
		want      *keptnapi.SLO
	}{
		{
			name:      "just some description - so no error",
			sloString: "Some description",
			want:      createSLO("", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "just some description, but with separator - so no error",
			sloString: "Some description;with separator",
			want:      createSLO("", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "multiple pass and warning criteria - AND",
			sloString: "Some description;sli=teststep_rt;pass=<500,<+10%;warning=<1000,<+20%;weight=1;key=true",
			want:      createSLO("teststep_rt", [][]string{{"<500", "<+10%"}}, [][]string{{"<1000", "<+20%"}}, 1, true),
		},
		{
			name:      "multiple pass and warning criteria - AND/OR",
			sloString: "Some description;sli=teststep_rt;pass=>=500,>-10%;pass=>=400,>=-15%;warning=<1000,<+20%;warning=<900,<+25%;weight=1;key=true",
			want:      createSLO("teststep_rt", [][]string{{">=500", ">-10%"}, {">=400", ">=-15%"}}, [][]string{{"<1000", "<+20%"}, {"<900", "<+25%"}}, 1, true),
		},
		{
			name:      "test with = in pass/warn expression",
			sloString: "Host Disk Queue Length (max);sli=host_disk_queue;pass==0;warning=<=1;key=false",
			want:      createSLO("host_disk_queue", [][]string{{"=0"}}, [][]string{{"<=1"}}, 1, false),
		},
		{
			name:      "test weight",
			sloString: "Host CPU %;sli=host_cpu;pass=<20;warning=<50;key=false;weight=2",
			want:      createSLO("host_cpu", [][]string{{"<20"}}, [][]string{{"<50"}}, 2, false),
		},
		{
			name:      "informational SLI only - no pass or warn",
			sloString: "Host CPU %;sli=host_cpu;just for informational purposes",
			want:      createSLO("host_cpu", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "informational SLI name with space - should be underscore",
			sloString: "Host CPU %;sli=host cpu;just for informational purposes",
			want:      createSLO("host cpu", [][]string{}, [][]string{}, 1, false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSLOFromString(tt.sloString)
			if assert.NoError(t, err) {
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestParseSLOFromString_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		sloString   string
		errMessages []string
	}{
		{
			name:        "invalid pass criterion - ms suffix",
			sloString:   "Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000,<+20%;weight=1;key=true",
			errMessages: []string{"pass", "<500ms"},
		},
		{
			name:        "invalid pass and warning criteria - ms suffixes",
			sloString:   "Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true",
			errMessages: []string{"pass", "<500ms", "warning", "<1000ms"},
		},
		{
			name:        "invalid pass criterion - wrong operator",
			sloString:   "sli=some_sli_name;pass=<<500",
			errMessages: []string{"pass", "<<500"},
		},
		{
			name:        "invalid pass criterion - wrong decimal notation",
			sloString:   "sli=some_sli_name;pass=<500.",
			errMessages: []string{"pass", "<500."},
		},
		{
			name:        "invalid pass criterion - wrong decimal notation with percent",
			sloString:   "sli=some_sli_name;pass=<500.%",
			errMessages: []string{"pass", "<500.%"},
		},
		{
			name:        "invalid warning criterion - some string",
			sloString:   "sli=some_sli_name;warning=no!",
			errMessages: []string{"warning", "no!"},
		},
		{
			name:        "invalid weight - not an int",
			sloString:   "sli=some_sli_name;weight=3.14",
			errMessages: []string{"weight", "3.14"},
		},
		{
			name:        "invalid keySli - not a bool",
			sloString:   "sli=some_sli_name;key=yes",
			errMessages: []string{"key", "yes"},
		},
		{
			name:        "sli name is empty",
			sloString:   "sli=;pass=<600",
			errMessages: []string{"sli", "is empty"},
		},
		{
			name:        "sli name is empty - only space",
			sloString:   "sli= ;pass=<600",
			errMessages: []string{"sli", "is empty"},
		},
		{
			name:        "duplicate sli name",
			sloString:   "sli=first_name;pass=<600;sli=last_name",
			errMessages: []string{"'sli'", "duplicate key"},
		},
		{
			name:        "duplicate key",
			sloString:   "sli=first_name;key=true;pass=<600;key=false",
			errMessages: []string{"'key'", "duplicate key"},
		},
		{
			name:        "duplicate weight",
			sloString:   "sli=first_name;weight=7;pass=<600;weight=3",
			errMessages: []string{"'weight'", "duplicate key"},
		},
		{
			name:        "duplication for sli, key, weight",
			sloString:   "sli=first_name;weight=7;key=false;sli=last_name;pass=<600;weight=3;key=true",
			errMessages: []string{"'weight'", "'key'", "'sli'", "duplicate key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSLOFromString(tt.sloString)
			if assert.Error(t, err) {
				for _, errMessage := range tt.errMessages {
					assert.Contains(t, err.Error(), errMessage)
				}
			}
		})
	}
}

func createSLO(indicatorName string, pass [][]string, warning [][]string, weight int, isKey bool) *keptnapi.SLO {
	var passCriteria []*keptnapi.SLOCriteria
	for _, criteria := range pass {
		passCriteria = append(passCriteria, &keptnapi.SLOCriteria{Criteria: criteria})
	}

	var warningCriteria []*keptnapi.SLOCriteria
	for _, criteria := range warning {
		warningCriteria = append(warningCriteria, &keptnapi.SLOCriteria{Criteria: criteria})
	}

	return &keptnapi.SLO{
		SLI:     indicatorName,
		Pass:    passCriteria,
		Warning: warningCriteria,
		Weight:  weight,
		KeySLI:  isKey,
	}
}

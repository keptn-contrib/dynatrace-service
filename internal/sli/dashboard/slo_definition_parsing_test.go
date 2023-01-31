package dashboard

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/result"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSLODefinition_SuccessCases(t *testing.T) {
	tests := []struct {
		name      string
		sloString string
		want      sloDefinitionParsingResult
	}{
		{
			name:      "just some SLI - so no error",
			sloString: "Some SLI",
			want:      createSLODefinitionParsingResult(false, "some_sli", "Some SLI", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "just some SLI, but with separator - so no error",
			sloString: "Some SLI;with separator",
			want:      createSLODefinitionParsingResult(false, "some_sli", "Some SLI", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "multiple pass and warning criteria - AND",
			sloString: "Test step response time;sli=teststep_rt;pass=<500,<+10%;warning=<1000,<+20%;weight=1;key=true",
			want:      createSLODefinitionParsingResult(false, "teststep_rt", "Test step response time", [][]string{{"<500", "<+10%"}}, [][]string{{"<1000", "<+20%"}}, 1, true),
		},
		{
			name:      "multiple pass and warning criteria - AND/OR",
			sloString: "Test step response time;sli=teststep_rt;pass=>=500,>-10%;pass=>=400,>=-15%;warning=<1000,<+20%;warning=<900,<+25%;weight=1;key=true",
			want:      createSLODefinitionParsingResult(false, "teststep_rt", "Test step response time", [][]string{{">=500", ">-10%"}, {">=400", ">=-15%"}}, [][]string{{"<1000", "<+20%"}, {"<900", "<+25%"}}, 1, true),
		},
		{
			name:      "multiple pass and warning criteria - AND/OR with decimals",
			sloString: "Test step response time;sli=teststep_rt;pass=>=500.74,>-10.3%;pass=>=400.89,>=-15.7%;warning=<1000.12,<+20.50%;warning=<900.34,<+25.29%;weight=1;key=true",
			want:      createSLODefinitionParsingResult(false, "teststep_rt", "Test step response time", [][]string{{">=500.74", ">-10.3%"}, {">=400.89", ">=-15.7%"}}, [][]string{{"<1000.12", "<+20.50%"}, {"<900.34", "<+25.29%"}}, 1, true),
		},
		{
			name:      "test with = in pass/warn expression",
			sloString: "Host Disk Queue Length (max);sli=host_disk_queue;pass==0;warning=<=1;key=false",
			want:      createSLODefinitionParsingResult(false, "host_disk_queue", "Host Disk Queue Length (max)", [][]string{{"=0"}}, [][]string{{"<=1"}}, 1, false),
		},
		{
			name:      "test weight",
			sloString: "Host CPU %;sli=host_cpu;pass=<20;warning=<50;key=false;weight=2",
			want:      createSLODefinitionParsingResult(false, "host_cpu", "Host CPU %", [][]string{{"<20"}}, [][]string{{"<50"}}, 2, false),
		},
		{
			name:      "informational SLI only - no pass or warn",
			sloString: "Host CPU %;sli=host_cpu;just for informational purposes",
			want:      createSLODefinitionParsingResult(false, "host_cpu", "Host CPU %", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "informational SLI name with space - changed to underscore",
			sloString: "Host CPU %;sli=host cpu;just for informational purposes",
			want:      createSLODefinitionParsingResult(false, "host_cpu", "Host CPU %", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "informational SLI name with uppercase - changed to lowercase",
			sloString: "Host CPU %;sli=HOST_CPU;just for informational purposes",
			want:      createSLODefinitionParsingResult(false, "host_cpu", "Host CPU %", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "informational SLI name with space - no display name",
			sloString: "sli=host cpu;just for informational purposes",
			want:      createSLODefinitionParsingResult(false, "host_cpu", "host cpu", [][]string{}, [][]string{}, 1, false),
		},
		{
			name:      "excluded tile",
			sloString: "example data explorer tile; exclude=true",
			want:      createSLODefinitionParsingResult(true, "example_data_explorer_tile", "example data explorer tile", [][]string{}, [][]string{}, 1, false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSLODefinition(tt.sloString)
			if assert.NoError(t, err) {
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestParseSLODefinition_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		sloString   string
		want        sloDefinitionParsingResult
		errMessages []string
	}{
		{
			name:        "invalid pass criterion - ms suffix",
			sloString:   "Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000,<+20%;weight=1;key=true",
			want:        createSLODefinitionParsingResult(false, "teststep_rt", "Some description", [][]string{}, [][]string{{"<1000", "<+20%"}}, 1, true),
			errMessages: []string{"pass", "<500ms"},
		},
		{
			name:        "invalid pass and warning criteria - ms suffixes",
			sloString:   "Some description;sli=teststep_rt;pass=<500ms,<+10%;warning=<1000ms,<+20%;weight=1;key=true",
			want:        createSLODefinitionParsingResult(false, "teststep_rt", "Some description", [][]string{}, [][]string{}, 1, true),
			errMessages: []string{"pass", "<500ms", "warning", "<1000ms"},
		},
		{
			name:        "invalid pass criterion - wrong operator",
			sloString:   "sli=some_sli_name;pass=<<500",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"pass", "<<500"},
		},
		{
			name:        "invalid pass criterion - wrong decimal notation",
			sloString:   "sli=some_sli_name;pass=<500.",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"pass", "<500."},
		},
		{
			name:        "invalid pass criterion - wrong decimal notation with percent",
			sloString:   "sli=some_sli_name;pass=<500.%",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"pass", "<500.%"},
		},
		{
			name:        "invalid warning criterion - wrong decimal notation with percent and wrong type",
			sloString:   "sli=some_sli_name;warning=<500.%,yes",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"warning", "<500.%", "yes"},
		},
		{
			name:        "invalid warning criterion - some string",
			sloString:   "sli=some_sli_name;warning=yes!",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"warning", "yes!"},
		},
		{
			name:        "invalid warning criterion - wrong operator",
			sloString:   "sli=some_sli_name;warning=<<500",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"warning", "<<500"},
		},
		{
			name:        "invalid warning criterion - wrong decimal notation",
			sloString:   "sli=some_sli_name;warning=<500.",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"warning", "<500."},
		},
		{
			name:        "invalid warning criterion - wrong decimal notation with percent",
			sloString:   "sli=some_sli_name;warning=<500.%",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"warning", "<500.%"},
		},
		{
			name:        "invalid warning criterion - wrong decimal notation with percent and wrong type",
			sloString:   "sli=some_sli_name;warning=<500.%,yes",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"warning", "<500.%", "yes"},
		},
		{
			name:        "invalid warning criterion - some string",
			sloString:   "sli=some_sli_name;warning=no!",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"warning", "no!"},
		},
		{
			name:        "invalid weight - not an int",
			sloString:   "sli=some_sli_name;weight=3.14",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"weight", "3.14"},
		},
		{
			name:        "invalid keySli - not a bool",
			sloString:   "sli=some_sli_name;key=yes",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"key", "yes"},
		},
		{
			name:        "invalid exclude - not a bool",
			sloString:   "sli=some_sli_name;exclude=enable",
			want:        createSLODefinitionParsingResult(false, "some_sli_name", "some_sli_name", [][]string{}, [][]string{}, 1, false),
			errMessages: []string{"exclude", "enable"},
		},
		{
			name:        "sli name is empty",
			sloString:   "sli=;pass=<600",
			want:        createSLODefinitionParsingResult(false, "", "", [][]string{{"<600"}}, [][]string{}, 1, false),
			errMessages: []string{"sli", "is empty"},
		},
		{
			name:        "sli name is empty - only space",
			sloString:   "sli= ;pass=<600",
			want:        createSLODefinitionParsingResult(false, "", "", [][]string{{"<600"}}, [][]string{}, 1, false),
			errMessages: []string{"sli", "is empty"},
		},
		{
			name:        "duplicate sli name",
			sloString:   "sli=first_name;pass=<600;sli=last_name",
			want:        createSLODefinitionParsingResult(false, "first_name", "first_name", [][]string{{"<600"}}, [][]string{}, 1, false),
			errMessages: []string{"'sli'", "duplicate key"},
		},
		{
			name:        "duplicate key",
			sloString:   "sli=first_name;key=true;pass=<600;key=false",
			want:        createSLODefinitionParsingResult(false, "first_name", "first_name", [][]string{{"<600"}}, [][]string{}, 1, true),
			errMessages: []string{"'key'", "duplicate key"},
		},
		{
			name:        "duplicate weight",
			sloString:   "sli=first_name;weight=7;pass=<600;weight=3",
			want:        createSLODefinitionParsingResult(false, "first_name", "first_name", [][]string{{"<600"}}, [][]string{}, 7, false),
			errMessages: []string{"'weight'", "duplicate key"},
		},
		{
			name:        "duplicate exclude",
			sloString:   "sli=first_name;exclude=true;pass=<600;exclude=false",
			want:        createSLODefinitionParsingResult(true, "first_name", "first_name", [][]string{{"<600"}}, [][]string{}, 1, false),
			errMessages: []string{"'exclude'", "duplicate key"},
		},
		{
			name:        "duplication for sli, key, weight, exclude",
			sloString:   "sli=first_name;weight=7;key=false;exclude=false;sli=last_name;pass=<600;weight=3;key=true;exclude=true",
			want:        createSLODefinitionParsingResult(false, "first_name", "first_name", [][]string{{"<600"}}, [][]string{}, 7, false),
			errMessages: []string{"'weight'", "'key'", "'sli'", "'exclude'", "duplicate key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSLODefinition(tt.sloString)
			if assert.Error(t, err) {
				for _, errMessage := range tt.errMessages {
					assert.Contains(t, err.Error(), errMessage)
				}
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func createSLODefinitionParsingResult(exclude bool, indicatorName string, displayName string, pass [][]string, warning [][]string, weight int, isKey bool) sloDefinitionParsingResult {
	var passCriteria result.SLOCriteriaList
	for _, criteria := range pass {
		passCriteria = append(passCriteria, &result.SLOCriteria{Criteria: criteria})
	}

	var warningCriteria result.SLOCriteriaList
	for _, criteria := range warning {
		warningCriteria = append(warningCriteria, &result.SLOCriteria{Criteria: criteria})
	}

	return sloDefinitionParsingResult{
		exclude: exclude,
		sloDefinition: result.SLO{
			SLI:         indicatorName,
			DisplayName: displayName,
			Pass:        passCriteria,
			Warning:     warningCriteria,
			Weight:      weight,
			KeySLI:      isKey,
		},
	}
}

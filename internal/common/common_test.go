package common

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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

func TestParseMarkdownConfigurationParams(t *testing.T) {
	testConfigs := []struct {
		input          string
		expectedResult *keptnapi.ServiceLevelObjectives
	}{
		// single result
		{
			"KQG.Total.Pass=90%;KQG.Total.Warning=70%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg",
			createSLO("90%", "70%", "single_result", "pass", 1, "avg"),
		},
		// several results, p50
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p50",
			createSLO("50%", "40%", "several_results", "pass", 3, "p50"),
		},
		// several results, p90
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p90",
			createSLO("50%", "40%", "several_results", "pass", 3, "p90"),
		},
		// several results, p95
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "pass", 3, "p95"),
		},
		// several results, p95, all
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=all;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "all", 3, "p95"),
		},
		// several results, p95, pass_or_warn
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=pass_or_warn;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "pass_or_warn", 3, "p95"),
		},

		// several results, p95, fallback to pass if compare function is unknown
		{
			"KQG.Total.Pass=50%;KQG.Total.Warning=40%;KQG.Compare.WithScore=warn;KQG.Compare.Results=3;KQG.Compare.Function=p95",
			createSLO("50%", "40%", "several_results", "pass", 3, "p95"),
		},
		// several results, fallback if function is unknown e.g. p97
		{
			"KQG.Total.Pass=51%;KQG.Total.Warning=41%;KQG.Compare.WithScore=pass;KQG.Compare.Results=4;KQG.Compare.Function=p97",
			createSLO("51%", "41%", "several_results", "pass", 4, "avg"),
		},
	}
	for _, config := range testConfigs {
		actualSLO := &keptnapi.ServiceLevelObjectives{
			Objectives: []*keptnapi.SLO{},
			TotalScore: &keptnapi.SLOScore{
				Pass:    "",
				Warning: "",
			},
			Comparison: &keptnapi.SLOComparison{
				CompareWith:               "",
				IncludeResultWithScore:    "",
				NumberOfComparisonResults: 0,
				AggregateFunction:         "",
			},
		}
		ParseMarkdownConfiguration(config.input, actualSLO)

		assert.EqualValues(t, config.expectedResult, actualSLO)
	}
}

func createSLO(pass string, warning string, compareWith string, include string, numberOfResults int, aggregateFunc string) *keptnapi.ServiceLevelObjectives {
	return &keptnapi.ServiceLevelObjectives{
		Objectives: []*keptnapi.SLO{},
		TotalScore: &keptnapi.SLOScore{
			Pass:    pass,
			Warning: warning,
		},
		Comparison: &keptnapi.SLOComparison{
			CompareWith:               compareWith,
			IncludeResultWithScore:    include,
			NumberOfComparisonResults: numberOfResults,
			AggregateFunction:         aggregateFunc,
		},
	}
}

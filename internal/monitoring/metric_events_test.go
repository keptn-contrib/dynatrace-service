package monitoring

import "testing"

func Test_getAlertCondition(t *testing.T) {
	type args struct {
		condition string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Expect ABOVE condition",
			args: args{
				condition: "<10",
			},
			want:    "ABOVE",
			wantErr: false,
		},
		{
			name: "Expect BELOW condition",
			args: args{
				condition: ">10",
			},
			want:    "BELOW",
			wantErr: false,
		},
		{
			name: "Expect error",
			args: args{
				condition: ">+10",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Expect error 2",
			args: args{
				condition: ">-10",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Expect error 3",
			args: args{
				condition: ">10%",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAlertCondition(tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAlertCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseAlertCondition() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMetricEventAggregation(t *testing.T) {
	type args struct {
		metricAPIAgg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Expect P90",
			args: args{
				metricAPIAgg: "percentile(90.0)",
			},
			want: "P90",
		},
		{
			name: "Expect P90 2",
			args: args{
				metricAPIAgg: "percentile(95)",
			},
			want: "P90",
		},
		{
			name: "Expect MEDIAN",
			args: args{
				metricAPIAgg: "percentile(10)",
			},
			want: "MEDIAN",
		},
		{
			name: "Expect empty string",
			args: args{
				metricAPIAgg: "avg",
			},
			want: "",
		},
		{
			name: "Expect empty string 2",
			args: args{
				metricAPIAgg: "foo",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMetricEventAggregation(tt.args.metricAPIAgg); got != tt.want {
				t.Errorf("getMetricEventAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}

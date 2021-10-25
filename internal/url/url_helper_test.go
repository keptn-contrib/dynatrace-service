package url

import "testing"

func TestCleanURL(t *testing.T) {
	type args struct {
		u string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "valid - do nothing",
			args:    args{u: "https://otherSampleEnv.live.dynatrace.com"},
			want:    "https://otherSampleEnv.live.dynatrace.com",
			wantErr: false,
		},
		{
			name:    "valid - add https",
			args:    args{u: "otherSampleEnv.live.dynatrace.com"},
			want:    "https://otherSampleEnv.live.dynatrace.com",
			wantErr: false,
		},
		{
			name:    "valid - preserve http",
			args:    args{u: "http://otherSampleEnv.live.dynatrace.com"},
			want:    "http://otherSampleEnv.live.dynatrace.com",
			wantErr: false,
		},
		{
			name:    "valid - with port",
			args:    args{u: "https://otherSampleEnv.live.dynatrace.com:8080"},
			want:    "https://otherSampleEnv.live.dynatrace.com:8080",
			wantErr: false,
		},
		{
			name:    "valid - remove trailing slash",
			args:    args{u: "otherSampleEnv.live.dynatrace.com/"},
			want:    "https://otherSampleEnv.live.dynatrace.com",
			wantErr: false,
		},
		{
			name:    "valid - remove extra whitespace",
			args:    args{u: " otherSampleEnv.live.dynatrace.com "},
			want:    "https://otherSampleEnv.live.dynatrace.com",
			wantErr: false,
		},
		{
			name:    "invalid - local path",
			args:    args{u: "/otherSampleEnv.live.dynatrace.com/"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid - non http or https scheme",
			args:    args{u: "ftp://otherSampleEnv.live.dynatrace.com/"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid - no host",
			args:    args{u: "http:///tenant/"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanURL(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeCleanURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MakeCleanURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

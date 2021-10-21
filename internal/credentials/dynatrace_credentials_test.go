package credentials

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDynatraceCredentials(t *testing.T) {

	validTestTenant := "https://mySampleEnv.live.dynatrace.com"

	type args struct {
		tenant   string
		apiToken string
	}
	tests := []struct {
		name    string
		args    args
		want    *DynatraceCredentials
		wantErr bool
	}{
		{
			name: "valid credentials",
			args: args{tenant: validTestTenant,
				apiToken: testDynatraceAPIToken,
			},
			want: &DynatraceCredentials{
				tenant:   validTestTenant,
				apiToken: testDynatraceAPIToken,
			},
			wantErr: false,
		},
		{
			name: "invalid token - empty",
			args: args{
				tenant:   validTestTenant,
				apiToken: "",
			},
			wantErr: true,
		},
		{
			name: "invalid token - 1 components",
			args: args{tenant: validTestTenant,
				apiToken: "dt0c01ST2EY72KQINMH574WMNVI7YNG3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM",
			},
			wantErr: true,
		},
		{
			name: "invalid token - 4 components",
			args: args{
				tenant:   validTestTenant,
				apiToken: "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM.AGH",
			},
			wantErr: true,
		},
		{
			name: "invalid token - bad public characters",
			args: args{
				tenant:   validTestTenant,
				apiToken: "dt0c01.qT2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM",
			},
			wantErr: true,
		},
		{
			name: "invalid token - public too short",
			args: args{
				tenant:   validTestTenant,
				apiToken: "dt0c01.T2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM",
			},
			wantErr: true,
		},
		{
			name: "invalid token - public too long",
			args: args{
				tenant:   validTestTenant,
				apiToken: "dt0c01.SST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM",
			},
			wantErr: true,
		},
		{
			name: "invalid token - bad secret characters",
			args: args{tenant: validTestTenant,
				apiToken: "dt0c01.ST2EY72KQINMH574WMNVI7YN.a3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM",
			},
			wantErr: true,
		},
		{
			name: "invalid token - secret too short",
			args: args{
				tenant:   validTestTenant,
				apiToken: "dt0c01.ST2EY72KQINMH574WMNVI7YN.3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM",
			},
			wantErr: true,
		},
		{
			name: "invalid token - secret too long",
			args: args{
				tenant:   validTestTenant,
				apiToken: "dt0c01.ST2EY72KQINMH574WMNVI7YN.GG3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDynatraceCredentials(tt.args.tenant, tt.args.apiToken)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
		})
	}
}

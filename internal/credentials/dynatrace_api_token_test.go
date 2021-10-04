package credentials

import (
	"reflect"
	"testing"
)

func TestNewDynatraceAPIToken(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name    string
		args    args
		want    *DynatraceAPIToken
		wantErr bool
	}{
		{
			name:    "valid token",
			args:    args{t: "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			want:    &DynatraceAPIToken{token: "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: false,
		},
		{
			name:    "invalid token - empty",
			args:    args{t: ""},
			wantErr: true,
		},
		{
			name:    "invalid token - 1 components",
			args:    args{t: "dt0c01ST2EY72KQINMH574WMNVI7YNG3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: true,
		},
		{
			name:    "invalid token - 4 components",
			args:    args{t: "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM.AGH"},
			wantErr: true,
		},
		{
			name:    "invalid token - bad public characters",
			args:    args{t: "dt0c01.qT2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: true,
		},
		{
			name:    "invalid token - pulic too short",
			args:    args{t: "dt0c01.T2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: true,
		},
		{
			name:    "invalid token - pulic too long",
			args:    args{t: "dt0c01.SST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: true,
		},
		{
			name:    "invalid token - bad secret characters",
			args:    args{t: "dt0c01.ST2EY72KQINMH574WMNVI7YN.a3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: true,
		},
		{
			name:    "invalid token - secret too short",
			args:    args{t: "dt0c01.ST2EY72KQINMH574WMNVI7YN.3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: true,
		},
		{
			name:    "invalid token - secret too long",
			args:    args{t: "dt0c01.ST2EY72KQINMH574WMNVI7YN.GG3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDynatraceAPIToken(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDynatraceAPIToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDynatraceAPIToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

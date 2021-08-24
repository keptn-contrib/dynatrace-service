package dynatrace

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

func TestDynatraceHelper_createClient(t *testing.T) {

	mockTenant := "https://mySampleEnv.live.dynatrace.com"
	mockReq, err := http.NewRequest("GET", mockTenant+"/api/v1/config/clusterversion", bytes.NewReader(make([]byte, 100)))
	if err != nil {
		t.Errorf("DynatraceHelper.createClient(): unable to make mock request: error = %v", err)
		return
	}

	mockProxy := "https://proxy:8080"
	t.Logf("Using mock proxy: %v", mockProxy)

	type proxyEnvVars struct {
		httpProxy  string
		httpsProxy string
		noProxy    string
	}
	type fields struct {
		DynatraceCreds *credentials.DTCredentials
	}
	type args struct {
		req *http.Request
	}

	// only one test can be run in a single test run due to the ProxyConfig environment being cached
	// see envProxyFunc() in transport.go for details
	tests := []struct {
		name         string
		proxyEnvVars proxyEnvVars
		fields       fields
		args         args
		wantErr      bool
		wantProxy    string
	}{
		{
			name: "testWithProxy",
			proxyEnvVars: proxyEnvVars{
				httpProxy:  mockProxy,
				httpsProxy: mockProxy,
				noProxy:    "localhost",
			},
			fields: fields{
				DynatraceCreds: &credentials.DTCredentials{
					Tenant:   mockTenant,
					ApiToken: "",
				},
			},
			args: args{
				req: mockReq,
			},
			wantProxy: mockProxy,
		},
		/*{
			name: "testWithNoProxy",
			fields: fields{
				DynatraceCreds: &credentials.DTCredentials{
					Tenant:   mockTenant,
					ApiToken: "",
				},
				Logger: keptncommon.NewLogger("", "", ""),
			},
			args: args{
				req: mockReq,
			},
			wantProxy: "",
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("HTTP_PROXY", tt.proxyEnvVars.httpProxy)
			os.Setenv("HTTPS_PROXY", tt.proxyEnvVars.httpsProxy)
			os.Setenv("NO_PROXY", tt.proxyEnvVars.noProxy)

			dt := &DynatraceHelper{
				DynatraceCreds: tt.fields.DynatraceCreds,
			}

			gotClient, err := dt.createClient(tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("DynatraceHelper.createClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotTransport := gotClient.Transport.(*http.Transport)
			gotProxyUrl, err := gotTransport.Proxy(tt.args.req)
			if err != nil {
				t.Errorf("DynatraceHelper.createClient() error = %v", err)
				return
			}

			if gotProxyUrl == nil {
				if tt.wantProxy != "" {
					t.Errorf("DynatraceHelper.createClient() error, got proxy is nil, wanted = %v", tt.wantProxy)
				}
			} else {
				gotProxy := gotProxyUrl.String()
				if tt.wantProxy == "" {
					t.Errorf("DynatraceHelper.createClient() error, got proxy = %v, wanted nil", gotProxy)
				} else if gotProxy != tt.wantProxy {
					t.Errorf("DynatraceHelper.createClient() error, got proxy = %v, wanted = %v", gotProxy, tt.wantProxy)
				}
			}

			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("NO_PROXY")
		})
	}
}

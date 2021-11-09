package metrics

import (
	"testing"
	"time"

	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/stretchr/testify/assert"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestBuildingMetricQueryWorks(t *testing.T) {

	ev := &test.EventData{
		Project: "my-project",
		Stage:   "my-stage",
		Service: "my-service",
		Labels: map[string]string{
			"good-label1": "tom and jerry",
			"good-label2": "tom + jerry/gerry",
			"bad-label1":  "tom & jerry",
		},
	}
	startTime := time.Unix(1636000000, 0)
	endTime := time.Unix(1636000120, 0)

	testConfigs := []struct {
		name                   string
		input                  string
		expectedMetricQuery    string
		expectedMetricSelector string
		sliFilter              []*keptnv2.SLIFilter
		shouldFail             bool
		errMessage             string
	}{
		{
			name:                   "simple old format transformed to new one",
			input:                  "builtin:service.requestCount.total:merge(0):sum?scope=tag(keptn_project:my-proj),tag(keptn_stage:dev),tag(keptn_service:carts),tag(keptn_deployment:direct)",
			expectedMetricQuery:    "entitySelector=tag%28keptn_project%3Amy-proj%29%2Ctag%28keptn_stage%3Adev%29%2Ctag%28keptn_service%3Acarts%29%2Ctag%28keptn_deployment%3Adirect%29%2Ctype%28SERVICE%29&from=1636000000000&metricSelector=builtin%3Aservice.requestCount.total%3Amerge%280%29%3Asum&resolution=Inf&to=1636000120000",
			expectedMetricSelector: "builtin:service.requestCount.total:merge(0):sum",
		},
		{
			name:                   "event context data is correctly encoded in metric V2 query",
			input:                  "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag($LABEL.good-label1),tag($LABEL.good-label2)",
			expectedMetricQuery:    "entitySelector=type%28SERVICE%29%2Ctag%28keptn_project%3Amy-project%29%2Ctag%28keptn_stage%3Amy-stage%29%2Ctag%28keptn_service%3Amy-service%29%2Ctag%28tom+and+jerry%29%2Ctag%28tom+%2B+jerry%2Fgerry%29&from=1636000000000&metricSelector=builtin%3Aservice.response.time%3Amerge%28%22dt.entity.service%22%29%3Apercentile%2895%29&resolution=Inf&to=1636000120000",
			expectedMetricSelector: "builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)",
		},
		{
			name:                   "uri reserved characters are encoded correctly",
			input:                  "metricSelector=(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
			expectedMetricQuery:    "from=1636000000000&metricSelector=%28calc%3Aservice.%24rt_csm%3Afilter%28and%28eq%28Dimension%2C%22request+Actions.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave_Rest_customer-profile.%22%29%2Cin%28%22dt.entity.service%22%2CentitySelector%28%22type%28service%29%2CrequestAttribute%28~%22%24svc_id~%22%29%22%29%29%29%29%3AsplitBy%28%29%3Aavg%3Aauto%3Asort%28value%28avg%2Cdescending%29%29%29%2F%28%28calc%3Aservice.%24reqcnt_csm%3Afilter%28and%28in%28%22dt.entity.service%22%2CentitySelector%28%22type%28service%29%2CrequestAttribute%28~%22%24svc_id~%22%29%22%29%29%2Ceq%28Dimension%2C%22request+Actions.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave.BO_EC11_NewCustAdd%2B_06_EntrDtlsAndSave_Rest_customer-profile.%22%29%29%29%29%3AsplitBy%28%29%3Asum%3Aauto%3Asort%28value%28sum%2Cdescending%29%29%2F%281%29%29&resolution=Inf&to=1636000120000",
			expectedMetricSelector: "(calc:service.$rt_csm:filter(and(eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\"),in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")))):splitBy():avg:auto:sort(value(avg,descending)))/((calc:service.$reqcnt_csm:filter(and(in(\"dt.entity.service\",entitySelector(\"type(service),requestAttribute(~\"$svc_id~\")\")),eq(Dimension,\"request Actions.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave.BO_EC11_NewCustAdd+_06_EntrDtlsAndSave_Rest_customer-profile.\")))):splitBy():sum:auto:sort(value(sum,descending))/(1))",
		},
		{
			// actually this is a short coming in the current SLI format design - Dynatrace API would not complain
			name:       "event context data cannot be correctly encoded because of '&' and fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag($LABEL.bad-label1)",
			shouldFail: true,
			errMessage: "could not parse metrics query",
		},
		{
			name:       "misspelled metricSelector key fails",
			input:      "metricsSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE)",
			shouldFail: true,
			errMessage: "unknown key",
		},
		{
			name:       "duplicate entitySelector key fails",
			input:      "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE)&entitySelector=type(SERVICE)",
			shouldFail: true,
			errMessage: "duplicate key",
		},
	}
	for _, testConfig := range testConfigs {
		tc := testConfig
		t.Run(tc.name, func(t *testing.T) {
			actualMetricQuery, actualMetricSelector, err := NewQueryBuilder(ev, tc.sliFilter).Build(tc.input, startTime, endTime)
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Empty(t, actualMetricQuery)
				assert.Empty(t, actualMetricSelector)
				assert.Contains(t, err.Error(), tc.errMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMetricQuery, actualMetricQuery)
				assert.Equal(t, tc.expectedMetricSelector, actualMetricSelector)
				assert.Empty(t, tc.errMessage, "fix test setup")
			}
		})
	}
}

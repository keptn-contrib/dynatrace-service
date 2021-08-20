package keptn

import (
	"errors"
	"fmt"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const sliResourceURI = "dynatrace/sli.yaml"

const throughput = "throughput"
const errorRate = "error_rate"
const responseTimeP50 = "response_time_p50"
const responseTimeP90 = "response_time_p90"
const responseTimeP95 = "response_time_p95"

type CustomQueries struct {
	values map[string]string
}

func (cq *CustomQueries) GetQueryByNameOrDefault(sliName string) (string, error) {
	query, exists := cq.values[sliName]
	if exists {
		return query, nil
	}

	defaultQuery, err := getDefaultQuery(sliName)
	if err != nil {
		return "", err
	}

	return defaultQuery, nil
}

type Client struct {
	client *keptnv2.Keptn
}

func NewClient(client *keptnv2.Keptn) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) GetCustomQueries(project string, stage string, service string) (*CustomQueries, error) {
	if c.client == nil {
		return nil, errors.New("could not retrieve SLI config: no Keptn client initialized")
	}

	customQueries, err := c.client.GetSLIConfiguration(project, stage, service, sliResourceURI)
	if err != nil {
		return nil, err
	}

	return &CustomQueries{values: customQueries}, nil
}

// based on the requested metric a dynatrace time series with its aggregation type is returned
func getDefaultQuery(sliName string) (string, error) {
	switch sliName {
	case throughput:
		return "builtin:service.requestCount.total:merge(0):count?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case errorRate:
		return "builtin:service.errors.total.rate:merge(0):avg?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP50:
		return "builtin:service.response.time:merge(0):percentile(50)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP90:
		return "builtin:service.response.time:merge(0):percentile(90)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP95:
		return "builtin:service.response.time:merge(0):percentile(95)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	default:
		return "", fmt.Errorf("unsupported SLI metric %s", sliName)
	}
}

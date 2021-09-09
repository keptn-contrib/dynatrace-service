package keptn

import (
	"errors"
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

const sliResourceURI = "dynatrace/sli.yaml"

const Throughput = "throughput"
const errorRate = "error_rate"
const ResponseTimeP50 = "response_time_p50"
const responseTimeP90 = "response_time_p90"
const responseTimeP95 = "response_time_p95"

type CustomQueries struct {
	values map[string]string
}

func NewEmptyCustomQueries() *CustomQueries {
	return &CustomQueries{
		values: make(map[string]string),
	}
}

func NewCustomQueries(values map[string]string) *CustomQueries {
	return &CustomQueries{
		values: values,
	}
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

type ClientInterface interface {
	GetCustomQueries(project string, stage string, service string) (*CustomQueries, error)
	GetShipyard() (*keptnv2.Shipyard, error)
	SendCloudEvent(factory adapter.CloudEventFactoryInterface) error
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

func (c *Client) GetShipyard() (*keptnv2.Shipyard, error) {
	shipyard, err := c.client.GetShipyard()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve shipyard for project %s: %v", c.client.Event.GetProject(), err)
	}

	return shipyard, nil
}

func (c *Client) SendCloudEvent(factory adapter.CloudEventFactoryInterface) error {
	ev, err := factory.CreateCloudEvent()
	if err != nil {
		return fmt.Errorf("could not create cloud event: %s", err)
	}

	if err := c.client.SendCloudEvent(*ev); err != nil {
		return fmt.Errorf("could not send %s event: %s", ev.Type(), err.Error())
	}

	return nil
}

// based on the requested metric a dynatrace time series with its aggregation type is returned
func getDefaultQuery(sliName string) (string, error) {
	// Switched to new metric v2 query language as discussed here: https://github.com/keptn-contrib/dynatrace-sli-service/issues/91
	switch sliName {
	case Throughput:
		return "metricSelector=builtin:service.requestCount.total:merge(0):sum&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case errorRate:
		return "metricSelector=builtin:service.errors.total.rate:merge(0):avg&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case ResponseTimeP50:
		return "metricSelector=builtin:service.response.time:merge(0):percentile(50)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP90:
		return "metricSelector=builtin:service.response.time:merge(0):percentile(90)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	case responseTimeP95:
		return "metricSelector=builtin:service.response.time:merge(0):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)", nil
	default:
		return "", fmt.Errorf("unsupported SLI %s", sliName)
	}
}

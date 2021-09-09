package dynatrace

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

const metricEventsPath = "/api/config/v1/anomalyDetection/metricEvents"

type MetricEvent struct {
	Metadata          MEMetadata        `json:"metadata"`
	ID                string            `json:"id,omitempty"`
	MetricID          string            `json:"metricId"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	AggregationType   string            `json:"aggregationType,omitempty"`
	EventType         string            `json:"eventType"`
	Severity          string            `json:"severity"`
	AlertCondition    string            `json:"alertCondition"`
	Samples           int               `json:"samples"`
	ViolatingSamples  int               `json:"violatingSamples"`
	DealertingSamples int               `json:"dealertingSamples"`
	Threshold         float64           `json:"threshold"`
	Enabled           bool              `json:"enabled"`
	TagFilters        []METagFilter     `json:"tagFilters,omitempty"`
	AlertingScope     []MEAlertingScope `json:"alertingScope"`
	Unit              string            `json:"unit,omitempty"`
}
type MEMetadata struct {
	ConfigurationVersions []int  `json:"configurationVersions"`
	ClusterVersion        string `json:"clusterVersion"`
}

type METagFilter struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}
type MEAlertingScope struct {
	FilterType       string       `json:"filterType"`
	TagFilter        *METagFilter `json:"tagFilter"`
	ManagementZoneID int64        `json:"managementZoneId,omitempty"`
}

type MetricEventsClient struct {
	client ClientInterface
}

func NewMetricEventsClient(client ClientInterface) *MetricEventsClient {
	return &MetricEventsClient{
		client: client,
	}
}

func (mec *MetricEventsClient) getAll() (*listResponse, error) {
	res, err := mec.client.Get(metricEventsPath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of existing Dynatrace metric events: %v", err)
	}

	response := &listResponse{}
	err = json.Unmarshal(res, response)
	if err != nil {
		return nil, fmt.Errorf("could not parse list of existing Dynatrace metric events: %v", err)
	}

	return response, nil
}

func (mec *MetricEventsClient) getByID(metricEventID string) (*MetricEvent, error) {
	res, err := mec.client.Get(metricEventsPath + "/" + metricEventID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve metric event with ID: %s, %v", metricEventID, err)
	}

	retrievedMetricEvent := &MetricEvent{}
	err = json.Unmarshal(res, retrievedMetricEvent)
	if err != nil {
		return nil, err
	}

	return retrievedMetricEvent, nil
}

func (mec *MetricEventsClient) Create(metricEvent *MetricEvent) error {
	mePayload, err := json.Marshal(metricEvent)
	if err != nil {
		return fmt.Errorf("could not marshal metric event: %v", err)
	}

	_, err = mec.client.Post(metricEventsPath, mePayload)
	if err != nil {
		return fmt.Errorf("could not create metric event: %v", err)
	}

	return nil
}

func (mec *MetricEventsClient) Update(metricEvent *MetricEvent) error {
	mePayload, err := json.Marshal(metricEvent)
	if err != nil {
		return fmt.Errorf("could not marshal metric event: %v", err)
	}

	_, err = mec.client.Put(metricEventsPath, mePayload)
	if err != nil {
		return fmt.Errorf("could not create metric event: %v", err)
	}

	return nil
}

func (mec *MetricEventsClient) deleteByID(metricEventID string) error {
	_, err := mec.client.Delete(metricEventsPath + "/" + metricEventID)
	if err != nil {
		return fmt.Errorf("could not delete metric event with ID: %s, %v", metricEventID, err)
	}

	return nil
}

// GetMetricEventByName retrieves the MetricEvent identified by metricEventName, or nil if not found
func (mec *MetricEventsClient) GetMetricEventByName(metricEventName string) (*MetricEvent, error) {
	res, err := mec.getAll()
	if err != nil {
		log.WithError(err).Error("Could not get existing Dynatrace metric events")
		return nil, err
	}

	for _, metricEvent := range res.Values {
		if metricEvent.Name == metricEventName {
			res, err := mec.getByID(metricEvent.ID)
			if err != nil {
				log.WithError(err).WithField("eventKey", metricEventName).Error("Could not get existing metric event")
				return nil, err
			}

			return res, nil
		}
	}
	return nil, nil
}

func (mec *MetricEventsClient) DeleteMetricEventByName(metricEventName string) error {
	res, err := mec.getAll()
	if err != nil {
		log.WithError(err).Error("Could not get existing Dynatrace metric events")
		return err
	}

	for _, metricEvent := range res.Values {
		if metricEvent.Name == metricEventName {
			err := mec.deleteByID(metricEvent.ID)
			if err != nil {
				log.WithError(err).WithField("eventKey", metricEventName).Error("Could not delete existing metric event")
				return err
			}
		}
	}
	return nil
}

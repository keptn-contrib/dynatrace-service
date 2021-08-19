package dynatrace

import (
	"errors"
	"strings"
)

// PROBLEM NOTIFICATION

const PROBLEM_NOTIFICATION_PAYLOAD string = `{ 
      "type": "WEBHOOK", 
      "name": "Keptn Problem Notification", 
      "alertingProfile": "$ALERTING_PROFILE_ID", 
      "active": true, 
      "url": "$KEPTN_DNS/v1/event", 
      "acceptAnyCertificate": true, 
      "headers": [ 
        { "name": "x-token", "value": "$KEPTN_TOKEN" },
        { "name": "Content-Type", "value": "application/cloudevents+json" }
      ],
      "payload": "{\n    \"specversion\":\"1.0\",\n    \"type\":\"sh.keptn.events.problem\",\n    \"shkeptncontext\":\"{PID}\",\n    \"source\":\"dynatrace\",\n    \"id\":\"{PID}\",\n    \"time\":\"\",\n    \"contenttype\":\"application/json\",\n    \"data\": {\n        \"State\":\"{State}\",\n        \"ProblemID\":\"{ProblemID}\",\n        \"PID\":\"{PID}\",\n        \"ProblemTitle\":\"{ProblemTitle}\",\n        \"ProblemURL\":\"{ProblemURL}\",\n        \"ProblemDetails\":{ProblemDetailsJSON},\n        \"Tags\":\"{Tags}\",\n        \"ImpactedEntities\":{ImpactedEntities},\n        \"ImpactedEntity\":\"{ImpactedEntity}\"\n    }\n}\n" 

      }`

const KeptnProject = "keptn_project"
const KeptnStage = "keptn_stage"
const KeptnService = "keptn_service"
const KeptnDeployment = "keptn_deployment"

const ServiceEntityType = "SERVICE"

const DefaultOperatorVersion = "v0.8.0"
const SliResourceURI = "dynatrace/sli.yaml"
const Throughput = "throughput"
const ErrorRate = "error_rate"
const ResponseTimeP50 = "response_time_p50"
const ResponseTimeP90 = "response_time_p90"
const ResponseTimeP95 = "response_time_p95"

type CriteriaObject struct {
	Operator        string
	Value           float64
	CheckPercentage bool
	IsComparison    bool
	CheckIncrease   bool
}

type DTAPIListResponse struct {
	Values []Values `json:"values"`
}
type Values struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ConfigResult struct {
	Name    string
	Success bool
	Message string
}

// ConfiguredEntities contains information about the entities configures in Dynatrace
type ConfiguredEntities struct {
	TaggingRulesEnabled         bool
	TaggingRules                []ConfigResult
	ProblemNotificationsEnabled bool
	ProblemNotifications        ConfigResult
	ManagementZonesEnabled      bool
	ManagementZones             []ConfigResult
	DashboardEnabled            bool
	Dashboard                   ConfigResult
	MetricEventsEnabled         bool
	MetricEvents                []ConfigResult
}

// ALERTING PROFILE TYPES
type AlertingProfile struct {
	Metadata         AlertingProfileMetadata           `json:"metadata"`
	ID               string                            `json:"id"`
	DisplayName      string                            `json:"displayName"`
	Rules            []AlertingProfileRules            `json:"rules"`
	ManagementZoneID interface{}                       `json:"managementZoneId"`
	EventTypeFilters []*AlertingProfileEventTypeFilter `json:"eventTypeFilters,omitempty"`
}
type AlertingProfileMetadata struct {
	ConfigurationVersions []int  `json:"configurationVersions"`
	ClusterVersion        string `json:"clusterVersion"`
}
type AlertingProfileTagFilter struct {
	IncludeMode string   `json:"includeMode"`
	TagFilters  []string `json:"tagFilters"`
}
type AlertingProfileRules struct {
	SeverityLevel  string                   `json:"severityLevel"`
	TagFilter      AlertingProfileTagFilter `json:"tagFilter"`
	DelayInMinutes int                      `json:"delayInMinutes"`
}

type AlertingProfileEventTypeFilter struct {
	CustomEventFilter CustomEventFilter `json:"customEventFilter"`
}
type CustomTitleFilter struct {
	Enabled         bool   `json:"enabled"`
	Value           string `json:"value"`
	Operator        string `json:"operator"`
	Negate          bool   `json:"negate"`
	CaseInsensitive bool   `json:"caseInsensitive"`
}
type CustomEventFilter struct {
	CustomTitleFilter CustomTitleFilter `json:"customTitleFilter"`
}

// CALCULATED METRIC TYPES
type CalculatedMetric struct {
	TsmMetricKey        string                       `json:"tsmMetricKey"`
	Name                string                       `json:"name"`
	Enabled             bool                         `json:"enabled"`
	MetricDefinition    MetricDefinition             `json:"metricDefinition"`
	Unit                string                       `json:"unit"`
	UnitDisplayName     string                       `json:"unitDisplayName"`
	Conditions          []CalculatedMetricConditions `json:"conditions"`
	DimensionDefinition DimensionDefinition          `json:"dimensionDefinition"`
}
type MetricDefinition struct {
	Metric           string      `json:"metric"`
	RequestAttribute interface{} `json:"requestAttribute"`
}
type Value struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}
type CalculatedMetricConditions struct {
	Attribute      string                         `json:"attribute"`
	ComparisonInfo CalculatedMetricComparisonInfo `json:"comparisonInfo"`
}
type CalculatedMetricComparisonInfo struct {
	Type             string `json:"type"`
	Comparison       string `json:"comparison"`
	Value            Value  `json:"value"`
	Negate           bool   `json:"negate"`
	RequestAttribute string `json:"requestAttribute"`
	CaseSensitive    bool   `json:"caseSensitive"`
}
type DimensionDefinition struct {
	Name            string   `json:"name"`
	Dimension       string   `json:"dimension"`
	Placeholders    []string `json:"placeholders"`
	TopX            int      `json:"topX"`
	TopXDirection   string   `json:"topXDirection"`
	TopXAggregation string   `json:"topXAggregation"`
}

// MANAGEMENT ZONE TYPES
type ManagementZone struct {
	Name  string    `json:"name"`
	Rules []MZRules `json:"rules"`
}

type MZKey struct {
	Attribute string `json:"attribute"`
}
type MZValue struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}
type MZComparisonInfo struct {
	Type     string  `json:"type"`
	Operator string  `json:"operator"`
	Value    MZValue `json:"value"`
	Negate   bool    `json:"negate"`
}
type MZConditions struct {
	Key            MZKey            `json:"key"`
	ComparisonInfo MZComparisonInfo `json:"comparisonInfo"`
}
type MZRules struct {
	Type             string         `json:"type"`
	Enabled          bool           `json:"enabled"`
	PropagationTypes []string       `json:"propagationTypes"`
	Conditions       []MZConditions `json:"conditions"`
}

type DTDashboardsResponse struct {
	Dashboards []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Owner string `json:"owner"`
	} `json:"dashboards"`
}

// CUSTOM METRIC EVENT
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

var supportedAggregations = [...]string{"avg", "max", "min", "count", "sum", "value", "percentile"}

func CreateKeptnMetricEvent(project string, stage string, service string, metric string, query string, condition string, threshold float64, managementZoneID int64) (*MetricEvent, error) {

	/*
		need to map queries used by SLI-service to metric event definition.
		example: builtin:service.response.time:merge(0):percentile(90)?scope=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT)

		1. split by '?' and get first part => builtin:service.response.time:merge(0):percentile(90)
		2. split by ':' => builtin:service.response.time | merge(0) | percentile(90) => merge(0) is not needed
		3. first part is the metricId and can be used for the Metric Event API => builtin:service.response.time
		4. Aggregation is limited to: AVG, COUNT, MAX, MEDIAN, MIN, OF_INTEREST, OF_INTEREST_RATIO, OTHER, OTHER_RATIO, P90, SUM, VALUE
	*/

	if project == "" || stage == "" || service == "" || metric == "" || query == "" {
		return nil, errors.New("missing input parameter values")
	}

	query = strings.TrimPrefix(query, "metricSelector=")
	// 1. split by '?' and get first part => builtin:service.response.time:merge(0):percentile(90)
	split := strings.Split(query, "?")

	// 2. split by ':' => builtin:service.response.time | merge(0) | percentile(90) => merge(0) is not needed/supported by MetricEvent API
	splittedQuery := strings.Split(split[0], ":")

	if len(splittedQuery) < 2 {
		return nil, errors.New("invalid metricId")
	}
	metricId := splittedQuery[0] + ":" + splittedQuery[1]
	meAggregation := ""
	for _, transformation := range splittedQuery {
		isSupportedAggregation := false
		for _, aggregationType := range supportedAggregations {
			if strings.Contains(strings.ToLower(transformation), aggregationType) {
				isSupportedAggregation = true
			}
		}

		if isSupportedAggregation {
			meAggregation = getMetricEventAggregation(transformation)

			/*
				if meAggregation == "" {
					return nil, errors.New("unsupported aggregation type: " + transformation)
				}

			*/
		}
	}
	/*
		if meAggregation == "" {
			return nil, errors.New("no aggregation provided in query")
		}
	*/

	meAlertCondition, err := getAlertCondition(condition)
	if err != nil {
		return nil, err
	}

	metricEvent := &MetricEvent{
		Metadata:          MEMetadata{},
		MetricID:          metricId,
		Name:              metric + " (Keptn." + project + "." + stage + "." + service + ")",
		Description:       "Keptn SLI violated: The {metricname} value of {severity} was {alert_condition} your custom threshold of {threshold}.",
		EventType:         "CUSTOM_ALERT",
		Severity:          "CUSTOM_ALERT",
		AlertCondition:    meAlertCondition,
		Samples:           5, // taken from default value of custom metric events
		ViolatingSamples:  3, // taken from default value of custom metric events
		DealertingSamples: 5, // taken from default value of custom metric events
		Threshold:         threshold,
		Enabled:           false,
		TagFilters:        nil, // not used anymore by MetricEvents API, replaced by AlertingScope
		AlertingScope: []MEAlertingScope{
			// LIMITATION: currently only a maximum of 3 tag filters is supported
			{
				FilterType:       "MANAGEMENT_ZONE",
				ManagementZoneID: managementZoneID,
			},
			{
				FilterType: "TAG",
				TagFilter: &METagFilter{
					Context: "CONTEXTLESS",
					Key:     KeptnService,
					Value:   service,
				},
			},
			{
				FilterType: "TAG",
				TagFilter: &METagFilter{
					Context: "CONTEXTLESS",
					Key:     KeptnDeployment,
					Value:   "primary",
				},
			},
		},
	}

	// LIMITATION: currently we do not have the possibility of specifying units => assume MILLI_SECONDS for response time metrics
	if strings.Contains(metric, "time") {
		metricEvent.Unit = "MILLI_SECOND"
	}

	if meAggregation != "" {
		metricEvent.AggregationType = meAggregation
	}

	return metricEvent, nil
}

func getAlertCondition(condition string) (string, error) {
	meAlertCondition := ""
	if strings.Contains(condition, "+") || strings.Contains(condition, "-") || strings.Contains(condition, "%") {
		return "", errors.New("unsupported condition. only fixed thresholds are supported")
	}

	if strings.Contains(condition, ">") {
		meAlertCondition = "BELOW"
	} else if strings.Contains(condition, "<") {
		meAlertCondition = "ABOVE"
	} else {
		return "", errors.New("unsupported condition. only fixed thresholds are supported")
	}
	return meAlertCondition, nil
}

func getMetricEventAggregation(metricAPIAgg string) string {
	// LIMITATION: currently, only single aggregations are supported, so, e.g. not (min,max)
	metricAPIAgg = strings.ToLower(metricAPIAgg)

	if strings.Contains(metricAPIAgg, "percentile") {
		// only MEDIAN and P90 are supported for MetricEvents
		// => if the percentile in the query is >= 90, use P90, otherwise assume MEDIAN
		if strings.Contains(metricAPIAgg, "(9") {
			return "P90"
		} else {
			return "MEDIAN"
		}
	}
	// due to incompatibilities between metrics and metric event API it's safer to not pass an aggregation in the MetricEvent definition in most cases
	// the Metric Event API will default it to an appropriate aggregation
	/*else if strings.Contains(metricAPIAgg, "min") {
		return "MIN"
	} else if strings.Contains(metricAPIAgg, "max") {
		return "MAX"
	} else if strings.Contains(metricAPIAgg, "count") {
		return "COUNT"
	} else if strings.Contains(metricAPIAgg, "sum") {
		return "SUM"
	} else if strings.Contains(metricAPIAgg, "value") {
		return "VALUE"
	} else if strings.Contains(metricAPIAgg, "avg") {
		return "AVG"
	}
	*/
	return ""
}

func CreateKeptnAlertingProfile() *AlertingProfile {
	return &AlertingProfile{
		Metadata:    AlertingProfileMetadata{},
		DisplayName: "Keptn",
		Rules: []AlertingProfileRules{
			createAlertingProfileRule("AVAILABILITY"),
			createAlertingProfileRule("ERROR"),
			createAlertingProfileRule("PERFORMANCE"),
			createAlertingProfileRule("RESOURCE_CONTENTION"),
			createAlertingProfileRule("CUSTOM_ALERT"),
			createAlertingProfileRule("MONITORING_UNAVAILABLE"),
		},
		ManagementZoneID: nil,
	}
}

func createAlertingProfileRule(severityLevel string) AlertingProfileRules {
	return AlertingProfileRules{
		SeverityLevel: severityLevel,
		TagFilter: AlertingProfileTagFilter{
			IncludeMode: "NONE",
			TagFilters:  nil,
		},
		DelayInMinutes: 0,
	}
}

func CreateManagementZoneForProject(project string) *ManagementZone {
	managementZone := &ManagementZone{
		Name: "Keptn: " + project,
		Rules: []MZRules{
			{
				Type:             ServiceEntityType,
				Enabled:          true,
				PropagationTypes: []string{},
				Conditions:       []MZConditions{creteManagementZoneConditionsFor(KeptnProject, project)},
			},
		},
	}

	return managementZone
}

func CreateManagementZoneForStage(project string, stage string) *ManagementZone {
	managementZone := &ManagementZone{
		Name: "Keptn: " + project + " " + stage,
		Rules: []MZRules{
			{
				Type:             ServiceEntityType,
				Enabled:          true,
				PropagationTypes: []string{},
				Conditions: []MZConditions{
					creteManagementZoneConditionsFor(KeptnProject, project),
					creteManagementZoneConditionsFor(KeptnStage, stage),
				},
			},
		},
	}

	return managementZone
}

func creteManagementZoneConditionsFor(key string, value string) MZConditions {
	return MZConditions{
		Key: MZKey{
			Attribute: "SERVICE_TAGS",
		},
		ComparisonInfo: MZComparisonInfo{
			Type:     "TAG",
			Operator: "EQUALS",
			Value: MZValue{
				Context: "CONTEXTLESS",
				Key:     key,
				Value:   value,
			},
			Negate: false,
		},
	}
}

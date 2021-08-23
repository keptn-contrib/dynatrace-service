package dynatrace

const KeptnProject = "keptn_project"
const KeptnStage = "keptn_stage"

const ServiceEntityType = "SERVICE"

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

func (response *DTAPIListResponse) ToStringSetWith(mapper func(Values) string) *StringSet {
	stringSet := &StringSet{
		values: make(map[string]struct{}, len(response.Values)),
	}
	for _, rule := range response.Values {
		stringSet.values[mapper(rule)] = struct{}{}
	}

	return stringSet
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

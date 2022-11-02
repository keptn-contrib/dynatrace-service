package dynatrace

const KeptnProject = "keptn_project"
const KeptnStage = "keptn_stage"

const ServiceEntityType = "SERVICE"

type listResponse struct {
	Values []values `json:"values"`
}
type values struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (response *listResponse) ToStringSetWith(mapper func(values) string) *StringSet {
	stringSet := &StringSet{
		values: make(map[string]struct{}, len(response.Values)),
	}
	for _, rule := range response.Values {
		stringSet.values[mapper(rule)] = struct{}{}
	}

	return stringSet
}

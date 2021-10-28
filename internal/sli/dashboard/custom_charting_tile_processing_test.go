package dashboard

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetEntitySelectorFromEntityFilter(t *testing.T) {
	expected := ",entityId(\"SERVICE-086C46F600BA1DC6\"),tag(\"keptn_deployment:primary\")"

	var filtersPerEntityType = map[string]dynatrace.FilterMap{
		"SERVICE": {
			"SPECIFIC_ENTITIES": {"SERVICE-086C46F600BA1DC6"},
			"AUTO_TAGS":         {"keptn_deployment:primary"},
		},
	}
	actual := getEntitySelectorFromEntityFilter(filtersPerEntityType, "SERVICE")

	assert.Equal(t, expected, actual)
}

package query

import (
	"github.com/keptn-contrib/dynatrace-service/internal/sli/unit"
	"testing"
)

func TestScaleData(t *testing.T) {
	if unit.ScaleData("", "MicroSecond", 1000000.0) != 1000.0 {
		t.Errorf("ScaleData incorrectly scales MicroSecond")
	}
	if unit.ScaleData("", "Byte", 1024.0) != 1.0 {
		t.Errorf("ScaleData incorrectly scales Bytes")
	}
	if unit.ScaleData("builtin:service.response.time", "", 1000000.0) != 1000.0 {
		t.Errorf("ScaleData incorrectly scales builtin:service.response.time")
	}
}

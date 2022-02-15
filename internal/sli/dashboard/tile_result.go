package dashboard

import (
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type TileResult struct {
	sliResult *keptnv2.SLIResult
	objective *keptnapi.SLO
	sliName   string
	sliQuery  string
}

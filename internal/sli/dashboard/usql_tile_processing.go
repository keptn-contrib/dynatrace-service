package dashboard

import (
	"fmt"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
	keptncommon "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

type USQLTileProcessing struct {
	client        dynatrace.ClientInterface
	eventData     adapter.EventContentAdapter
	customFilters []*keptnv2.SLIFilter
	startUnix     time.Time
	endUnix       time.Time
}

func NewUSQLTileProcessing(client dynatrace.ClientInterface, eventData adapter.EventContentAdapter, customFilters []*keptnv2.SLIFilter, startUnix time.Time, endUnix time.Time) *USQLTileProcessing {
	return &USQLTileProcessing{
		client:        client,
		eventData:     eventData,
		customFilters: customFilters,
		startUnix:     startUnix,
		endUnix:       endUnix,
	}
}

func (p *USQLTileProcessing) Process(tile *dynatrace.Tile) []*TileResult {
	// for Dynatrace Query Language we currently support the following
	// SINGLE_VALUE: we just take the one value that comes back
	// PIE_CHART, COLUMN_CHART: we assume the first column is the dimension and the second column is the value column
	// TABLE: we assume the first column is the dimension and the last is the value
	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(dynatrace.NewUSQLClientQueryParameters(usql.NewQuery(tile.Query), p.startUnix, p.endUnix))
	if err != nil {
		log.WithError(err).Warn("executeGetDynatraceUSQLQuery returned an error")
		return nil
	}

	tileTitle := tile.Title()

	// first - lets figure out if this tile should be included in SLI validation or not - we parse the title and look for "sli=sliname"
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(tileTitle)
	if sloDefinition.SLI == "" {
		log.WithField("tileTitle", tileTitle).Debug("Tile not included as name doesnt include sli=SLINAME")
		return nil
	}

	var tileResults []*TileResult

	for _, rowValue := range usqlResult.Values {
		dimensionName := ""
		dimensionValue := 0.0

		switch tile.Type {
		case "SINGLE_VALUE":
			dimensionValue = rowValue[0].(float64)
		case "PIE_CHART":
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		case "COLUMN_CHART":
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[1].(float64)
		case "TABLE":
			dimensionName = rowValue[0].(string)
			dimensionValue = rowValue[len(rowValue)-1].(float64)
		default:
			log.WithField("tileType", tile.Type).Debug("Unsupport USQL tile type")
			continue
		}

		// lets scale the metric
		// value = scaleData(metricDefinition.MetricID, metricDefinition.Unit, value)

		// we got our metric, slos and the value
		indicatorName := sloDefinition.SLI
		if dimensionName != "" {
			indicatorName = indicatorName + "_" + dimensionName
		}

		log.WithFields(
			log.Fields{
				"name":           indicatorName,
				"dimensionValue": dimensionValue,
			}).Debug("Appending SLIResult")

		// add this to our SLI Indicator JSON in case we need to generate an SLI.yaml
		// in that case we also need to mask it with USQL, TITLE_TYPE, DIMENSIONNAME
		// we also add the SLO definition in case we need to generate an SLO.yaml
		tileResults = append(
			tileResults,
			&TileResult{
				sliResult: &keptnv2.SLIResult{
					Metric:  indicatorName,
					Value:   dimensionValue,
					Success: true,
				},
				objective: &keptncommon.SLO{
					SLI:     indicatorName,
					Weight:  sloDefinition.Weight,
					KeySLI:  sloDefinition.KeySLI,
					Pass:    sloDefinition.Pass,
					Warning: sloDefinition.Warning,
				},
				sliName:  indicatorName,
				sliQuery: fmt.Sprintf("USQL;%s;%s;%s", tile.Type, dimensionName, tile.Query),
			})
	}

	return tileResults
}

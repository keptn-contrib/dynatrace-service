package dashboard

import (
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/adapter"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/usql"
	v1usql "github.com/keptn-contrib/dynatrace-service/internal/sli/v1/usql"
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
	// first - lets figure out if this tile should be included in SLI validation or not - we parse the title and look for "sli=sliname"
	sloDefinition := common.ParsePassAndWarningWithoutDefaultsFrom(tile.Title())
	if sloDefinition.SLI == "" {
		log.WithField("tileTitle", tile.Title()).Debug("Tile not included as name doesnt include sli=SLINAME")
		return nil
	}

	// for Dynatrace Query Language we currently support the following
	// SINGLE_VALUE: we just take the one value that comes back
	// PIE_CHART, COLUMN_CHART: we assume the first column is the dimension and the second column is the value column
	// TABLE: we assume the first column is the dimension and the last is the value
	query, err := usql.NewQuery(tile.Query)
	if err != nil {
		unsuccessfulTileResult := newUnsuccessfulTileResultFromSLODefinition(sloDefinition, "could not create USQL query: "+err.Error())
		return []*TileResult{&unsuccessfulTileResult}
	}

	usqlResult, err := dynatrace.NewUSQLClient(p.client).GetByQuery(dynatrace.NewUSQLClientQueryParameters(*query, p.startUnix, p.endUnix))
	if err != nil {
		unsuccessfulTileResult := newUnsuccessfulTileResultFromSLODefinition(sloDefinition, "error executing USQL query: "+err.Error())
		return []*TileResult{&unsuccessfulTileResult}
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

		tileResult := createTileResultForDimensionNameAndValue(dimensionName, dimensionValue, sloDefinition, tile.Type, *query)
		tileResults = append(tileResults, &tileResult)
	}

	return tileResults
}

func createTileResultForDimensionNameAndValue(dimensionName string, dimensionValue float64, sloDefinition *keptncommon.SLO, tileType string, baseQuery usql.Query) TileResult {
	indicatorName := sloDefinition.SLI
	if dimensionName != "" {
		indicatorName = indicatorName + "_" + dimensionName
	}

	v1USQLQuery, err := v1usql.NewQuery(tileType, dimensionName, baseQuery)
	if err != nil {
		return newUnsuccessfulTileResultFromSLODefinition(sloDefinition, "could not create USQL v1 query: "+err.Error())
	}

	return TileResult{
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
		sliQuery: v1usql.NewQueryProducer(*v1USQLQuery).Produce(),
	}
}

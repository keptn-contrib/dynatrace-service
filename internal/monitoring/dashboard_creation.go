package monitoring

import (
	"context"
	"fmt"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

const dashboardNameSuffix = "@keptn: Digital Delivery & Operations Dashboard"

const customChartingTileType = "CUSTOM_CHARTING"
const customChartName = "Custom Chart"
const timeSeriesChartType = "TIMESERIES"
const dashboardStageWidth int = 456

type dashboardCreation struct {
	client dynatrace.ClientInterface
}

func newDashboardCreation(client dynatrace.ClientInterface) *dashboardCreation {
	return &dashboardCreation{
		client: client,
	}
}

// create creates a new dashboard for the provided project.
func (dc *dashboardCreation) create(ctx context.Context, project string, shipyard keptnv2.Shipyard) *configResult {
	// first, check if dashboard for this project already exists and delete that
	dashboardClient := dynatrace.NewDashboardsClient(dc.client)
	err := deleteExistingDashboard(ctx, project, dashboardClient)
	if err != nil {
		log.WithError(err).Error("Could not delete existing dashboard")
		return &configResult{
			Success: false,
			Message: "Could not delete existing dashboard: " + err.Error(),
		}
	}

	log.WithField("project", project).Info("Creating Dashboard for project")
	dashboard := createDynatraceDashboard(project, shipyard)
	err = dashboardClient.Create(ctx, dashboard)
	if err != nil {
		log.WithError(err).Error("Failed to create Dynatrace dashboards")
		return &configResult{
			Success: false,
			Message: err.Error(),
		}
	}
	log.WithField("dashboardUrl", dc.client.Credentials().GetTenant()+"/#dashboards").Info("Dynatrace dashboard created successfully")
	return &configResult{
		Success: true, // I guess this should be true not false?
		Message: "Dynatrace dashboard created successfully. You can view it here: " + dc.client.Credentials().GetTenant() + "/#dashboards",
	}
}

// deleteExistingDashboard deletes an existing dashboard for the provided project
func deleteExistingDashboard(ctx context.Context, project string, dashboardClient *dynatrace.DashboardsClient) error {
	response, err := dashboardClient.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, dashboardItem := range response.Dashboards {
		if dashboardItem.Name == getDashboardName(project) {
			err = dashboardClient.Delete(ctx, dashboardItem.ID)
			if err != nil {
				return fmt.Errorf("could not delete dashboard for project %s: %v", project, err)
			}
		}
	}
	return nil
}

func getDashboardName(projectName string) string {
	return projectName + dashboardNameSuffix
}

// Dashboard creation stuff below

func createDynatraceDashboard(projectName string, shipyard keptnv2.Shipyard) *dynatrace.Dashboard {
	dtDashboard := &dynatrace.Dashboard{
		DashboardMetadata: dynatrace.DashboardMetadata{
			Name:   getDashboardName(projectName),
			Shared: true,
			Owner:  "",
			SharingDetails: dynatrace.SharingDetails{
				LinkShared: true,
				Published:  false,
			},
			DashboardFilter: &dynatrace.DashboardFilter{
				Timeframe:      "l_7_DAYS",
				ManagementZone: nil,
			},
		},
		Tiles: []dynatrace.Tile{},
	}

	infrastructureHeaderTile := createHeaderTile("Infrastructure")
	infrastructureHeaderTile.Bounds = dynatrace.Bounds{
		Top:    0,
		Left:   0,
		Width:  494,
		Height: 38,
	}
	dtDashboard.Tiles = append(dtDashboard.Tiles, infrastructureHeaderTile)

	hostsTile := createTileWith(
		"",
		"HOSTS",
		&dynatrace.FilterConfig{
			Type:        "HOST",
			CustomName:  "Hosts",
			DefaultName: "Hosts",
			ChartConfig: dynatrace.ChartConfig{
				Type:           timeSeriesChartType,
				Series:         []dynatrace.Series{},
				ResultMetadata: dynatrace.ResultMetadata{},
			},
			FiltersPerEntityType: map[string]dynatrace.FilterMap{},
		})
	hostsTile.Bounds = createBounds(38, 0, 152)
	dtDashboard.Tiles = append(dtDashboard.Tiles, hostsTile)

	networkTile := dynatrace.Tile{
		Name:             "Network Status",
		TileType:         "NETWORK_MEDIUM",
		Configured:       true,
		TileFilter:       dynatrace.TileFilter{},
		AssignedEntities: nil,
		Bounds:           createBounds(38, 912, 152),
	}
	dtDashboard.Tiles = append(dtDashboard.Tiles, networkTile)

	cpuLoadTile := createHostCPULoadTile()
	cpuLoadTile.Bounds = createBounds(38, dashboardStageWidth, 152)
	dtDashboard.Tiles = append(dtDashboard.Tiles, cpuLoadTile)

	// create stage service tiles
	for index, stage := range shipyard.Spec.Stages {

		headerTile := createHeaderTile(stage.Name)
		headerTile.Bounds = createBounds(266, index*dashboardStageWidth, 38)

		servicesTile := createStageServicesTile(projectName, stage.Name)
		servicesTile.Bounds = createStandardTileBounds(304, index*dashboardStageWidth)

		throughputTile := createServiceThroughputTile(projectName, stage.Name)
		throughputTile.Bounds = createStandardTileBounds(456, index*dashboardStageWidth)

		errorRateTile := createServiceErrorRateTile(projectName, stage.Name)
		errorRateTile.Bounds = createStandardTileBounds(608, index*dashboardStageWidth)

		responseTimeTile := createServiceResponseTimeTile(projectName, stage.Name)
		responseTimeTile.Bounds = createStandardTileBounds(760, index*dashboardStageWidth)

		dtDashboard.Tiles = append(dtDashboard.Tiles, headerTile, servicesTile, throughputTile, errorRateTile, responseTimeTile)
	}

	return dtDashboard
}

func createStandardTileBounds(top int, left int) dynatrace.Bounds {
	return createBounds(top, left, 152)
}

func createBounds(top int, left int, height int) dynatrace.Bounds {
	return dynatrace.Bounds{
		Top:    top,
		Left:   left,
		Width:  dashboardStageWidth,
		Height: height,
	}
}

func createHeaderTile(name string) dynatrace.Tile {
	return createTileWith(name, "HEADER", nil)
}

func createServiceResponseTimeTile(project string, stage string) dynatrace.Tile {
	name := "Response Time " + stage
	return createTileWith(
		name,
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:                 "MIXED",
			CustomName:           name,
			DefaultName:          customChartName,
			ChartConfig:          createTimeSeriesChartConfig("builtin:service.response.time", "AVG", "LINE", dynatrace.ServiceEntityType),
			FiltersPerEntityType: createServiceAutoTagsEntityFilter(project, stage),
		})
}

func createHostCPULoadTile() dynatrace.Tile {
	return createTileWith(
		"Host CPU Load",
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:                 "MIXED",
			CustomName:           "CPU",
			DefaultName:          customChartName,
			ChartConfig:          createTimeSeriesChartConfig("builtin:host.cpu.load", "AVG", "LINE", "HOST"),
			FiltersPerEntityType: map[string]dynatrace.FilterMap{},
		})
}

func createServiceErrorRateTile(project string, stage string) dynatrace.Tile {
	name := "Failure Rate " + stage
	return createTileWith(
		name,
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:                 "MIXED",
			CustomName:           name,
			DefaultName:          customChartName,
			ChartConfig:          createTimeSeriesChartConfig("builtin:service.errors.server.rate", "AVG", "BAR", dynatrace.ServiceEntityType),
			FiltersPerEntityType: createServiceAutoTagsEntityFilter(project, stage),
		})
}

func createServiceThroughputTile(project string, stage string) dynatrace.Tile {
	name := "Throughput " + stage
	return createTileWith(
		name,
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:                 "MIXED",
			CustomName:           name,
			DefaultName:          customChartName,
			ChartConfig:          createTimeSeriesChartConfig("builtin:service.requestCount.total", "NONE", "BAR", dynatrace.ServiceEntityType),
			FiltersPerEntityType: createServiceAutoTagsEntityFilter(project, stage),
		})
}

func createTimeSeriesChartConfig(metric string, aggregation string, seriesType string, entity string) dynatrace.ChartConfig {
	return dynatrace.ChartConfig{
		Type: timeSeriesChartType,
		Series: []dynatrace.Series{
			{
				Metric:          metric,
				Aggregation:     aggregation,
				Percentile:      nil,
				Type:            seriesType,
				EntityType:      entity,
				Dimensions:      []dynatrace.Dimensions{},
				SortAscending:   false,
				SortColumn:      true,
				AggregationRate: "TOTAL",
			},
		},
	}
}

func createStageServicesTile(project string, stage string) dynatrace.Tile {
	name := "Services: " + stage
	return createTileWith(
		name,
		"SERVICES",
		&dynatrace.FilterConfig{
			Type:        dynatrace.ServiceEntityType,
			CustomName:  name,
			DefaultName: name,
			ChartConfig: dynatrace.ChartConfig{
				Type:           timeSeriesChartType,
				Series:         []dynatrace.Series{},
				ResultMetadata: dynatrace.ResultMetadata{},
			},
			FiltersPerEntityType: createServiceAutoTagsEntityFilter(project, stage),
		})
}

func createTileWith(name string, tileType string, filterConfig *dynatrace.FilterConfig) dynatrace.Tile {
	return dynatrace.Tile{
		Name:                      name,
		TileType:                  tileType,
		Configured:                true,
		TileFilter:                dynatrace.TileFilter{},
		FilterConfig:              filterConfig,
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
		Markdown:                  "",
	}
}

func createServiceAutoTagsEntityFilter(project string, stage string) map[string]dynatrace.FilterMap {
	const service = "SERVICE"
	const autoTags = "AUTO_TAGS"

	result := make(map[string]dynatrace.FilterMap)
	result[service] = make(map[string][]string)
	result[service][autoTags] = []string{getKeptnProjectTag(project), getKeptnStageTag(stage)}

	return result
}

func getTag(name string, value string) string {
	return name + ":" + value
}

func getKeptnProjectTag(value string) string {
	return getTag(dynatrace.KeptnProject, value)
}

func getKeptnStageTag(value string) string {
	return getTag(dynatrace.KeptnStage, value)
}

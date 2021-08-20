package monitoring

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	log "github.com/sirupsen/logrus"
)

const dashboardNameSuffix = "@keptn: Digital Delivery & Operations Dashboard"

const customChartingTileType = "CUSTOM_CHARTING"
const customChartName = "Custom Chart"
const timeSeriesChartType = "TIMESERIES"
const dashboardStageWidth int = 456

type DashboardCreation struct {
	client *dynatrace.DynatraceHelper
}

func NewDashboardCreation(client *dynatrace.DynatraceHelper) *DashboardCreation {
	return &DashboardCreation{
		client: client,
	}
}

// Create creates a new dashboard for the provided project
func (dc *DashboardCreation) Create(project string, shipyard keptnv2.Shipyard) dynatrace.ConfigResult {
	if !lib.IsDashboardsGenerationEnabled() {
		return dynatrace.ConfigResult{}
	}

	// first, check if dashboard for this project already exists and delete that
	dashboardClient := dynatrace.NewDashboardsClient(dc.client)
	err := deleteExistingDashboard(project, dashboardClient)
	if err != nil {
		log.WithError(err).Error("Could not delete existing dashboard")
		return dynatrace.ConfigResult{
			Success: false,
			Message: "Could not delete existing dashboard: " + err.Error(),
		}
	}

	log.WithField("project", project).Info("Creating Dashboard for project")
	dashboard := createDynatraceDashboard(project, shipyard)
	_, err = dashboardClient.Create(dashboard)
	if err != nil {
		log.WithError(err).Error("Failed to create Dynatrace dashboards")
		return dynatrace.ConfigResult{
			Success: false,
			Message: err.Error(),
		}
	}
	log.WithField("dashboardUrl", "https://"+dc.client.DynatraceCreds.Tenant+"/#dashboards").Info("Dynatrace dashboard created successfully")
	return dynatrace.ConfigResult{
		Success: true, // I guess this should be true not false?
		Message: "Dynatrace dashboard created successfully. You can view it here: https://" + dc.client.DynatraceCreds.Tenant + "/#dashboards",
	}
}

// deleteExistingDashboard deletes an existing dashboard for the provided project
func deleteExistingDashboard(project string, dashboardClient *dynatrace.DashboardsClient) error {
	response, err := dashboardClient.GetAll()
	if err != nil {
		return err
	}

	for _, dashboardItem := range response.Dashboards {
		if dashboardItem.Name == createDashboardNameFor(project) {
			_, err = dashboardClient.Delete(dashboardItem.ID)
			if err != nil {
				return fmt.Errorf("could not delete dashboard for project %s: %v", project, err)
			}
		}
	}
	return nil
}

func createDashboardNameFor(projectName string) string {
	return projectName + dashboardNameSuffix
}

// Dashboard creation stuff below

func createDynatraceDashboard(projectName string, shipyard keptnv2.Shipyard) *dynatrace.DynatraceDashboard {
	dtDashboard := &dynatrace.DynatraceDashboard{
		DashboardMetadata: dynatrace.DashboardMetadata{
			Name:   createDashboardNameFor(projectName),
			Shared: true,
			Owner:  "",
			SharingDetails: dynatrace.SharingDetails{
				LinkShared: true,
				Published:  false,
			},
			DashboardFilter: dynatrace.DashboardFilter{
				Timeframe:      "l_7_DAYS",
				ManagementZone: nil,
			},
		},
		Tiles: []dynatrace.Tiles{},
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
			FiltersPerEntityType: dynatrace.FiltersPerEntityType{},
		})
	hostsTile.Bounds = createBounds(38, 0, 152)
	dtDashboard.Tiles = append(dtDashboard.Tiles, hostsTile)

	networkTile := dynatrace.Tiles{
		Name:       "Network Status",
		TileType:   "NETWORK_MEDIUM",
		Configured: true,
		TileFilter: dynatrace.TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
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

func createHeaderTile(name string) dynatrace.Tiles {
	return createTileWith(name, "HEADER", nil)
}

func createServiceResponseTimeTile(project string, stage string) dynatrace.Tiles {
	name := "Response Time " + stage
	return createTileWith(
		name,
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:        "MIXED",
			CustomName:  name,
			DefaultName: customChartName,
			ChartConfig: createTimeSeriesChartConfig("builtin:service.response.time", "AVG", "LINE", dynatrace.ServiceEntityType),
			FiltersPerEntityType: dynatrace.FiltersPerEntityType{
				Service: &dynatrace.EntityFilter{
					AutoTags: []string{createKeptnProjectTag(project), createKeptnStageTag(stage)},
				},
			},
		})
}

func createHostCPULoadTile() dynatrace.Tiles {
	return createTileWith(
		"Host CPU Load",
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:        "MIXED",
			CustomName:  "CPU",
			DefaultName: customChartName,
			ChartConfig: createTimeSeriesChartConfig("builtin:host.cpu.load", "AVG", "LINE", "HOST"),
		})
}

func createServiceErrorRateTile(project string, stage string) dynatrace.Tiles {
	name := "Failure Rate " + stage
	return createTileWith(
		name,
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:        "MIXED",
			CustomName:  name,
			DefaultName: customChartName,
			ChartConfig: createTimeSeriesChartConfig("builtin:service.errors.server.rate", "AVG", "BAR", dynatrace.ServiceEntityType),
			FiltersPerEntityType: dynatrace.FiltersPerEntityType{
				Service: &dynatrace.EntityFilter{
					AutoTags: []string{createKeptnProjectTag(project), createKeptnStageTag(stage)},
				},
			},
		})
}

func createServiceThroughputTile(project string, stage string) dynatrace.Tiles {
	name := "Throughput " + stage
	return createTileWith(
		name,
		customChartingTileType,
		&dynatrace.FilterConfig{
			Type:        "MIXED",
			CustomName:  name,
			DefaultName: customChartName,
			ChartConfig: createTimeSeriesChartConfig("builtin:service.requestCount.total", "NONE", "BAR", dynatrace.ServiceEntityType),
			FiltersPerEntityType: dynatrace.FiltersPerEntityType{
				Service: &dynatrace.EntityFilter{
					AutoTags: []string{createKeptnProjectTag(project), createKeptnStageTag(stage)},
				},
			},
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

func createStageServicesTile(project string, stage string) dynatrace.Tiles {
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
			FiltersPerEntityType: dynatrace.FiltersPerEntityType{
				Service: &dynatrace.EntityFilter{
					AutoTags: []string{createKeptnProjectTag(project), createKeptnStageTag(stage)},
				},
			},
		})
}

func createTileWith(name string, tileType string, filterConfig *dynatrace.FilterConfig) dynatrace.Tiles {
	return dynatrace.Tiles{
		Name:       name,
		TileType:   tileType,
		Configured: true,
		TileFilter: dynatrace.TileFilter{
			Timeframe:      nil,
			ManagementZone: nil,
		},
		FilterConfig:              filterConfig,
		ChartVisible:              true,
		AssignedEntities:          nil,
		ExcludeMaintenanceWindows: false,
		Markdown:                  "",
	}
}

func createTagFor(name string, value string) string {
	return name + ":" + value
}

func createKeptnProjectTag(value string) string {
	return createTagFor(dynatrace.KeptnProject, value)
}

func createKeptnStageTag(value string) string {
	return createTagFor(dynatrace.KeptnStage, value)
}

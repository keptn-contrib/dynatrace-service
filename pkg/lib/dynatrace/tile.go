package dynatrace

type Tile struct {
	Name       string `json:"name"`
	TileType   string `json:"tileType"`
	Configured bool   `json:"configured"`
	Query      string `json:"query"`
	Type       string `json:"type"`
	CustomName string `json:"customName"`
	Markdown   string `json:"markdown"`
	Bounds     struct {
		Top    int `json:"top"`
		Left   int `json:"left"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"bounds"`
	TileFilter struct {
		Timeframe      string          `json:"timeframe"`
		ManagementZone *ManagementZone `json:"managementZone,omitempty"`
	} `json:"tileFilter"`
	Queries          []DataExplorerQuery `json:"queries"`
	AssignedEntities []string            `json:"assignedEntities"`
	FilterConfig     struct {
		Type        string `json:"type"`
		CustomName  string `json:"customName"`
		DefaultName string `json:"defaultName"`
		ChartConfig struct {
			LegendShown    bool          `json:"legendShown"`
			Type           string        `json:"type"`
			Series         []ChartSeries `json:"series"`
			ResultMetadata struct {
			} `json:"resultMetadata"`
		} `json:"chartConfig"`
		FiltersPerEntityType map[string]map[string][]string `json:"filtersPerEntityType"`
		/* FiltersPerEntityType struct {
			HOST struct {
				SPECIFIC_ENTITIES    []string `json:"SPECIFIC_ENTITIES"`
				HOST_DATACENTERS     []string `json:"HOST_DATACENTERS"`
				AUTO_TAGS            []string `json:"AUTO_TAGS"`
				HOST_SOFTWARE_TECH   []string `json:"HOST_SOFTWARE_TECH"`
				HOST_VIRTUALIZATION  []string `json:"HOST_VIRTUALIZATION"`
				HOST_MONITORING_MODE []string `json:"HOST_MONITORING_MODE"`
				HOST_STATE           []string `json:"HOST_STATE"`
				HOST_HOST_GROUPS     []string `json:"HOST_HOST_GROUPS"`
			} `json:"HOST"`
			PROCESS_GROUP struct {
				SPECIFIC_ENTITIES     []string `json:"SPECIFIC_ENTITIES"`
				HOST_TAG_OF_PROCESS   []string `json:"HOST_TAG_OF_PROCESS"`
				AUTO_TAGS             []string `json:"AUTO_TAGS"`
				PROCESS_SOFTWARE_TECH []string `json:"PROCESS_SOFTWARE_TECH"`
			} `json:"PROCESS_GROUP"`
			PROCESS_GROUP_INSTANCE struct {
				SPECIFIC_ENTITIES     []string `json:"SPECIFIC_ENTITIES"`
				HOST_TAG_OF_PROCESS   []string `json:"HOST_TAG_OF_PROCESS"`
				AUTO_TAGS             []string `json:"AUTO_TAGS"`
				PROCESS_SOFTWARE_TECH []string `json:"PROCESS_SOFTWARE_TECH"`
			} `json:"PROCESS_GROUP_INSTANCE"`
			SERVICE struct {
				SPECIFIC_ENTITIES     []string `json:"SPECIFIC_ENTITIES"`
				SERVICE_SOFTWARE_TECH []string `json:"SERVICE_SOFTWARE_TECH"`
				AUTO_TAGS             []string `json:"AUTO_TAGS"`
				SERVICE_TYPE          []string `json:"SERVICE_TYPE"`
				SERVICE_TO_PG         []string `json:"SERVICE_TO_PG"`
			} `json:"SERVICE"`
			APPLICATION struct {
				SPECIFIC_ENTITIES          []string `json:"SPECIFIC_ENTITIES"`
				APPLICATION_TYPE           []string `json:"APPLICATION_TYPE"`
				AUTO_TAGS                  []string `json:"AUTO_TAGS"`
				APPLICATION_INJECTION_TYPE []string `json:"PROCESS_SOFTWARE_TECH"`
				APPLICATION_STATUS         []string `json:"APPLICATION_STATUS"`
			} `json:"APPLICATION"`
			APPLICATION_METHOD struct {
				SPECIFIC_ENTITIES []string `json:"SPECIFIC_ENTITIES"`
			} `json:"APPLICATION_METHOD"`
		} `json:"filtersPerEntityType"`*/
	} `json:"filterConfig"`
}

// Title custom chart and usql have different ways to define their tile names - so - lets figure it out by looking at the potential values
func (tile Tile) Title() string {
	if tile.FilterConfig.CustomName != "" {
		return tile.FilterConfig.CustomName
	}

	if tile.CustomName != "" {
		return tile.CustomName
	}

	return tile.Name
}

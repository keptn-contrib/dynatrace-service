package lib

import (
	"encoding/json"
	"strings"

	"github.com/keptn/go-utils/pkg/models"
)

// WEB APPLICATION
type WebApplication struct {
	Metadata                         *WAMetadata                 `json:"metadata,omitempty"`
	Identifier                       *string                     `json:"identifier,omitempty"`
	Name                             string                      `json:"name"`
	Type                             string                      `json:"type"`
	RealUserMonitoringEnabled        bool                        `json:"realUserMonitoringEnabled"`
	CostControlUserSessionPercentage float64                     `json:"costControlUserSessionPercentage"`
	LoadActionKeyPerformanceMetric   string                      `json:"loadActionKeyPerformanceMetric"`
	XhrActionKeyPerformanceMetric    string                      `json:"xhrActionKeyPerformanceMetric"`
	LoadActionApdexSettings          WALoadActionApdexSettings   `json:"loadActionApdexSettings"`
	XhrActionApdexSettings           WAXhrActionApdexSettings    `json:"xhrActionApdexSettings"`
	CustomActionApdexSettings        WACustomActionApdexSettings `json:"customActionApdexSettings"`
	WaterfallSettings                WAWaterfallSettings         `json:"waterfallSettings"`
	MonitoringSettings               WAMonitoringSettings        `json:"monitoringSettings"`
	UserActionNamingSettings         WAUserActionNamingSettings  `json:"userActionNamingSettings"`
	MetaDataCaptureSettings          []interface{}               `json:"metaDataCaptureSettings,omitempty"`
	ConversionGoals                  []interface{}               `json:"conversionGoals,omitempty"`
}
type WAMetadata struct {
	ConfigurationVersions []int  `json:"configurationVersions"`
	ClusterVersion        string `json:"clusterVersion"`
}
type WALoadActionApdexSettings struct {
	Threshold                    float64 `json:"threshold"`
	ToleratedThreshold           int     `json:"toleratedThreshold"`
	FrustratingThreshold         int     `json:"frustratingThreshold"`
	ToleratedFallbackThreshold   int     `json:"toleratedFallbackThreshold"`
	FrustratingFallbackThreshold int     `json:"frustratingFallbackThreshold"`
	ConsiderJavaScriptErrors     bool    `json:"considerJavaScriptErrors"`
}
type WAXhrActionApdexSettings struct {
	Threshold                    float64 `json:"threshold"`
	ToleratedThreshold           int     `json:"toleratedThreshold"`
	FrustratingThreshold         int     `json:"frustratingThreshold"`
	ToleratedFallbackThreshold   int     `json:"toleratedFallbackThreshold"`
	FrustratingFallbackThreshold int     `json:"frustratingFallbackThreshold"`
	ConsiderJavaScriptErrors     bool    `json:"considerJavaScriptErrors"`
}
type WACustomActionApdexSettings struct {
	Threshold                    float64 `json:"threshold"`
	ToleratedThreshold           int     `json:"toleratedThreshold"`
	FrustratingThreshold         int     `json:"frustratingThreshold"`
	ToleratedFallbackThreshold   int     `json:"toleratedFallbackThreshold"`
	FrustratingFallbackThreshold int     `json:"frustratingFallbackThreshold"`
	ConsiderJavaScriptErrors     bool    `json:"considerJavaScriptErrors"`
}
type WAWaterfallSettings struct {
	UncompressedResourcesThreshold           int `json:"uncompressedResourcesThreshold"`
	ResourcesThreshold                       int `json:"resourcesThreshold"`
	ResourceBrowserCachingThreshold          int `json:"resourceBrowserCachingThreshold"`
	SlowFirstPartyResourcesThreshold         int `json:"slowFirstPartyResourcesThreshold"`
	SlowThirdPartyResourcesThreshold         int `json:"slowThirdPartyResourcesThreshold"`
	SlowCdnResourcesThreshold                int `json:"slowCdnResourcesThreshold"`
	SpeedIndexVisuallyCompleteRatioThreshold int `json:"speedIndexVisuallyCompleteRatioThreshold"`
}
type WAJavaScriptFrameworkSupport struct {
	Angular       bool `json:"angular"`
	Dojo          bool `json:"dojo"`
	ExtJS         bool `json:"extJS"`
	Icefaces      bool `json:"icefaces"`
	JQuery        bool `json:"jQuery"`
	MooTools      bool `json:"mooTools"`
	Prototype     bool `json:"prototype"`
	ActiveXObject bool `json:"activeXObject"`
}
type WAResourceTimingSettings struct {
	W3CResourceTimings                        bool   `json:"w3cResourceTimings"`
	NonW3CResourceTimings                     bool   `json:"nonW3cResourceTimings"`
	NonW3CResourceTimingsInstrumentationDelay int    `json:"nonW3cResourceTimingsInstrumentationDelay"`
	ResourceTimingCaptureType                 string `json:"resourceTimingCaptureType"`
	ResourceTimingsDomainLimit                int    `json:"resourceTimingsDomainLimit"`
}
type WATimeoutSettings struct {
	TimedActionSupport          bool `json:"timedActionSupport"`
	TemporaryActionLimit        int  `json:"temporaryActionLimit"`
	TemporaryActionTotalTimeout int  `json:"temporaryActionTotalTimeout"`
}
type WAContentCapture struct {
	ResourceTimingSettings        WAResourceTimingSettings `json:"resourceTimingSettings"`
	JavaScriptErrors              bool                     `json:"javaScriptErrors"`
	TimeoutSettings               WATimeoutSettings        `json:"timeoutSettings"`
	VisuallyCompleteAndSpeedIndex bool                     `json:"visuallyCompleteAndSpeedIndex"`
}
type WAAdditionalEventHandlers struct {
	UserMouseupEventForClicks bool `json:"userMouseupEventForClicks"`
	ClickEventHandler         bool `json:"clickEventHandler"`
	MouseupEventHandler       bool `json:"mouseupEventHandler"`
	BlurEventHandler          bool `json:"blurEventHandler"`
	ChangeEventHandler        bool `json:"changeEventHandler"`
	ToStringMethod            bool `json:"toStringMethod"`
	MaxDomNodesToInstrument   int  `json:"maxDomNodesToInstrument"`
}
type WAEventWrapperSettings struct {
	Click      bool `json:"click"`
	MouseUp    bool `json:"mouseUp"`
	Change     bool `json:"change"`
	Blur       bool `json:"blur"`
	TouchStart bool `json:"touchStart"`
	TouchEnd   bool `json:"touchEnd"`
}
type WAGlobalEventCaptureSettings struct {
	MouseUp                            bool   `json:"mouseUp"`
	MouseDown                          bool   `json:"mouseDown"`
	Click                              bool   `json:"click"`
	DoubleClick                        bool   `json:"doubleClick"`
	KeyUp                              bool   `json:"keyUp"`
	KeyDown                            bool   `json:"keyDown"`
	Scroll                             bool   `json:"scroll"`
	AdditionalEventCapturedAsUserInput string `json:"additionalEventCapturedAsUserInput"`
}
type WAAdvancedJavaScriptTagSettings struct {
	SyncBeaconFirefox                   bool                         `json:"syncBeaconFirefox"`
	SyncBeaconInternetExplorer          bool                         `json:"syncBeaconInternetExplorer"`
	InstrumentUnsupportedAjaxFrameworks bool                         `json:"instrumentUnsupportedAjaxFrameworks"`
	SpecialCharactersToEscape           string                       `json:"specialCharactersToEscape"`
	MaxActionNameLength                 int                          `json:"maxActionNameLength"`
	MaxErrorsToCapture                  int                          `json:"maxErrorsToCapture"`
	AdditionalEventHandlers             WAAdditionalEventHandlers    `json:"additionalEventHandlers"`
	EventWrapperSettings                WAEventWrapperSettings       `json:"eventWrapperSettings"`
	GlobalEventCaptureSettings          WAGlobalEventCaptureSettings `json:"globalEventCaptureSettings"`
}
type WAMonitoringSettings struct {
	FetchRequests                    bool                            `json:"fetchRequests"`
	XMLHTTPRequest                   bool                            `json:"xmlHttpRequest"`
	JavaScriptFrameworkSupport       WAJavaScriptFrameworkSupport    `json:"javaScriptFrameworkSupport"`
	ContentCapture                   WAContentCapture                `json:"contentCapture"`
	ExcludeXhrRegex                  string                          `json:"excludeXhrRegex"`
	InjectionMode                    string                          `json:"injectionMode"`
	AddCrossOriginAnonymousAttribute bool                            `json:"addCrossOriginAnonymousAttribute"`
	ScriptTagCacheDurationInHours    int                             `json:"scriptTagCacheDurationInHours"`
	LibraryFileLocation              string                          `json:"libraryFileLocation"`
	MonitoringDataPath               string                          `json:"monitoringDataPath"`
	CustomConfigurationProperties    string                          `json:"customConfigurationProperties"`
	ServerRequestPathID              string                          `json:"serverRequestPathId"`
	SecureCookieAttribute            bool                            `json:"secureCookieAttribute"`
	CookiePlacementDomain            string                          `json:"cookiePlacementDomain"`
	CacheControlHeaderOptimizations  bool                            `json:"cacheControlHeaderOptimizations"`
	AdvancedJavaScriptTagSettings    WAAdvancedJavaScriptTagSettings `json:"advancedJavaScriptTagSettings"`
}
type WAUserActionNamingSettings struct {
	Placeholders             []interface{} `json:"placeholders,omitempty"`
	LoadActionNamingRules    []interface{} `json:"loadActionNamingRules,omitempty"`
	XhrActionNamingRules     []interface{} `json:"xhrActionNamingRules,omitempty"`
	IgnoreCase               bool          `json:"ignoreCase"`
	SplitUserActionsByDomain bool          `json:"splitUserActionsByDomain"`
}

// APP DETECTION RULES

type AppDetectionRule struct {
	Metadata              *ADRMetadata    `json:"metadata,omitempty"`
	ID                    *string         `json:"id,omitempty"`
	ApplicationIdentifier string          `json:"applicationIdentifier"`
	FilterConfig          ADRFilterConfig `json:"filterConfig"`
}
type ADRMetadata struct {
	ConfigurationVersions []int  `json:"configurationVersions"`
	ClusterVersion        string `json:"clusterVersion"`
}
type ADRFilterConfig struct {
	Pattern                string `json:"pattern"`
	ApplicationMatchType   string `json:"applicationMatchType"`
	ApplicationMatchTarget string `json:"applicationMatchTarget"`
}

func (dt *DynatraceHelper) CreateWebApplications(project string, shipyard models.Shipyard) error {
	for _, stage := range shipyard.Stages {
		applicationID, err := dt.createWebApplication(project, stage.Name)
		if err != nil {
			// try to create the other applications
			continue
		}

		_, _ = dt.createAppDetectionRule(project, stage.Name, applicationID)
	}
	return nil
}

func (dt *DynatraceHelper) createWebApplication(project string, stage string) (string, error) {
	application := createWebApplication(project, stage)

	dt.Logger.Info("Checking for existing Web Application: Keptn " + project + " " + stage)
	resp, err := dt.sendDynatraceAPIRequest("/api/config/v1/applications/web", "GET", "")
	if err != nil {
		return "", err
	}

	items := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(resp), items)
	if err != nil {
		return "", err
	}

	for _, item := range items.Values {
		if strings.ToLower(item.Name) == strings.ToLower(application.Name) {
			dt.Logger.Info("Web Application Keptn " + project + " " + stage + " already exists.")
			return item.ID, nil
		}
	}

	dt.Logger.Info("Creating new Web Application: Keptn " + project + " " + stage)
	payload, err := json.Marshal(application)

	if err != nil {
		return "", err
	}
	resp, err = dt.sendDynatraceAPIRequest("/api/config/v1/applications/web", "POST", string(payload))

	if err != nil {
		return "", err
	}

	responseItem := &Values{}

	err = json.Unmarshal([]byte(resp), responseItem)
	if err != nil {
		return "", err
	}

	dt.Logger.Info("Successfully created Web Application: Keptn " + project + " " + stage + ". ID=" + responseItem.ID)
	return responseItem.ID, nil

}

func (dt *DynatraceHelper) createAppDetectionRule(project string, stage string, webAppID string) (string, error) {
	detectionRule := createAppDetectionRule(project, stage, webAppID)
	dt.Logger.Info("Checking for existing Application Detection Rules for Keptn " + project + " " + stage)
	resp, err := dt.sendDynatraceAPIRequest("/api/config/v1/applicationDetectionRules", "GET", "")
	if err != nil {
		return "", err
	}

	items := &DTAPIListResponse{}

	err = json.Unmarshal([]byte(resp), items)
	if err != nil {
		return "", err
	}

	for _, item := range items.Values {
		if strings.ToLower(item.Name) == strings.ToLower(getWebApplicationName(project, stage)) {
			dt.Logger.Info("Web Application Detection Rule for Keptn " + project + " " + stage + " already exists.")
			return item.ID, nil
		}
	}

	dt.Logger.Info("Creating new Application Detection Rule for Keptn " + project + " " + stage)
	payload, err := json.Marshal(detectionRule)

	if err != nil {
		return "", err
	}
	resp, err = dt.sendDynatraceAPIRequest("/api/config/v1/applicationDetectionRules", "POST", string(payload))

	if err != nil {
		return "", err
	}

	responseItem := &Values{}

	err = json.Unmarshal([]byte(resp), responseItem)
	if err != nil {
		return "", err
	}

	dt.Logger.Info("Successfully created Application Detection Rule for Keptn " + project + " " + stage + ". ID=" + responseItem.ID)
	return responseItem.ID, nil
}

func createAppDetectionRule(project string, stage string, webAppID string) *AppDetectionRule {
	return &AppDetectionRule{
		ApplicationIdentifier: webAppID,
		FilterConfig: ADRFilterConfig{
			Pattern:                project + "-" + stage,
			ApplicationMatchType:   "CONTAINS",
			ApplicationMatchTarget: "DOMAIN",
		},
	}
}

func createWebApplication(project string, stage string) *WebApplication {
	return &WebApplication{
		Name:                             getWebApplicationName(project, stage),
		Type:                             "AUTO_INJECTED",
		RealUserMonitoringEnabled:        true,
		CostControlUserSessionPercentage: 100.0,
		LoadActionKeyPerformanceMetric:   "VISUALLY_COMPLETE",
		XhrActionKeyPerformanceMetric:    "VISUALLY_COMPLETE",
		LoadActionApdexSettings: WALoadActionApdexSettings{
			Threshold:                    3.0,
			ToleratedThreshold:           3000,
			FrustratingThreshold:         12000,
			ToleratedFallbackThreshold:   3000,
			FrustratingFallbackThreshold: 12000,
			ConsiderJavaScriptErrors:     true,
		},
		XhrActionApdexSettings: WAXhrActionApdexSettings{
			Threshold:                    3.0,
			ToleratedThreshold:           3000,
			FrustratingThreshold:         12000,
			ToleratedFallbackThreshold:   3000,
			FrustratingFallbackThreshold: 12000,
			ConsiderJavaScriptErrors:     true,
		},
		CustomActionApdexSettings: WACustomActionApdexSettings{
			Threshold:                    3.0,
			ToleratedThreshold:           3000,
			FrustratingThreshold:         12000,
			ToleratedFallbackThreshold:   3000,
			FrustratingFallbackThreshold: 12000,
			ConsiderJavaScriptErrors:     true,
		},
		WaterfallSettings: WAWaterfallSettings{
			UncompressedResourcesThreshold:           860,
			ResourcesThreshold:                       100000,
			ResourceBrowserCachingThreshold:          50,
			SlowFirstPartyResourcesThreshold:         200000,
			SlowThirdPartyResourcesThreshold:         200000,
			SlowCdnResourcesThreshold:                200000,
			SpeedIndexVisuallyCompleteRatioThreshold: 50,
		},
		MonitoringSettings: WAMonitoringSettings{
			FetchRequests:  false,
			XMLHTTPRequest: false,
			JavaScriptFrameworkSupport: WAJavaScriptFrameworkSupport{
				Angular:       false,
				Dojo:          false,
				ExtJS:         false,
				Icefaces:      false,
				JQuery:        false,
				MooTools:      false,
				Prototype:     false,
				ActiveXObject: false,
			},
			ContentCapture: WAContentCapture{
				ResourceTimingSettings: WAResourceTimingSettings{
					W3CResourceTimings:                        true,
					NonW3CResourceTimings:                     false,
					NonW3CResourceTimingsInstrumentationDelay: 50,
					ResourceTimingCaptureType:                 "CAPTURE_FULL_DETAILS",
					ResourceTimingsDomainLimit:                10,
				},
				JavaScriptErrors: true,
				TimeoutSettings: WATimeoutSettings{
					TimedActionSupport:          false,
					TemporaryActionLimit:        0,
					TemporaryActionTotalTimeout: 100,
				},
				VisuallyCompleteAndSpeedIndex: true,
			},
			ExcludeXhrRegex:                  "",
			InjectionMode:                    "JAVASCRIPT_TAG",
			AddCrossOriginAnonymousAttribute: true,
			ScriptTagCacheDurationInHours:    1,
			LibraryFileLocation:              "",
			MonitoringDataPath:               "",
			CustomConfigurationProperties:    "",
			ServerRequestPathID:              "",
			SecureCookieAttribute:            false,
			CookiePlacementDomain:            "",
			CacheControlHeaderOptimizations:  true,
			AdvancedJavaScriptTagSettings: WAAdvancedJavaScriptTagSettings{
				SyncBeaconFirefox:                   false,
				SyncBeaconInternetExplorer:          false,
				InstrumentUnsupportedAjaxFrameworks: false,
				SpecialCharactersToEscape:           "",
				MaxActionNameLength:                 100,
				MaxErrorsToCapture:                  10,
				AdditionalEventHandlers: WAAdditionalEventHandlers{
					UserMouseupEventForClicks: false,
					ClickEventHandler:         false,
					MouseupEventHandler:       false,
					BlurEventHandler:          false,
					ChangeEventHandler:        false,
					ToStringMethod:            false,
					MaxDomNodesToInstrument:   5000,
				},
				EventWrapperSettings: WAEventWrapperSettings{
					Click:      false,
					MouseUp:    false,
					Change:     false,
					Blur:       false,
					TouchStart: false,
					TouchEnd:   false,
				},
				GlobalEventCaptureSettings: WAGlobalEventCaptureSettings{
					MouseUp:                            true,
					MouseDown:                          true,
					Click:                              true,
					DoubleClick:                        true,
					KeyUp:                              true,
					KeyDown:                            true,
					Scroll:                             true,
					AdditionalEventCapturedAsUserInput: "",
				},
			},
		},
		UserActionNamingSettings: WAUserActionNamingSettings{
			IgnoreCase:               false,
			SplitUserActionsByDomain: false,
		},
	}
}

func getWebApplicationName(project string, stage string) string {
	return "Keptn " + project + " " + stage
}

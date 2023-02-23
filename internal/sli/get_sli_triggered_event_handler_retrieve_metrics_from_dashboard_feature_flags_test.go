package sli

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/sli/ff"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
	keptnapi "github.com/keptn/go-utils/pkg/lib"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestDashboardUseCaseProducesCorrectResultsWithSkipLowercaseSLINamesFeatureFlagEnabled(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/ff/combined"
	const dashboardID = "f40fb23d-66e2-414b-9ec1-cd761820da43"

	request := newMetricsV2QueryRequestBuilder("(builtin:service.response.time:filter(and(or(in(\"dt.entity.service\",entitySelector(\"type(service),entityId(~\"SERVICE-C6876D601CA5DDFD~\")\")),in(\"dt.entity.service\",entitySelector(\"type(service),entityId(~\"SERVICE-A9AD48F41E6A8034~\")\"))))):splitBy():percentile(90.0):auto:sort(value(percentile(90.0),descending)):limit(20)):limit(100):names").copyWithResolution(resolutionInf).build()

	handler := test.NewCombinedURLHandler(t)
	handler.AddExactFile(dynatrace.DashboardsPath+"/"+dashboardID, filepath.Join(testDataFolder, "dashboard.json"))
	handler.AddExactFile(request, filepath.Join(testDataFolder, "metrics_get_by_query.json"))

	testConfigs := []struct {
		name              string
		expectedObjective *keptnapi.SLO
		featureFlags      ff.GetSLIFeatureFlags
	}{
		{
			name: "feature flags disabled",
			expectedObjective: createTestSLO(
				"srt_p90_all",
				"Service Response Time P90 2 services"),
			featureFlags: ff.GetSLIFeatureFlags{},
		},
		{
			name:              "skip lowercase SLI names only",
			expectedObjective: createTestSLO("srt_P90_all", "Service Response Time P90 2 services"),
			featureFlags:      ff.NewGetSLIFeatureFlags(true, false, false),
		},
		{
			name:              "skip display names in SLOs",
			expectedObjective: createTestSLO("srt_p90_all", ""), // no display name
			featureFlags:      ff.NewGetSLIFeatureFlags(false, true, false),
		},
		{
			name:              "skip lowercase SLI names and skip display names in SLOs",
			expectedObjective: createTestSLO("srt_P90_all", ""),
			featureFlags:      ff.NewGetSLIFeatureFlags(true, true, false),
		},
	}

	for _, cfg := range testConfigs {
		t.Run(cfg.name, func(t *testing.T) {
			sliResultsAssertionsFuncs := []func(t *testing.T, actual sliResult){
				createSuccessfulSLIResultAssertionsFunc(cfg.expectedObjective.SLI, 197588.84577351453, request),
			}

			uploadedSLOsAssertionsFunc := func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
				if !assert.NotNil(t, actual) {
					return
				}
				assert.EqualValues(t, createSLOScore(), actual.TotalScore)
				assert.EqualValues(t, createSLOComparison(), actual.Comparison)

				if !assert.Equal(t, 1, len(actual.Objectives)) {
					assert.EqualValues(t, cfg.expectedObjective, actual.Objectives[0])
				}
			}

			runGetSLIsFromDashboardTestWithDashboardParameterAndFeatureFlagsAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, dashboardID, cfg.featureFlags, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
		})
	}
}

func TestDashboardUseCaseProducesCorrectResultsWithSkipDuplicationCheckFeatureFlagEnabled(t *testing.T) {
	const testDataFolder = "./testdata/dashboards/ff/skip_duplication_checks"

	const dashboardID = "f40fb23d-66e2-414b-9ec1-cd761820da43"

	handler := test.NewCombinedURLHandler(t)
	handler.AddExactFile(dynatrace.DashboardsPath+"/"+dashboardID, filepath.Join(testDataFolder, "dashboard.json"))

	expectedRequest := addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(
		handler,
		testDataFolder,
		newMetricsV2QueryRequestBuilder("(builtin:service.response.time:filter(and(or(in(\"dt.entity.service\",entitySelector(\"type(service),entityName.equals(~\"EasytravelService~\")\"))))):splitBy(\"dt.entity.service\"):avg:auto:sort(value(avg,descending)):limit(20)):limit(100):names"))

	testConfigs := []struct {
		name               string
		expectedObjective  *keptnapi.SLO
		featureFlags       ff.GetSLIFeatureFlags
		shouldBeSuccessful bool
	}{
		{
			name:               "skip duplication check only",
			expectedObjective:  createTestSLO("srt_avg_et_easytravelservice", "Service Response Time AVG ET services split (EasytravelService)"),
			featureFlags:       ff.NewGetSLIFeatureFlags(false, false, true),
			shouldBeSuccessful: true,
		},
		{
			name:               "skip duplication check and lowercase SLI names",
			expectedObjective:  createTestSLO("srt_avg_ET_EasytravelService", "Service Response Time AVG ET services split (EasytravelService)"),
			featureFlags:       ff.NewGetSLIFeatureFlags(true, false, true),
			shouldBeSuccessful: true,
		},
		{
			name:               "skip duplication check and display names in SLOs",
			expectedObjective:  createTestSLO("srt_avg_et_easytravelservice", ""),
			featureFlags:       ff.NewGetSLIFeatureFlags(false, true, true),
			shouldBeSuccessful: true,
		},
		{
			name:               "skip duplication check, lowercase SLI names and display names in SLOs",
			expectedObjective:  createTestSLO("srt_avg_ET_EasytravelService", ""),
			featureFlags:       ff.NewGetSLIFeatureFlags(true, true, true),
			shouldBeSuccessful: true,
		},
		// error cases below (duplication produces errors)
		{
			name:               "nothing skipped",
			expectedObjective:  createTestSLO("srt_avg_et_easytravelservice", "Service Response Time AVG ET services split (EasytravelService)"),
			featureFlags:       ff.GetSLIFeatureFlags{},
			shouldBeSuccessful: false,
		},
		{
			name:               "skip lowercase SLI names",
			expectedObjective:  createTestSLO("srt_avg_ET_EasytravelService", "Service Response Time AVG ET services split (EasytravelService)"),
			featureFlags:       ff.NewGetSLIFeatureFlags(true, false, false),
			shouldBeSuccessful: false,
		},
		{
			name:               "skip display names in SLOs",
			expectedObjective:  createTestSLO("srt_avg_et_easytravelservice", ""),
			featureFlags:       ff.NewGetSLIFeatureFlags(false, true, false),
			shouldBeSuccessful: false,
		},
		{
			name:               "skip lowercase SLI names and display names in SLOs",
			expectedObjective:  createTestSLO("srt_avg_ET_EasytravelService", ""),
			featureFlags:       ff.NewGetSLIFeatureFlags(true, true, false),
			shouldBeSuccessful: false,
		},
	}

	for _, cfg := range testConfigs {
		t.Run(cfg.name, func(t *testing.T) {
			var sliResultsAssertionsFuncs []func(t *testing.T, actual sliResult)
			var uploadedSLOsAssertionsFunc func(t *testing.T, actual *keptnapi.ServiceLevelObjectives)

			if cfg.shouldBeSuccessful {
				sliResultsAssertionsFuncs = []func(t *testing.T, actual sliResult){
					createSuccessfulSLIResultAssertionsFunc(cfg.expectedObjective.SLI, 115445.40697872869, expectedRequest),
					createSuccessfulSLIResultAssertionsFunc(cfg.expectedObjective.SLI, 110568.30653321775, expectedRequest),
				}

				uploadedSLOsAssertionsFunc = func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
					if !assert.NotNil(t, actual) {
						return
					}

					assert.EqualValues(t, createSLOScore(), actual.TotalScore)
					assert.EqualValues(t, createSLOComparison(), actual.Comparison)

					if assert.Equal(t, 2, len(actual.Objectives)) {
						assert.EqualValues(t, cfg.expectedObjective, actual.Objectives[0])
						assert.EqualValues(t, cfg.expectedObjective, actual.Objectives[1])

					}
				}
				runGetSLIsFromDashboardTestWithDashboardParameterAndFeatureFlagsAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, dashboardID, cfg.featureFlags, getSLIFinishedEventSuccessAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
			} else {
				const errorMessage = "duplicate SLI and display name"

				sliResultsAssertionsFuncs = []func(t *testing.T, actual sliResult){
					createFailedSLIResultWithQueryAssertionsFunc(cfg.expectedObjective.SLI, expectedRequest, errorMessage),
					createFailedSLIResultWithQueryAssertionsFunc(cfg.expectedObjective.SLI, expectedRequest, errorMessage),
				}

				uploadedSLOsAssertionsFunc = func(t *testing.T, actual *keptnapi.ServiceLevelObjectives) {
					if !assert.NotNil(t, actual) {
						return
					}

					assert.EqualValues(t, createSLOScore(), actual.TotalScore)
					assert.EqualValues(t, createSLOComparison(), actual.Comparison)

					if assert.Equal(t, 2, len(actual.Objectives)) {
						assert.EqualValues(t, cfg.expectedObjective, actual.Objectives[0])
						assert.EqualValues(t, cfg.expectedObjective, actual.Objectives[1])

					}
				}

				runGetSLIsFromDashboardTestWithDashboardParameterAndFeatureFlagsAndCheckSLIsAndSLOs(t, handler, testGetSLIEventData, dashboardID, cfg.featureFlags, getSLIFinishedEventFailureAssertionsFunc, uploadedSLOsAssertionsFunc, sliResultsAssertionsFuncs...)
			}
		})
	}
}

func createTestSLO(sliName string, displayName string) *keptnapi.SLO {
	return &keptnapi.SLO{
		SLI:         sliName,
		DisplayName: displayName,
		Pass:        createUpperBoundSLOCriteria(600, true),
		Warning:     createUpperBoundSLOCriteria(800, false),
		Weight:      1,
		KeySLI:      false,
	}
}

func createUpperBoundSLOCriteria(upperBound int, inclusive bool) []*keptnapi.SLOCriteria {
	if inclusive {
		return []*keptnapi.SLOCriteria{{Criteria: []string{fmt.Sprintf("<=%d", upperBound)}}}
	}
	return []*keptnapi.SLOCriteria{{Criteria: []string{fmt.Sprintf("<%d", upperBound)}}}
}

func createSLOScore() *keptnapi.SLOScore {
	return &keptnapi.SLOScore{
		Pass:    "90%",
		Warning: "75%",
	}
}

func createSLOComparison() *keptnapi.SLOComparison {
	return &keptnapi.SLOComparison{
		CompareWith:               "single_result",
		IncludeResultWithScore:    "pass",
		NumberOfComparisonResults: 1,
		AggregateFunction:         "avg",
	}
}

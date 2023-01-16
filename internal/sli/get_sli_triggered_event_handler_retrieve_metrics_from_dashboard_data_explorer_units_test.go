package sli

import (
	"path/filepath"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func TestRetrieveMetricsFromDashboardDataExplorerTile_UnitTransform(t *testing.T) {
	const (
		testDataFolder             = "./testdata/dashboards/data_explorer/unit_transform/"
		sliName                    = "sli"
		metricUnitsConvertFileName = "metrics_units_convert1.json"
	)

	const (
		autoUnit     = "auto"
		noneUnit     = "none"
		dayUnit      = "Day"
		thousandUnit = "Kilo"
		millionUnit  = "Million"
		billionUnit  = "Billion"
		trillionUnit = "Trillion"
		specialUnit  = "Special"
		countUnit    = "Count"
	)

	const (
		microSecondsPerMilliSecond = 1000
		microSecondsPerDay         = 86400 * 1000000

		countPerThousand = 1000
		countPerMillion  = 1000 * 1000
		countPerBillion  = 1000 * 1000 * 1000
		countPerTrillion = 1000 * 1000 * 1000 * 1000
	)

	const (
		noUnitFoundSubstring   = "No unit found"
		cannotConvertSubstring = "Cannot convert"
		unknownUnitSubstring   = "unknown unit"
	)

	serviceResponseTimeMetricSelector := "(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	serviceResponseTimeRequestBuilder := newMetricsV2QueryRequestBuilder(serviceResponseTimeMetricSelector)
	serviceResponseTimeWithNoConversionHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder)
	}

	const unconvertedServiceResponseTimeValue = 54896.485186544574

	serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc := createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).build())
	serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc := createFailedSLIResultWithQueryAssertionsFunc(sliName, serviceResponseTimeRequestBuilder.build(), noUnitFoundSubstring)

	const unconvertedNonDbChildCallCountValue = 341746808.0

	nonDbChildCallCountMetricSelector := "(builtin:service.nonDbChildCallCount:splitBy():sort(value(auto,descending)):limit(20)):limit(100):names"
	nonDbChildCallCountRequestBuilder := newMetricsV2QueryRequestBuilder(nonDbChildCallCountMetricSelector)
	nonDbChildCallCountHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder)
	}

	nonDbChildCallCountUnknownUnitHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForInitialMetricsDefinition(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder)
	}

	nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc := createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).build())
	nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc := createFailedSLIResultWithQueryAssertionsFunc(sliName, nonDbChildCallCountRequestBuilder.build(), unknownUnitSubstring)

	const unconvertedUnspecifiedUnitValue = 1.0

	unspecifiedUnitMetricSelector := "builtin:service.response.time:splitBy() / builtin:service.response.time:splitBy()"
	unspecifiedUnitRequestBuilder := newMetricsV2QueryRequestBuilder(unspecifiedUnitMetricSelector)
	unspecifiedUnitHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
	}

	unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc := createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).build())
	unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc := createFailedSLIResultWithQueryAssertionsFunc(sliName, unspecifiedUnitRequestBuilder.build(), unknownUnitSubstring)

	const (
		divideByThousandConversionSnippet = "/1000"
		divideByMillionConversionSnippet  = "/1000000"
		divideByBillionConversionSnippet  = "/1000000000"
		divideByTrillionConversionSnippet = "/1000000000000"
	)

	tests := []struct {
		name                              string
		unit                              string
		metricSelector                    string
		handlerAdditionalSetupFunc        func(handler *test.CombinedURLHandler, testVariantDataFolder string)
		getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData)
		sliResultAssertionsFunc           func(t *testing.T, actual sliResult)
	}{
		// service response time (sourceUnitID = MicroSecond)

		// simple success cases that require no conversion

		{
			name:                              "success_srt_empty",
			unit:                              "",
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_srt_auto",
			unit:                              autoUnit,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_srt_none",
			unit:                              noneUnit,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_srt_microsecond",
			unit:                              microSecondUnitID,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},

		// success cases that require conversion

		{
			name:           "success_srt_millisecond",
			unit:           milliSecondUnitID,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder,
					createToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue/microSecondsPerMilliSecond, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(createToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID)).build()),
		},
		{
			name:           "success_srt_day",
			unit:           dayUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, dayUnit))

			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, 6.353759859553769e-7, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(createToUnitConversionSnippet(microSecondUnitID, dayUnit)).build()),
		},

		// error cases

		{
			name:           "error_srt_byte",
			unit:           byteUnitID,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, byteUnitID))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(createToUnitConversionSnippet(microSecondUnitID, byteUnitID)).build()),
		},
		{
			name:           "error_srt_thousand",
			unit:           thousandUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForFailedMetricsQueryWithUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, thousandUnit))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc,
		},
		{
			name:           "error_srt_special",
			unit:           specialUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForFailedMetricsQueryWithUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, specialUnit))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc,
		},

		// builtin:service.nonDbChildCallCount (sourceUnitID = Count)

		// success cases where no conversion is required

		{
			name:                              "success_ndbccc_empty",
			unit:                              "",
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_ndbccc_auto",
			unit:                              autoUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_ndbccc_none",
			unit:                              noneUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc,
		},

		// success cases where conversion is done by dynatrace-service

		{
			name:           "success_ndbccc_thousand",
			unit:           thousandUnit,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByThousandConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerThousand, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByThousandConversionSnippet).build()),
		},
		{
			name:           "success_ndbccc_million",
			unit:           millionUnit,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByMillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerMillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByMillionConversionSnippet).build()),
		},
		{
			name:           "success_ndbccc_billion",
			unit:           billionUnit,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByBillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerBillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByBillionConversionSnippet).build()),
		},
		{
			name:           "success_ndbccc_trillion",
			unit:           trillionUnit,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByTrillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerTrillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByTrillionConversionSnippet).build()),
		},

		// error cases where no conversion is possible

		{
			name:                              "error_ndbccc_byte",
			unit:                              byteUnitID,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountUnknownUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc, //createFailedSLIResultWithQueryAssertionsFunc(sliName, nonDbChildCallCountRequestBuilder.build()),
		},
		{
			name:                              "error_ndbccc_millisecond",
			unit:                              milliSecondUnitID,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountUnknownUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:                              "error_ndbccc_special",
			unit:                              specialUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountUnknownUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc,
		},

		// builtin:service.response.time:splitBy() / builtin:service.response.time:splitBy() (sourceUnitID = Unspecified)

		// success cases where no conversion is required

		{
			name:                              "success_uum_empty",
			unit:                              "",
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_uum_auto",
			unit:                              autoUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_uum_none",
			unit:                              noneUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc,
		},

		// success cases where conversion is done by dynatrace-service

		{
			name:           "success_uum_thousand",
			unit:           thousandUnit,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByThousandConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerThousand, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByThousandConversionSnippet).build()),
		},
		{
			name:           "success_uum_million",
			unit:           millionUnit,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByMillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerMillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByMillionConversionSnippet).build()),
		},
		{
			name:           "success_uum_billion",
			unit:           billionUnit,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByBillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerBillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByBillionConversionSnippet).build()),
		},
		{
			name:           "success_uum_trillion",
			unit:           trillionUnit,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByTrillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerTrillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByTrillionConversionSnippet).build()),
		},

		// error cases where no conversion is possible

		{
			name:           "error_uum_byte",
			unit:           byteUnitID,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForInitialMetricsDefinition(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:           "error_uum_millisecond",
			unit:           milliSecondUnitID,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForInitialMetricsDefinition(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:           "error_uum_special",
			unit:           specialUnit,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForInitialMetricsDefinition(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				filepath.Join(testDataFolder, "dashboard.template.json"),
				struct {
					Unit           string
					MetricSelector string
				}{
					Unit:           tt.unit,
					MetricSelector: tt.metricSelector,
				})

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)
			tt.handlerAdditionalSetupFunc(handler, testVariantDataFolder)
			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, tt.getSLIFinishedEventAssertionsFunc, tt.sliResultAssertionsFunc)

		})
	}
}

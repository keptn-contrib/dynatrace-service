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
		autoUnit        = "auto"
		noneUnit        = "none"
		milliSecondUnit = "MilliSecond"
		microSecondUnit = "MicroSecond"
		dayUnit         = "Day"
		thousandUnit    = "Kilo"
		millionUnit     = "Million"
		billionUnit     = "Billion"
		trillionUnit    = "Trillion"
		byteUnit        = "Byte"
		specialUnit     = "Special"
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

	const unconvertedServiceResponseTimeValue = 54896.485383423984

	serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc := createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).build())
	serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc := createFailedSLIResultWithQueryAssertionsFunc(sliName, serviceResponseTimeRequestBuilder.build(), noUnitFoundSubstring)
	serviceResponseTimeFailedCannotConvertSLIResultAssertionsFunc := createFailedSLIResultWithQueryAssertionsFunc(sliName, serviceResponseTimeRequestBuilder.build(), cannotConvertSubstring)

	const unconvertedNonDbChildCallCountValue = 341746808.0

	nonDbChildCallCountMetricSelector := "(builtin:service.nonDbChildCallCount:splitBy():sort(value(auto,descending)):limit(20)):limit(100):names"
	nonDbChildCallCountRequestBuilder := newMetricsV2QueryRequestBuilder(nonDbChildCallCountMetricSelector)
	nonDbChildCallCountHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder)
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
			unit:                              microSecondUnit,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},

		// success cases that require conversion

		{
			name:           "success_srt_millisecond",
			unit:           milliSecondUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder)
				handler.AddExactFile(buildMetricsUnitsConvertRequest(microSecondUnit, unconvertedServiceResponseTimeValue, milliSecondUnit), filepath.Join(testVariantDataFolder, metricUnitsConvertFileName))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue/microSecondsPerMilliSecond, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).build()),
		},
		{
			name:           "success_srt_day",
			unit:           dayUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder)
				handler.AddExactFile(buildMetricsUnitsConvertRequest(microSecondUnit, unconvertedServiceResponseTimeValue, dayUnit), filepath.Join(testVariantDataFolder, metricUnitsConvertFileName))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue/microSecondsPerDay, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).build()),
		},

		// error cases
		{
			name:           "error_srt_byte",
			unit:           byteUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder)
				handler.AddExactError(buildMetricsUnitsConvertRequest(microSecondUnit, unconvertedServiceResponseTimeValue, byteUnit), 400, filepath.Join(testVariantDataFolder, metricUnitsConvertFileName))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeFailedCannotConvertSLIResultAssertionsFunc,
		},
		{
			name:           "error_srt_thousand",
			unit:           thousandUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder)
				handler.AddExactError(buildMetricsUnitsConvertRequest(microSecondUnit, unconvertedServiceResponseTimeValue, thousandUnit), 400, filepath.Join(testVariantDataFolder, metricUnitsConvertFileName))

			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc,
		},
		{
			name:           "error_srt_special",
			unit:           specialUnit,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder)
				handler.AddExactError(buildMetricsUnitsConvertRequest(microSecondUnit, unconvertedServiceResponseTimeValue, specialUnit), 400, filepath.Join(testVariantDataFolder, metricUnitsConvertFileName))

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
			name:                              "success_ndbccc_thousand",
			unit:                              thousandUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerThousand, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).build()),
		},
		{
			name:                              "success_ndbccc_million",
			unit:                              millionUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerMillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).build()),
		},
		{
			name:                              "success_ndbccc_billion",
			unit:                              billionUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerBillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).build()),
		},
		{
			name:                              "success_ndbccc_trillion",
			unit:                              trillionUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerTrillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).build()),
		},

		// error cases where no conversion is possible

		{
			name:                              "error_ndbccc_byte",
			unit:                              byteUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:                              "error_ndbccc_millisecond",
			unit:                              milliSecondUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:                              "error_ndbccc_special",
			unit:                              specialUnit,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
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
			name:                              "success_uum_thousand",
			unit:                              thousandUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerThousand, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).build()),
		},
		{
			name:                              "success_uum_million",
			unit:                              millionUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerMillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).build()),
		},
		{
			name:                              "success_uum_billion",
			unit:                              billionUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerBillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).build()),
		},
		{
			name:                              "success_uum_trillion",
			unit:                              trillionUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFunc:           createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerTrillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).build()),
		},

		// error cases where no conversion is possible

		{
			name:                              "error_uum_byte",
			unit:                              byteUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:                              "error_uum_millisecond",
			unit:                              milliSecondUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFunc:           unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:                              "error_uum_special",
			unit:                              specialUnit,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
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

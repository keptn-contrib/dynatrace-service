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

	const (
		resolutionNull      = "null"
		resolution10Minutes = "10m"
	)

	serviceResponseTimeMetricSelector := "(builtin:service.response.time:splitBy():avg:auto:sort(value(avg,descending)):limit(10)):limit(100):names"
	serviceResponseTimeRequestBuilder := newMetricsV2QueryRequestBuilder(serviceResponseTimeMetricSelector)
	serviceResponseTimeWithNoConversionHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder)
	}

	const unconvertedServiceResponseTimeValue = 54896.485186544574

	serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc := toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).build()))
	serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc := toSlice(createFailedSLIResultWithQueryAssertionsFunc(sliName, serviceResponseTimeRequestBuilder.build(), noUnitFoundSubstring))

	const unconvertedNonDbChildCallCountValue = 341746808.0

	nonDbChildCallCountMetricSelector := "(builtin:service.nonDbChildCallCount:splitBy():sort(value(auto,descending)):limit(20)):limit(100):names"
	nonDbChildCallCountRequestBuilder := newMetricsV2QueryRequestBuilder(nonDbChildCallCountMetricSelector)
	nonDbChildCallCountHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder)
	}

	nonDbChildCallCountUnknownUnitHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder)
	}

	nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc := toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).build()))
	nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc := toSlice(createFailedSLIResultWithQueryAssertionsFunc(sliName, nonDbChildCallCountRequestBuilder.build(), unknownUnitSubstring))

	const unconvertedUnspecifiedUnitValue = 1.0

	unspecifiedUnitMetricSelector := "builtin:service.response.time:splitBy() / builtin:service.response.time:splitBy()"
	unspecifiedUnitRequestBuilder := newMetricsV2QueryRequestBuilder(unspecifiedUnitMetricSelector)
	unspecifiedUnitHandlerSetupFunc := func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
		addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
	}

	unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc := toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).build()))
	unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc := toSlice(createFailedSLIResultWithQueryAssertionsFunc(sliName, unspecifiedUnitRequestBuilder.build(), unknownUnitSubstring))

	const serviceResponseTimeNoAggregationMetricSelector = "(builtin:service.response.time:splitBy():sort(value(auto,descending)):limit(20)):limit(100):names"
	serviceResponseTimeNoAggregationResolution10MinutesRequestBuilder := newMetricsV2QueryRequestBuilder(serviceResponseTimeNoAggregationMetricSelector).copyWithResolution(resolution10Minutes)

	const openSecurityProblemsMetricSelector = "(builtin:security.securityProblem.open.global:splitBy():sort(value(auto,descending)):limit(20)):limit(100):names"
	openSecurityProblemsRequestBuilder := newMetricsV2QueryRequestBuilder(openSecurityProblemsMetricSelector)

	const openSecurityProblemsByTypeMetricSelector = "(builtin:security.securityProblem.open.global:splitBy(Type):sort(value(auto,descending)):limit(20)):limit(100):names"
	openSecurityProblemsByTypeRequestBuilder := newMetricsV2QueryRequestBuilder(openSecurityProblemsByTypeMetricSelector)

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
		resolution                        string
		handlerAdditionalSetupFunc        func(handler *test.CombinedURLHandler, testVariantDataFolder string)
		getSLIFinishedEventAssertionsFunc func(t *testing.T, actual *getSLIFinishedEventData)
		sliResultAssertionsFuncs          []func(t *testing.T, actual sliResult)
	}{
		// service response time (sourceUnitID = MicroSecond)

		// simple success cases that require no conversion

		{
			name:                              "success_srt_empty",
			unit:                              "",
			resolution:                        resolutionNull,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_srt_auto",
			unit:                              autoUnit,
			resolution:                        resolutionNull,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_srt_none",
			unit:                              noneUnit,
			resolution:                        resolutionNull,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_srt_microsecond",
			unit:                              microSecondUnitID,
			resolution:                        resolutionNull,
			metricSelector:                    serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc:        serviceResponseTimeWithNoConversionHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          serviceResponseTimeNoConversionRequiredSLIResultAssertionsFunc,
		},

		// success cases that require conversion

		{
			name:           "success_srt_millisecond",
			unit:           milliSecondUnitID,
			resolution:     resolutionNull,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder,
					createToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue/microSecondsPerMilliSecond, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(createToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID)).build())),
		},
		{
			name:           "success_srt_day",
			unit:           dayUnit,
			resolution:     resolutionNull,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, dayUnit))

			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, 6.353759859553769e-7, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(createToUnitConversionSnippet(microSecondUnitID, dayUnit)).build())),
		},

		// error cases

		{
			name:           "error_srt_byte",
			unit:           byteUnitID,
			resolution:     resolutionNull,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, byteUnitID))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedServiceResponseTimeValue, serviceResponseTimeRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(createToUnitConversionSnippet(microSecondUnitID, byteUnitID)).build())),
		},
		{
			name:           "error_srt_thousand",
			unit:           thousandUnit,
			resolution:     resolutionNull,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForFailedMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, thousandUnit))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc,
		},
		{
			name:           "error_srt_special",
			unit:           specialUnit,
			resolution:     resolutionNull,
			metricSelector: serviceResponseTimeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForFailedMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeRequestBuilder, createToUnitConversionSnippet(microSecondUnitID, specialUnit))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          serviceResponseTimeFailedNoUnitSLIResultAssertionsFunc,
		},

		// builtin:service.nonDbChildCallCount (sourceUnitID = Count)

		// success cases where no conversion is required

		{
			name:                              "success_ndbccc_empty",
			unit:                              "",
			resolution:                        resolutionNull,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_ndbccc_auto",
			unit:                              autoUnit,
			resolution:                        resolutionNull,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_ndbccc_none",
			unit:                              noneUnit,
			resolution:                        resolutionNull,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          nonDbChildCallCountNoConversionRequiredSLIResultAssertionsFunc,
		},

		// success cases where conversion is done by dynatrace-service

		{
			name:           "success_ndbccc_thousand",
			unit:           thousandUnit,
			resolution:     resolutionNull,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByThousandConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerThousand, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByThousandConversionSnippet).build())),
		},
		{
			name:           "success_ndbccc_million",
			unit:           millionUnit,
			resolution:     resolutionNull,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByMillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerMillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByMillionConversionSnippet).build())),
		},
		{
			name:           "success_ndbccc_billion",
			unit:           billionUnit,
			resolution:     resolutionNull,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByBillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerBillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByBillionConversionSnippet).build())),
		},
		{
			name:           "success_ndbccc_trillion",
			unit:           trillionUnit,
			resolution:     resolutionNull,
			metricSelector: nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, nonDbChildCallCountRequestBuilder, divideByTrillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedNonDbChildCallCountValue/countPerTrillion, nonDbChildCallCountRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByTrillionConversionSnippet).build())),
		},

		// error cases where no conversion is possible

		{
			name:                              "error_ndbccc_byte",
			unit:                              byteUnitID,
			resolution:                        resolutionNull,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountUnknownUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc, //createFailedSLIResultWithQueryAssertionsFunc(sliName, nonDbChildCallCountRequestBuilder.build()),
		},
		{
			name:                              "error_ndbccc_millisecond",
			unit:                              milliSecondUnitID,
			resolution:                        resolutionNull,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountUnknownUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:                              "error_ndbccc_special",
			unit:                              specialUnit,
			resolution:                        resolutionNull,
			metricSelector:                    nonDbChildCallCountMetricSelector,
			handlerAdditionalSetupFunc:        nonDbChildCallCountUnknownUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          nonDbChildCallCountFailedUnknownUnitSLIResultAssertionsFunc,
		},

		// builtin:service.response.time:splitBy() / builtin:service.response.time:splitBy() (sourceUnitID = Unspecified)

		// success cases where no conversion is required

		{
			name:                              "success_uum_empty",
			unit:                              "",
			resolution:                        resolutionNull,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_uum_auto",
			unit:                              autoUnit,
			resolution:                        resolutionNull,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc,
		},
		{
			name:                              "success_uum_none",
			unit:                              noneUnit,
			resolution:                        resolutionNull,
			metricSelector:                    unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc:        unspecifiedUnitHandlerSetupFunc,
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          unspecifiedUnitNoConversionRequiredSLIResultAssertionsFunc,
		},

		// success cases where conversion is done by dynatrace-service

		{
			name:           "success_uum_thousand",
			unit:           thousandUnit,
			resolution:     resolutionNull,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByThousandConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerThousand, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByThousandConversionSnippet).build())),
		},
		{
			name:           "success_uum_million",
			unit:           millionUnit,
			resolution:     resolutionNull,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByMillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerMillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByMillionConversionSnippet).build())),
		},
		{
			name:           "success_uum_billion",
			unit:           billionUnit,
			resolution:     resolutionNull,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByBillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerBillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByBillionConversionSnippet).build())),
		},
		{
			name:           "success_uum_trillion",
			unit:           trillionUnit,
			resolution:     resolutionNull,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder, divideByTrillionConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, unconvertedUnspecifiedUnitValue/countPerTrillion, unspecifiedUnitRequestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(divideByTrillionConversionSnippet).build())),
		},

		// error cases where no conversion is possible

		{
			name:           "error_uum_byte",
			unit:           byteUnitID,
			resolution:     resolutionNull,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:           "error_uum_millisecond",
			unit:           milliSecondUnitID,
			resolution:     resolutionNull,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},
		{
			name:           "error_uum_special",
			unit:           specialUnit,
			resolution:     resolutionNull,
			metricSelector: unspecifiedUnitMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testVariantDataFolder, unspecifiedUnitRequestBuilder)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventFailureAssertionsFunc,
			sliResultAssertionsFuncs:          unspecifiedUnitFailedUnknownUnitSLIResultAssertionsFunc,
		},

		// additional test cases

		// fold with toUnit and no aggregation specified
		{
			name:           "success_srt_resolution_10m_millisecond",
			unit:           milliSecondUnitID,
			resolution:     resolution10Minutes,
			metricSelector: serviceResponseTimeNoAggregationMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithFoldAndUnitsConversionSnippet(handler, testVariantDataFolder, serviceResponseTimeNoAggregationResolution10MinutesRequestBuilder,
					createAutoToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID))
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, 54.89648587772039, serviceResponseTimeNoAggregationResolution10MinutesRequestBuilder.copyWithFold().copyWithMetricSelectorConversionSnippet(createAutoToUnitConversionSnippet(microSecondUnitID, milliSecondUnitID)).build())),
		},

		// fold with scaling producing a single value
		{
			name:           "success_osp_thousand_single_value",
			unit:           thousandUnit,
			resolution:     resolutionNull,
			metricSelector: openSecurityProblemsMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithFoldAndUnitsConversionSnippet(handler, testVariantDataFolder, openSecurityProblemsRequestBuilder, divideByThousandConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs:          toSlice(createSuccessfulSLIResultAssertionsFunc(sliName, 0.07657418459403192, openSecurityProblemsRequestBuilder.copyWithFold().copyWithMetricSelectorConversionSnippet(divideByThousandConversionSnippet).build())),
		},

		// fold with scaling producing multiple values
		{
			name:           "success_osp_thousand_multiple_values",
			unit:           thousandUnit,
			resolution:     resolutionNull,
			metricSelector: openSecurityProblemsByTypeMetricSelector,
			handlerAdditionalSetupFunc: func(handler *test.CombinedURLHandler, testVariantDataFolder string) {
				addRequestsToHandlerForSuccessfulMetricsQueryWithFoldAndUnitsConversionSnippet(handler, testVariantDataFolder, openSecurityProblemsByTypeRequestBuilder, divideByThousandConversionSnippet)
			},
			getSLIFinishedEventAssertionsFunc: getSLIFinishedEventSuccessAssertionsFunc,
			sliResultAssertionsFuncs: toSlice(
				createSuccessfulSLIResultAssertionsFunc(sliName+"_third-party_vulnerability", 0.0917177307425399, openSecurityProblemsByTypeRequestBuilder.copyWithFold().copyWithMetricSelectorConversionSnippet(divideByThousandConversionSnippet).build()),
				createSuccessfulSLIResultAssertionsFunc(sliName+"_code-level_vulnerability", 0.016, openSecurityProblemsByTypeRequestBuilder.copyWithFold().copyWithMetricSelectorConversionSnippet(divideByThousandConversionSnippet).build())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			handler := createHandlerWithTemplatedDashboard(t,
				filepath.Join(testDataFolder, "dashboard.template.json"),
				struct {
					Unit           string
					MetricSelector string
					Resolution     string
				}{
					Unit:           tt.unit,
					MetricSelector: tt.metricSelector,
					Resolution:     tt.resolution,
				})

			testVariantDataFolder := filepath.Join(testDataFolder, tt.name)
			tt.handlerAdditionalSetupFunc(handler, testVariantDataFolder)
			runGetSLIsFromDashboardTestAndCheckSLIs(t, handler, testGetSLIEventData, tt.getSLIFinishedEventAssertionsFunc, tt.sliResultAssertionsFuncs...)

		})
	}
}

func addRequestsToHandlerForFailedMetricsQueryWithResolutionInfAndUnitsConversionSnippet(handler *test.CombinedURLHandler, testDataFolder string, requestBuilder *metricsV2QueryRequestBuilder, metricSelectorConversionSnippet string) string {
	addRequestsToHandlerForSuccessfulMetricsQueryWithResolutionInf(handler, testDataFolder, requestBuilder)

	// note: no additional metrics definition needs to be added as the metric selector is the same
	finalExpectedMetricsRequest := requestBuilder.copyWithResolution(resolutionInf).copyWithMetricSelectorConversionSnippet(metricSelectorConversionSnippet).build()
	handler.AddExactError(finalExpectedMetricsRequest, 400, filepath.Join(testDataFolder, metricsQueryFilename3))

	return finalExpectedMetricsRequest
}

func toSlice[V any](v ...V) []V {
	return v
}

package sli

import (
	"fmt"
	"testing"

	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI could not be found because of a misspelled indicator name - e.g. 'response_time_p59' instead of 'response_time_p95'
//   - this would have lead to a fallback to default SLIs, but should return an error now.
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButIndicatorCannotBeMatched(t *testing.T) {
	// no need to have something here, because we should not send an API request
	handler := test.NewFileBasedURLHandler(t)

	// error here in the misspelled indicator:
	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		"response_time_p59": "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	assertThatCustomSLITestIsCorrect(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "SLI definition", "not found"))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI is valid YAML, but Dynatrace cannot process the query correctly and returns a 400 error
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreValidYAMLButQueryIsNotValid(t *testing.T) {
	// error here: metric(s)Selector=
	handler := test.NewFileBasedURLHandler(t)

	// error here as well: metric(s)Selector=
	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorResponseTimeP95: "metricsSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_project:sockshop),tag(keptn_stage:staging)",
	})

	assertThatCustomSLITestIsCorrect(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, "error parsing Metrics v2 query"))
}

// In case we do not use the dashboard for defining SLIs we can use the file 'dynatrace/sli.yaml'.
//
// prerequisites:
// * a file called 'dynatrace/sli.yaml' exists and a SLI that we would want to evaluate (as defined in the slo.yaml) is defined
// * the defined SLI has errors, so parsing the YAML file would not be possible
func TestNoDefaultSLIsAreUsedWhenCustomSLIsAreInvalidYAML(t *testing.T) {
	// make sure we would not be able to query any metric due to a parsing error
	handler := test.NewFileBasedURLHandler(t)

	const errorMessage = "invalid YAML file - some parsing issue"
	rClient := newResourceClientMockWithGetSLIsError(t, fmt.Errorf(errorMessage))

	assertThatCustomSLITestIsCorrect(t, handler, testIndicatorResponseTimeP95, rClient, getSLIFinishedEventFailureAssertionsFunc, createFailedSLIResultAssertionsFunc(testIndicatorResponseTimeP95, errorMessage))
}

// TestRetrieveMetricsFromFile_SecurityProblemsV2 tests the success case for file-based SecurityProblemsV2 SLIs.
func TestRetrieveMetricsFromFile_SecurityProblemsV2(t *testing.T) {
	const (
		securityProblemsRequest           = "/api/v2/securityProblems?from=1632834999000&securityProblemSelector=status%28%22open%22%29&to=1632835299000"
		testDataFolder                    = "./testdata/sli_files/secpv2_success/"
		testIndicatorSecurityProblemCount = "security_problem_count"
	)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(securityProblemsRequest, testDataFolder+"security_problems_status_open.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorSecurityProblemCount: "SECPV2;securityProblemSelector=status(\"open\")",
	})

	assertThatCustomSLITestIsCorrect(t, handler, testIndicatorSecurityProblemCount, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSecurityProblemCount, 103, securityProblemsRequest))
}

// TestRetrieveMetricsFromFile_ProblemsV2 tests the success case for file-based ProblemsV2 SLIs.
func TestRetrieveMetricsFromFile_ProblemsV2(t *testing.T) {
	const (
		problemsRequest           = dynatrace.ProblemsV2Path + "?from=1632834999000&problemSelector=status%28%22open%22%29&to=1632835299000"
		testDataFolder            = "./testdata/sli_files/pv2_success/"
		testIndicatorProblemCount = "problem_count"
	)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(problemsRequest, testDataFolder+"problems_status_open.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorProblemCount: "PV2;problemSelector=status(\"open\")",
	})

	assertThatCustomSLITestIsCorrect(t, handler, testIndicatorProblemCount, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorProblemCount, 0, problemsRequest))
}

// TestRetrieveMetricsFromFile_SLO tests the success case for file-based SLO SLIs.
func TestRetrieveMetricsFromFile_SLO(t *testing.T) {
	const (
		sloRequest            = dynatrace.SLOPath + "/7d07efde-b714-3e6e-ad95-08490e2540c4?from=1632834999000&timeFrame=GTF&to=1632835299000"
		testDataFolder        = "./testdata/sli_files/slo_success/"
		testIndicatorSLOValue = "slo_value"
	)

	handler := test.NewFileBasedURLHandler(t)
	handler.AddExact(sloRequest, testDataFolder+"slo_7d07efde-b714-3e6e-ad95-08490e2540c4.json")

	rClient := newResourceClientMockWithSLIs(t, map[string]string{
		testIndicatorSLOValue: "SLO;7d07efde-b714-3e6e-ad95-08490e2540c4",
	})

	assertThatCustomSLITestIsCorrect(t, handler, testIndicatorSLOValue, rClient, getSLIFinishedEventSuccessAssertionsFunc, createSuccessfulSLIResultAssertionsFunc(testIndicatorSLOValue, 95, sloRequest))
}

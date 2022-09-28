package action

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn-contrib/dynatrace-service/internal/dynatrace"
	"github.com/keptn-contrib/dynatrace-service/internal/test"
)

func nilAttachRulesFunc() *dynatrace.AttachRules { return nil }

// multiple PGIs found will be returned, when no custom rules are defined
func TestEventHandlers_HandleEvent_MultipleEntities(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), filepath.Join(testdataFolder, "multiple_entities.json"))
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_multiple_200.json"))

	expectedAttachRules := getPGIOnlyAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, nilAttachRulesFunc, expectedAttachRules, nil)
	setups.assertAllEventsCorrectlySent(t, handler)
}

// multiple PGIs found will be returned, when no custom rules are defined and version information is based on event label
func TestEventHandlers_HandleEvent_MultipleEntitiesBasedOnLabel(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), filepath.Join(testdataFolder, "multiple_entities.json"))
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_multiple_200.json"))

	labels := map[string]string{
		"releasesVersion": "1.2.3",
	}

	expectedAttachRules := getPGIOnlyAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, nilAttachRulesFunc, expectedAttachRules, labels)
	setups.assertAllEventsCorrectlySent(t, handler)
}

// single PGI found will be combined with the custom attach rules
func TestEventHandlers_HandleEvent_SingleEntityAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), filepath.Join(testdataFolder, "single_entity.json"))
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_multiple_200.json"))

	expectedAttachRules := getCustomAttachRulesWithEntityIds("PROCESS_GROUP_INSTANCE-D23E64F62FDC200A")

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, getCustomAttachRules, expectedAttachRules, nil)
	setups.assertAllEventsCorrectlySent(t, handler)
}

// single PGI found will be combined with the custom attach rules that include version information from event label
func TestEventHandlers_HandleEvent_SingleEntityAndUserSpecifiedAttachRulesBasedOnLabel(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), filepath.Join(testdataFolder, "single_entity.json"))
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_single_200.json"))

	labels := map[string]string{
		"releasesVersion": "1.2.3",
	}

	expectedAttachRules := getCustomAttachRulesWithEntityIds("PROCESS_GROUP_INSTANCE-D23E64F62FDC200A")

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, getCustomAttachRules, expectedAttachRules, labels)
	setups.assertAllEventsCorrectlySent(t, handler)
}

// no PGIs found and no custom attach rules will result in default attach rules
func TestEventHandlers_HandleEvent_NoEntitiesAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), filepath.Join(testdataFolder, "no_entity.json"))
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_single_200.json"))

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, nilAttachRulesFunc, expectedAttachRules, nil)
	setups.assertAllEventsCorrectlySent(t, handler)
}

// no deployment.started resp. deployment.finished events found and no custom attach rules will result in default attach rules
// does not apply to evaluation finished, because the timeframe is taken from the event payload
// does only partially apply for deployment finished - there's a different logic that is tested separately
func TestEventHandlers_HandleEvent_NoEventsFoundAndNoCustomAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_single_200.json"))

	expectedAttachRules := getDefaultAttachRules()

	testConfigs := []struct {
		name    string
		eClient *eventClientFake
	}{
		{
			name: "deployment.started event not found",
			eClient: &eventClientFake{
				t:           t,
				imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
				eventTimestamps: timestampsForType{
					"sh.keptn.event.deployment.started": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.started event"),
					},
					"sh.keptn.event.deployment.finished": {
						time: time.Unix(1654000313, 0),
						err:  nil,
					},
				},
			},
		},
		{
			name: "deployment.finished event not found",
			eClient: &eventClientFake{
				t:           t,
				imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
				eventTimestamps: timestampsForType{
					"sh.keptn.event.deployment.started": {
						time: time.Unix(1654000240, 0),
						err:  nil,
					},
					"sh.keptn.event.deployment.finished": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.finished event"),
					},
				},
			},
		},
		{
			name: "both deployment events not found",
			eClient: &eventClientFake{
				t:           t,
				imageAndTag: common.NewImageAndTag("registry/my-image", "1.2.3"),
				eventTimestamps: timestampsForType{
					"sh.keptn.event.deployment.started": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.started event"),
					},
					"sh.keptn.event.deployment.finished": {
						time: time.Time{},
						err:  fmt.Errorf("could not find deployment.finished event"),
					},
				},
			},
		},
	}

	for _, testConfig := range testConfigs {
		t.Run(testConfig.name, func(t *testing.T) {
			setups := testSetups{
				aeSetups: []testSetup[dynatrace.AnnotationEvent]{
					testTriggeredTestSetup{
						t:                   t,
						handler:             handler,
						eClient:             testConfig.eClient,
						customAttachRules:   nilAttachRulesFunc(),
						expectedAttachRules: expectedAttachRules,
						labels:              nil,
					},
					testFinishedTestSetup{
						t:                   t,
						handler:             handler,
						eClient:             testConfig.eClient,
						customAttachRules:   nilAttachRulesFunc(),
						expectedAttachRules: expectedAttachRules,
						labels:              nil,
					},
				},
				ieSetups: []testSetup[dynatrace.InfoEvent]{
					releaseTriggeredTestSetup{
						t:                   t,
						handler:             handler,
						eClient:             testConfig.eClient,
						customAttachRules:   nilAttachRulesFunc(),
						expectedAttachRules: expectedAttachRules,
						labels:              nil,
					},
				},
			}

			setups.assertAllEventsCorrectlySent(t, handler)
		})
	}
}

// no PGIs found but custom attach rules will result custom attach rules only
func TestEventHandlers_HandleEvent_NoEntitiesAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact(getDefaultPGIQuery(), filepath.Join(testdataFolder, "no_entity.json"))
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_single_200.json"))

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewImageAndTag("registry/my-image", "1.2.3"),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, getCustomAttachRules, *customAttachRules, nil)
	setups.assertAllEventsCorrectlySent(t, handler)
}

// no entities will be queried, because there is no version information. Default attach rules will be returned if there are no custom rules
func TestEventHandlers_HandleEvent_NoVersionInformationAndNoUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_single_200.json"))

	expectedAttachRules := getDefaultAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, nilAttachRulesFunc, expectedAttachRules, nil)
	setups.assertAllEventsCorrectlySent(t, handler)
}

// no entities will be queried, because there is no version information. Custom attach rules will be returned if they are present
func TestEventHandlers_HandleEvent_NoVersionInformationAndUserSpecifiedAttachRules(t *testing.T) {
	handler := test.NewFileBasedURLHandlerWithSink(t)
	handler.AddExact("/api/v1/events", filepath.Join(testdataFolder, "events_response_single_200.json"))

	customAttachRules := getCustomAttachRules()

	eClient := &eventClientFake{
		t:               t,
		imageAndTag:     common.NewNotAvailableImageAndTag(),
		eventTimestamps: setupCorrectTimestampResultsForDeploymentTimeframe(),
	}

	setups := createAllTestSetups(t, handler, eClient, getCustomAttachRules, *customAttachRules, nil)
	setups.assertAllEventsCorrectlySent(t, handler)
}

type testSetups struct {
	aeSetups []testSetup[dynatrace.AnnotationEvent]
	ieSetups []testSetup[dynatrace.InfoEvent]
	deSetups []testSetup[dynatrace.DeploymentEvent]
}

func (s testSetups) assertAllEventsCorrectlySent(t *testing.T, handler *test.FileBasedURLHandlerWithSink) {
	for _, aeSetup := range s.aeSetups {
		t.Run(fmt.Sprintf("%T", aeSetup), func(t *testing.T) {
			assertThatCorrectEventWasSent[dynatrace.AnnotationEvent](t, handler, aeSetup)
		})

	}

	for _, ieSetup := range s.ieSetups {
		t.Run(fmt.Sprintf("%T", ieSetup), func(t *testing.T) {
			assertThatCorrectEventWasSent[dynatrace.InfoEvent](t, handler, ieSetup)
		})
	}

	for _, deSetup := range s.deSetups {
		t.Run(fmt.Sprintf("%T", deSetup), func(t *testing.T) {
			assertThatCorrectEventWasSent[dynatrace.DeploymentEvent](t, handler, deSetup)
		})
	}
}

// we need a function for custom attach rules here, otherwise we would need to deep-copy the attach rules in createOrUpdateAttachRules
func createAllTestSetups(t *testing.T, handler *test.FileBasedURLHandlerWithSink, eClient *eventClientFake, customAttachRulesFunc func() *dynatrace.AttachRules, expectedAttachRules dynatrace.AttachRules, labels map[string]string) testSetups {
	return testSetups{
		aeSetups: []testSetup[dynatrace.AnnotationEvent]{
			testTriggeredTestSetup{
				t:                   t,
				handler:             handler,
				eClient:             eClient,
				customAttachRules:   customAttachRulesFunc(),
				expectedAttachRules: expectedAttachRules,
				labels:              labels,
			},
			testFinishedTestSetup{
				t:                   t,
				handler:             handler,
				eClient:             eClient,
				customAttachRules:   customAttachRulesFunc(),
				expectedAttachRules: expectedAttachRules,
				labels:              labels,
			},
		},
		ieSetups: []testSetup[dynatrace.InfoEvent]{
			evaluationFinishedTestSetup{
				t:                   t,
				handler:             handler,
				eClient:             eClient,
				customAttachRules:   customAttachRulesFunc(),
				expectedAttachRules: expectedAttachRules,
				labels:              labels,
			},
			releaseTriggeredTestSetup{
				t:                   t,
				handler:             handler,
				eClient:             eClient,
				customAttachRules:   customAttachRulesFunc(),
				expectedAttachRules: expectedAttachRules,
				labels:              labels,
			},
		},
		deSetups: []testSetup[dynatrace.DeploymentEvent]{
			deploymentFinishedTestSetup{
				t:                   t,
				handler:             handler,
				eClient:             eClient,
				customAttachRules:   customAttachRulesFunc(),
				expectedAttachRules: expectedAttachRules,
				labels:              labels,
			},
		},
	}
}

func setupCorrectTimestampResultsForDeploymentTimeframe() timestampsForType {
	return timestampsForType{
		"sh.keptn.event.deployment.started": {
			time: time.Unix(1654000240, 0),
			err:  nil,
		},
		"sh.keptn.event.deployment.finished": {
			time: time.Unix(1654000313, 0),
			err:  nil,
		},
	}
}

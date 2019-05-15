# Release Notes 0.1.0

## Release Goal

This service sends information about the current state of a pipeline run for a service to Dynatrace by sending events for the correlating detected service.
The service is subscribed to the following keptn events:

- `sh.keptn.events.deployment-finished`
When an event of this type is received, a *Deployment Info* event is sent to Dynatrace, indicating that a new version of a service has been deployed. Additionally, a *Custom Info* event is sent, which marks the start of a test run execution for this service.
*NOTE*: It's planned to have an additional keptn event in later releases to better determine the exact time of the test run execution.
- `sh.keptn.events.tests-finished`
Receiving this event causes a *Custom Info* event to be sent to Dynatrace. This event marks the end of a test run execution in Dynatrace.
- `sh.keptn.events.evaluation-done`
Receiving this event causes a *Custom Info* event to be sent to Dynatrace. This event indicates wether an artifact will be promoted to the next stage (i.e., the tests were successful), or if it is rejected.

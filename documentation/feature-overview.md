# Feature overview

The dynatrace-service supports the following use cases:

- [**SLI-provider**](sli-provider.md): To support the evaluation of the quality gates, the dynatrace-service can be configured to retrieve SLIs for a Keptn project, stage or service. Two modes are available: [SLIs via a combination of `SLI.yaml` files](slis-via-files.md) located on the Keptn service, stage and project, or [SLIs and SLOs based on a Dynatrace dashboard](slis-via-dashboard.md).

- [**Forwarding events from Keptn to Dynatrace**](event-forwarding-to-dynatrace.md): The dynatrace-service can forward events such as remediation, deployment, test start/stop, evaluation or release events to Dynatrace using attach rules to ensure that the correct monitored entities are associated with the event.

- [**Forwarding problem notifications from Dynatrace to Keptn**](problem-forwarding-to-keptn.md): The dynatrace-service can support triggering remediation sequences by forwarding problem notifications from Dynatrace to a Keptn environment and ensuring that the `sh.keptn.events.problem` event is mapped to the correct project, service and stage.

- [**Automatic onboarding of monitored service entities**](auto-service-onboarding.md): The dynatrace-service can be configured to periodically check for new service entities detected by Dynatrace and automatically import these into Keptn.


## Keptn events

The dynatrace-service listens for the following events:

- `sh.keptn.event.get-sli.triggered`
- `sh.keptn.event.action.triggered`
- `sh.keptn.event.action.started`
- `sh.keptn.event.action.finished`
- `sh.keptn.event.deployment.finished`
- `sh.keptn.event.test.triggered`
- `sh.keptn.event.test.finished`
- `sh.keptn.event.evaluation.finished`
- `sh.keptn.event.release.triggered`
- `sh.keptn.event.release.finished`
- `sh.keptn.events.problem`
- `sh.keptn.event.monitoring.configure`


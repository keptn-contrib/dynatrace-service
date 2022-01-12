# Forwarding problem notifications from Dynatrace to Keptn

To allow a Dynatrace problem to trigger a remediation workflow in Keptn, the dynatrace-service listens for `sh.keptn.events.problem` events originating from Dynatrace (i.e. with `"source": "dynatrace"`) and emits corresponding `sh.keptn.event.<stage>.remediation.triggered` events for problems with `State="OPEN"` or `sh.keptn.events.problem` events for problems with `State="RESOLVED"`. All relevant problem details such as `PID`, `ProblemTitle` and `ProblemURL`are forwarded.

## Routing problem notifications to a specific Keptn service

To route problem notifications to a specific Keptn service, simply specify the project, stage and service using the `KeptnProject`, `KeptnStage` and `KeptnService` fields in the Dynatrace custom notification integration payload. For example, the following payload will send problems to a Keptn project called `dynatrace`, stage called `production` and service called `allproblems`:

```json
{
    "specversion":"1.0",
    "shkeptncontext":"{PID}",
    "type":"sh.keptn.events.problem",
    "source":"dynatrace",
    "id":"{PID}",
    "time":"",
    "contenttype":"application/json",
    "data": {
        "State":"{State}",
        "ProblemID":"{ProblemID}",
        "PID":"{PID}",
        "ProblemTitle":"{ProblemTitle}",
        "ProblemURL":"{ProblemURL}",
        "ProblemDetails":{ProblemDetailsJSON},
        "ImpactedEntities":{ImpactedEntities},
        "ImpactedEntity":"{ImpactedEntity}",
        "KeptnProject" : "dynatrace",
        "KeptnStage" : "production",
        "KeptnService" : "allproblems"        
    }
}
```

## Routing problem notifications for problems detected in Keptn deployed services

If you use Keptn to deploy your microservices and follow the standard tagging practices, Dynatrace will tag your monitored services with `keptn_project`, `keptn_service` and `keptn_stage`. By including the `Tags` field in the payload, the dynatrace-service will use the values of these tags from the impacted entities to ensure that the event is mapped to the correct Keptn project, service and stage. This is demonstrated in the following custom notification integration payload: 

```json
{
    "specversion":"1.0",
    "shkeptncontext":"{PID}",
    "type":"sh.keptn.events.problem",
    "source":"dynatrace",
    "id":"{PID}",
    "time":"",
    "contenttype":"application/json",
    "data": {
        "State":"{State}",
        "ProblemID":"{ProblemID}",
        "PID":"{PID}",
        "ProblemTitle":"{ProblemTitle}",
        "ProblemURL":"{ProblemURL}",
        "ProblemDetails":{ProblemDetailsJSON},
        "Tags":"{Tags}",
        "ImpactedEntities":{ImpactedEntities},
        "ImpactedEntity":"{ImpactedEntity}",
        "KeptnProject" : "dynatrace",
    }
}
```

If present, these tags override any values specified in the `KeptnProject`, `KeptnStage` and `KeptnService` fields described above.

The dynatrace-service can [configure this feature automatically in a Dynatrace tenant](auto-tenant-configuration.md#problem-notifications).

**Notes**
1. The dynatrace-service requires a valid project to process problem events. We recommend always including a `KeptnProject` field set to a valid project in the custom notification integration payload definition.
2. Dynatrace alerting profiles can be used to filter certain problem types, e.g. infrastructure problems in production or slow performance in a developer environment. By creating a Keptn project to handle these remediation workflows and a Keptn service for each alerting profile, it is easy to define workflows for particular problem types. Furthermore, individual environment names such as `pre-prod` or `production` can be represented as stages within the project.

Here is a screenshot of a workflow triggered by a Dynatrace problem and how it then executes in Keptn:

![Workflow triggered by a Dynatrace problem](images/remediation_workflow.png "Workflow triggered by a Dynatrace problem")

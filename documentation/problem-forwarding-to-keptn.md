# Forwarding problem notifications from Dynatrace to Keptn

If you use Keptn to deploy your microservices and follow the standard tagging practices, Dynatrace will tag your monitored services with `keptn_project`, `keptn_service` and `keptn_stage`. If Dynatrace then detects a problem in one of these deployed services, e.g: High Failure Rate, Slow response time, you can let Dynatrace send these problems back to Keptn. To allow a Dynatrace problem to trigger a remediation workflow in Keptn, the dynatrace-service can intercept `sh.keptn.events.problem` events and forward the payload as a `sh.keptn.event.problem.open` event to the correct Keptn project, service and stage. All relevant problem details such as `PID`, `ProblemTitle` and `ProblemURL`are included.


## Setting up problem notifications for problems detected on Keptn deployed services

To enable this feature, add the follow Custom Problem Notification to your Dynatrace tenant:

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
    }
}
```

The dynatrace-service can [configure this automatically in a Dynatrace tenant](auto-tenant-configuration.md#problem-notifications).

The dynatrace-service will parse the `Tags` field and use the `keptn_project`, `keptn_service` and `keptn_stage` tags that come directly from the impacted entities to ensure that the event is mapped to the correct project service and stage.

*Note:* Dynatrace Alerting Profiles that include problems on services without Keptn tags cannot be mapped through this capability.


## Setting up problem notifications for any type of detected problem, e.g., infrastructure

To send any type of problem to Keptn and orchestrate auto-remediation workflows, simply specify the project, service and stage as part of the data structure that is sent to Keptn. For example, the following Dynatrace custom notification integration payload will send all problems to a Keptn project called `dynatrace`, stage called `production` and service called `allproblems`:

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
        "KeptnProject" : "demo-remediation",
        "KeptnService" : "allproblems",
        "KeptnStage" : "production"
    }
}
```

When the dynatrace-service receives this `sh.keptn.events.problem` event it will parse the fields `KeptnProject`, `KeptnService` and `KeptnStage` and will then send a `sh.keptn.event.problem.open` event to Keptn including the rest of the problem details. This allows you to send any type of Dynatrace detected problem to Keptn and let Keptn execute a remediation workflow.

*Best Practice:* Dynatrace alerting profiles can be used to filter certain problem types, e.g: infrastructure problems in production or slow Performance in developer environment. By creating a Keptn project to handle these remediation workflows and a Keptn service for each alerting profile, it is easy to define workflows for particular problem types. Furthermore, individual environment names such as `pre-prod` or `production` can be represented as stages within the project.

Here is a screenshot of a workflow triggered by a Dynatrace problem and how it then executes in Keptn:

![Workflow triggered by a Dynatrace problem](images/remediation_workflow.png "Workflow triggered by a Dynatrace problem")

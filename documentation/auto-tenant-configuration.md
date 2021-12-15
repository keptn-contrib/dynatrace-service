# Automatic configuration of a Dynatrace tenant

This section describes the configuration entities created by the dynatrace-service on the Dynatrace tenant when it receives a `sh.keptn.event.monitoring.configure` event. This makes it easy to configure your Dynatrace tenant to fully interact with the Keptn installation using the `keptn configure monitoring` CLI command.

To enable or disable the creation of the following entity types, please see [Configuring automatic generation of Dynatrace entities](additional-installation-options.md#configuring-automatic-dynatrace-tenant-configuration). Once processing of the configure monitoring event is complete, the dynatrace-service sends a `sh.keptn.event.configure-monitoring.finished` event with a summary of the operations performed.


## Tagging rules

When `dynatraceService.config.generateTaggingRules` is set to `true`, the dynatrace-service will create tagging rules for `keptn_service`, `keptn_stage`, `keptn_project`, `keptn_deployment` tags. For example the rule for `keptn_project` is created as follows:

```json
{
    "name": "keptn_project",
    "rules": [
        {
            "type": "SERVICE",
            "enabled": true,
            "valueFormat": "{ProcessGroup:Environment:keptn_project}",
            "propagationTypes": [
                "SERVICE_TO_PROCESS_GROUP_LIKE"
            ],
            "conditions": [
                {
                    "key": {
                        "attribute": "PROCESS_GROUP_CUSTOM_METADATA",
                        "dynamicKey": {
                            "source": "ENVIRONMENT",
                            "key": "keptn_project"
                        },
                        "type": "PROCESS_CUSTOM_METADATA_KEY"
                    },
                    "comparisonInfo": {
                        "type": "STRING",
                        "operator": "EXISTS",
                        "value": null,
                        "negate": false,
                        "caseSensitive": null
                    }
                }
            ]
        }
    ]
}
```


## Problem notifications

When `dynatraceService.config.generateProblemNotifications` is set to `true`, the dynatrace-service will try to create a problem alerting profile named `Keptn` with rules for `AVAILABILITY`, `ERROR`, `PERFORMANCE`, `RESOURCE_CONTENTION`, `CUSTOM_ALERT` and `MONITORING_UNAVAILABLE` that trigger problem notifications after 0 minutes for all entities in all management zones. If an alerting profile is already available it is not overwritten.

The alerting profile is then used to create a webhook named `Keptn Problem Notification` to send problem events to Keptn using the event API. The webhook has the following form:

```json
{
    "type": "WEBHOOK",
    "name": "Keptn Problem Notification",
    "alertingProfile": "<alerting-profile-id>",
    "active": true,
    "url": "<keptn-endpoint>/api/v1/event",
    "acceptAnyCertificate": true,
    "headers": [
        {
            "name": "x-token",
            "value": "<keptn-api-token>"
        },
        {
            "name": "Content-Type",
            "value": "application/cloudevents+json"
        }
    ],
    "payload": "<payload>"
}
```

Values are set for `<alerting-profile-id>`, `<keptn-endpoint>` and `<keptn-api-token>`. The actual template, added as `<payload>`, has the form:

```json
{
    "specversion": "1.0",
    "type": "sh.keptn.events.problem",
    "shkeptncontext": "{PID}",
    "source": "dynatrace",
    "id": "{PID}",
    "time": "",
    "contenttype": "application/json",
    "data": {
        "State": "{State}",
        "ProblemID": "{ProblemID}",
        "PID": "{PID}",
        "ProblemTitle": "{ProblemTitle}",
        "ProblemURL": "{ProblemURL}",
        "ProblemDetails": {ProblemDetailsJSON},
        "Tags": "{Tags}",
        "ImpactedEntities": {ImpactedEntities},
        "ImpactedEntity": "{ImpactedEntity}"
    }
}
```

If a problem notification named `Keptn Problem Notification` already exists it is overwritten.


## Management zones

When `dynatraceService.config.generateManagementZones` is set to `true`, the dynatrace-service tries to create a management zone for the project and for each stage it contains. The project management zone, named `Keptn: <project-name>`, contains services tagged with `keptn_project: <project-name>`, whereas each stage management zone, named `Keptn: <project-name> <stage-name>`, contains services tagged with `keptn_project: <project-name>` and `keptn_stage: <stage-name>`. If a management zone with the same name already exists, it is not overwritten.


## Dashboards

When `dynatraceService.config.generateDashboards` is set to `true`, the dynatrace-service creates (or overwrites) a dashboard called `<project-name>@keptn: Digital Delivery & Operations Dashboard`. The dashboard contains some basic infrastructure monitoring tiles for the health of hosts, CPU load and network status, as well as a default quality-gate comprised of service health, throughput, failure rate and response time.


## Metric events

When `dynatraceService.config.generateMetricEvents` is set to `true`, the dynatrace-service tries to create custom alerts for each service on each stage in the project based on the associated SLIs and SLOs.
# Forwarding events from Keptn to Dynatrace

The dynatrace-service will forward `sh.keptn.event.deployment.finished`, `sh.keptn.event.test.triggered`, `sh.keptn.event.test.finished`, `sh.keptn.event.evaluation.finished` and `sh.keptn.event.release.triggered` events to Dynatrace by creating the appropriate events in the Dynatrace tenant. For `sh.keptn.event.action.triggered`, `sh.keptn.event.action.started` and `sh.keptn.event.action.finished` events raised as part of a remediation action, it will create information and configuration events if a Dynatrace problem is associated with the event.


## Targeting specific entities using attach rules

By default, the dynatrace-service assumes that all events are sent to monitored entities that satisfy the following attach rules:

```yaml
attachRules:
  tagRule:
    - meTypes:
        - SERVICE
      tags:
        - context: CONTEXTLESS
          key: keptn_project
          value: $PROJECT
        - context: CONTEXTLESS
          key: keptn_service
          value: $SERVICE
        - context: CONTEXTLESS
          key: keptn_stage
          value: $STAGE
```

Most services that are deployed with Keptn's helm-service are automatically tagged like this. Here is a screenshot of how these tags show up in Dynatrace for a service deployed with Keptn:

![Keptn tags in Dynatrace](images/keptn_tags_in_dynatrace.png "Keptn tags in Dynatrace")

**Note**

For some events we will try to push the event payload to *PGI* level instead of *SERVICE* level. Please refer to the section [targeting specific entities for deployment, test and evaluation information](event-forwarding-to-dynatrace-to-specific-entities.md).


If your services are however not tagged with these but other tags or if you want the dynatrace-service to send the events not to a service but rather an application, process group or host, overwrite the default [attach rules in a `dynatrace/dynatrace.conf.yaml` file](dynatrace-conf-yaml-file.md#attach-rules-for-connecting-dynatrace-entities-with-events-attachrules).

The following example instructs the dynatrace-service to send its events to a monitored entity that holds a tag with the key that matches your Keptn service name (`$SERVICE`) as well as holds an additional auto-tag that defines the environment to be pulled from a label that has been sent to Keptn:

```yaml
---
spec_version: '0.1.0'
attachRules:
  tagRule:
    - meTypes:
        - SERVICE
      tags:
        - context: CONTEXTLESS
          key: $SERVICE
        - context: CONTEXTLESS
          key: environment
          value: $LABEL.environment
```

## Enriching events sent to Dynatrace with more context

The dynatrace-service sends `CUSTOM_DEPLOYMENT`, `CUSTOM_INFO` and `CUSTOM_ANNOTATION` events when it handles Keptn events such as `sh.keptn.event.deployment.finished`, `sh.keptn.event.test.finished`, `sh.keptn.event.release.triggered` or `sh.keptn.event.evaluation.finished`. The dynatrace-service will parse all labels in the Keptn event and will pass them on to Dynatrace as custom properties. This makes it easy to pass more context to Dynatrace, e.g: `ciBackLink` for a `CUSTOM_DEPLOYMENT` or ensure that things like Jenkins Job ID, Jenkins Job URL, etc. show up in Dynatrace as well. 


## Sending events to different Dynatrace environments per project, stage or service

To instruct the dynatrace-service to send events to a specific Dynatrace environment for a specific Keptn project, stage or service, overwrite the credentials secret name in a `dynatrace/dynatrace.conf.yaml` file and add it to the appropriate stage of the Keptn project.

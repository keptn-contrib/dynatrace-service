# Forwarding events from Keptn to Dynatrace

The dynatrace-service will forward `sh.keptn.event.deployment.finished`, `sh.keptn.event.test.triggered`, `sh.keptn.event.test.finished`, `sh.keptn.event.evaluation.finished`, `sh.keptn.event.release.triggered` and `sh.keptn.event.release.finished` events to Dynatrace by creating the appropriate events in the Dynatrace tenant. For `sh.keptn.event.action.triggered`, `sh.keptn.event.action.started` and `sh.keptn.event.action.finished` events raised as part of a remediation action, it will create information and configuration events if a Dynatrace problem is associated with the event.


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

## Targeting specific entities for evaluation information

As stated above, dynatrace-service will use the default attach rules in case users have not supplied their own via a `dynatrace/dynatrace.conf.yaml` file. While this is true for most event types, there is a special behaviour for `sh.keptn.event.evaluation.finished` events. These events will not be attached to the *Service* level, but to a certain *Process Group Instance* (aka. *Process*) if possible. This is done because a *Service* entity in Dynatrace can have multiple instances of *Processes* from different versions. So dynatrace-service tries to push evaluation information found in `sh.keptn.event.evaluation.finished` events to the *Process* entity identified by **version information**, instead of the generic *Service* entity. If the desired *Process* version could be found, then the event will also be available on *Service* level in addition to the *Process* level as it is propagated automatically by Dynatrace.

Currently, there are two ways of providing **version information** to dynatrace-service:

### Version information derived from a deployment task

If you run an evaluation *task* after a deployment *task* in the course of the same Keptn *sequence*, then dynatrace-service will extract the version information (the image *tag*) from the field `data.configurationChange.values.image` of the `sh.keptn.event.deployment.triggered` event payload.

Below you can see an exemplary payload of a `sh.keptn.event.deployment.triggered` event including the `configurationChange` data.
```json
{
  "data": {
    "configurationChange": {
      "values": {
        "image": "registry/my-service:v0.1.1",
        "version": "v0.1.1"
      }
    },
    "deployment": {
      "deploymentURIsLocal": null,
      "deploymentstrategy": "blue_green_service"
    },
    "project": "my-project",
    "service": "my-service",
    "stage": "my-stage"
  },
  "id": "3ab9545f-1f1f-480a-b8c7-ff8cc37e33ca",
  "shkeptncontext": "e88818e5-ca15-453a-90bd-187576744bef",
  "shkeptnspecversion": "0.2.4",
  "source": "shipyard-controller",
  "specversion": "1.0",
  "time": "2022-06-29T11:48:31.543736386Z",
  "gitcommitid": "9389d3b09c36cd3a5970315b10a57b4e571a1d4b",
  "type": "sh.keptn.event.deployment.triggered"
}
```

### Version information derived from event labels

While the above method would work well e.g. for *delivery* sequences, it does not work for a simple evaluation, that a user would want to do. For these scenarios, there is another way of supplying the version information by using event labels.

A user you can trigger an evaluation e.g. with the Keptn CLI:

```shell
keptn trigger evaluation --project="my-project" --stage="my-stage" --service="my-service" --start="2022-06-02T07:08:00" --end="2022-06-02T09:08:00"
```

If a user adds a label called `releasesVersion` then this will be picked up and dynatrace-service can use this as version information.

```shell
keptn trigger evaluation --project="my-project" --stage="my-stage" --service="my-service" --start="2022-06-02T07:08:00" --end="2022-06-02T09:08:00" --labels="releasesVersion=v0.1.1"
```

### Prerequisites 

In order to correctly identify a *Process* by its version in Dynatrace, you need to set it up beforehand. Please refer to [the official Dynatrace release monitoring documentation](https://www.dynatrace.com/support/help/how-to-use-dynatrace/cloud-automation/release-monitoring/version-detection-strategies) on how to achieve this.

### General approach

If your releases are monitored by Dynatrace, dynatrace-service can make use of the **version information** provided (either by deployment events or event labels) to query Dynatrace APIs in order to receive the desired *Process Group Instance* id(s) which will then be used in the attach rules.

### Version information order

If **version information** would be provided in both ways, then the information found in `sh.keptn.event.deployment.triggered` events will have precedence over the one found in the event label `releasesVersion`.

### How attach rules are created

* Version information is not provided for dynatrace-service
    * either default attach rules are used, as described [above](#targeting-specific-entities-using-attach-rules), or
    * user provided attach rules are used if available
* Version information is available
    * if *Process Group Instance* ids could be retrieved, then
        * either only these are used, or
        * they are combined with user defined attach rules if available
    * if *Process Group Instance* ids could not be retrieved, then
        * either default attach rules are used, or
        * user provided attach rules are used if available

## Enriching events sent to Dynatrace with more context

The dynatrace-service sends `CUSTOM_DEPLOYMENT`, `CUSTOM_INFO` and `CUSTOM_ANNOTATION` events when it handles Keptn events such as `sh.keptn.event.deployment.finished`, `sh.keptn.event.test.finished` or `sh.keptn.event.evaluation.finished`. The dynatrace-service will parse all labels in the Keptn event and will pass them on to Dynatrace as custom properties. This makes it easy to pass more context to Dynatrace, e.g: `ciBackLink` for a `CUSTOM_DEPLOYMENT` or ensure that things like Jenkins Job ID, Jenkins Job URL, etc. show up in Dynatrace as well. 


## Sending events to different Dynatrace environments per project, stage or service

To instruct the dynatrace-service to send events to a specific Dynatrace environment for a specific Keptn project, stage or service, overwrite the credentials secret name in a `dynatrace/dynatrace.conf.yaml` file and add it to the appropriate stage of the Keptn project.
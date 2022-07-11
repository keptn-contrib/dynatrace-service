# Targeting specific entities for deployment, test, evaluation and release information

As stated in the section [targeting specific entities using attach rules](event-forwarding-to-dynatrace.md#targeting-specific-entities-using-attach-rules), the dynatrace-service will use the default attach rules in case users have not supplied their own via a `dynatrace/dynatrace.conf.yaml` file. While this is true for some event types, there is a special behavior for `sh.keptn.event.deployment.finished`, `sh.keptn.event.test.triggered`, `sh.keptn.event.test.finished`, `sh.keptn.event.evaluation.finished` and `sh.keptn.event.release.triggered` events. 

These events will not be attached to the *Service* level, but to a certain *Process Group Instance* (aka. *Process*) if possible. This is done because a *Service* entity in Dynatrace can consist of multiple *Processes* of different versions. So the dynatrace-service tries to push the information found in these events to the *Process* entity identified by **version information**, instead of the generic *Service* entity. If the desired *Process* version could be found, then the event will also be available on *Service* level in addition to the *Process* level as it is propagated automatically by Dynatrace.

Currently, there are two ways of providing **version information** to the dynatrace-service:

### Version information derived from a deployment task

If you run an evaluation *task* after a deployment *task* in the course of the same Keptn *sequence*, then the dynatrace-service will extract the version information (the image *tag*) from the field `data.configurationChange.values.image` of the `sh.keptn.event.deployment.triggered` event payload.

Below you can see an exemplary payload of a `sh.keptn.event.deployment.triggered` event including the `configurationChange` data.
```json
{
  "data": {
    "configurationChange": {
      "values": {
        "image": "registry/my-service:v0.1.1"
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

While the above method would work well for *delivery* sequences, if such version information is not available during e.g. a simple user-triggered evaluation, the version number of your release can be supplied by providing an event label called `releasesVersion`.

As a user, you can trigger an evaluation e.g. with the Keptn CLI:

```shell
keptn trigger evaluation --project="my-project" --stage="my-stage" --service="my-service" --start="2022-06-02T07:08:00" --end="2022-06-02T09:08:00"
```

By simply adding a label called `releasesVersion`, the dynatrace-service can use this as version information.

```shell
keptn trigger evaluation --project="my-project" --stage="my-stage" --service="my-service" --start="2022-06-02T07:08:00" --end="2022-06-02T09:08:00" --labels="releasesVersion=v0.1.1"
```

### Prerequisites 

In order to correctly identify a *Process* by its version in Dynatrace, you need to set it up beforehand. Please refer to [the official Dynatrace release monitoring documentation](https://www.dynatrace.com/support/help/how-to-use-dynatrace/cloud-automation/release-monitoring/version-detection-strategies) on how to achieve this.

### General approach

If your releases are monitored by Dynatrace, the dynatrace-service can make use of the **version information** provided (either by deployment events or event labels) to query Dynatrace APIs in order to receive the desired *Process Group Instance* ID(s) which will then be used in the attach rules.

**Note**:

* If you are using the helm-service to deploy your service and use a deployment strategy other than `user_managed` then you need to use the property `image` defined in your `values.yaml` file inside your `deployment.yaml` accordingly. Please refer to the [Keptn documentation](https://keptn.sh/docs/0.16.x/continuous_delivery/deployment_helm/#direct-deployments) or these example files: [values.yaml](https://github.com/keptn/examples/blob/0.11.0/onboarding-carts/carts/values.yaml#L1), [deployment.yaml](https://github.com/keptn/examples/blob/0.11.0/onboarding-carts/carts/templates/deployment.yaml#L24) for more details.

* If you are using a deployment strategy of `user_managed` then you can also use the label - as mentioned above - to provide **version information**.

### Version information order

If **version information** would be provided in both ways, then the information found in `sh.keptn.event.deployment.triggered` events will have precedence over the one found in the event label `releasesVersion`.

### How attach rules are created
* If version information is available to the dynatrace-service:
    * if *Process Group Instance* IDs can be retrieved, then
        * either only these are used, or
        * they are combined with user defined attach rules if available
    * if *Process Group Instance* IDs could not be retrieved, then
        * either default attach rules are used, or
        * user provided attach rules are used if available

* If version information is not available:
    * either default attach rules are used, as described in the section [targeting specific entities using attach rules](event-forwarding-to-dynatrace.md#targeting-specific-entities-using-attach-rules), or
    * user provided attach rules are used (if available)

## Enriching events sent to Dynatrace with more context

The dynatrace-service sends `CUSTOM_DEPLOYMENT`, `CUSTOM_INFO` and `CUSTOM_ANNOTATION` events when it handles Keptn events such as `sh.keptn.event.deployment.finished`, `sh.keptn.event.test.finished`, `sh.keptn.event.release.triggered` or `sh.keptn.event.evaluation.finished`. The dynatrace-service will parse all labels in the Keptn event and will pass them on to Dynatrace as custom properties. This makes it easy to pass more context to Dynatrace, e.g: `ciBackLink` for a `CUSTOM_DEPLOYMENT` or ensure that things like Jenkins Job ID, Jenkins Job URL, etc. show up in Dynatrace as well. 


## Sending events to different Dynatrace environments per project, stage or service

To instruct the dynatrace-service to send events to a specific Dynatrace environment for a specific Keptn project, stage or service, overwrite the credentials secret name in a `dynatrace/dynatrace.conf.yaml` file and add it to the appropriate stage of the Keptn project.

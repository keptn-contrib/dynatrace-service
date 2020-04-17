# (grabnerandi in dev) Dynatrace Service

This is the readme information including changes that come in with this branch. The standard documentation for this service can be found further down!

## Compatibility Matrix

| Keptn Version    | [Dynatrace Service] |
|:----------------:|:----------------------------------------:|
|       0.6.1      | grabnerandi/dynatrace-service:0.1 |

## Installation

The *dynatrace-service* can either replace your existing installation or can be installed fresh in case you do not yet have a Dynatrace Service Installed.

### New Installation if dynatrace-service not yet existing

If you havent deployed the *dynatrace-service* yet into your Keptn installation you first need to create a secret that holds your Dynatrace Credentials. If you want to learn how to obtain tokens and tenant check out the Keptns doc: https://keptn.sh/docs/0.6.0/reference/monitoring/dynatrace/

Now we are ready to install this version of the dynatrace-service. We are going to use kubectl apply. Double check the version in the dynatrace-service.yaml to be the one you want to install from the list above!

```console
kubectl apply -f deploy/manifests/dynatrace-service/dynatrace-service.yaml
```

When the service is deployed, use the following command to let the `dynatrace-service` install Dynatrace OneAgent on your cluster. If Dynatrace OneAgents is already deployed, the current deployment of Dynatrace will not be modified.

```console
keptn configure monitoring dynatrace
```

### Replace the core dynatrace-service version with this one

To replace the existing jmeter-service with *jmeter-extened-service* simply replace the image in the jmeter-service deployment like this
```console
kubectl -n keptn set image deployment/dynatrace-service dynatrace-service=grabnerandi/dynatrace-service:0.1 --record
```

If you want to revert back to the core jmeter-service do this
```console
kubectl -n keptn set image deployment/dynatrace-service dynatrace-service=keptn/dynatrace-service:0.6.2 --record
```

## Usage Information

### Sending Events to Dynatrace Monitored Entities

The *dynatrace-service* by default assumes that all events it sends to Dynatrace, e.g: Deployment or Test Start/Stop Events are sent to a monitored Dynatrace SERVICE entity that has the following attachRule definition:
```
attachRules:
  tagRule:
    meTypes:
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

If your services are deployed with Keptn's Helm Service chances are that your services are automatically tagged like this. Here is a screenshot of how these tags show up in Dynatrace for a service deployed with Keptn:
![](./assets/keptn_tags_in_dynatrace.png)

If your services are however not tagged with these but other tags - or if you want the *dynatrace-service* to send the events not to a service but rather an application, process group or host then you can overwrite the default behavior by providing a *dynatrace/dynatrace.conf.yaml* file. This file can either be located on project, stage or service level. This file allows you to define your own attachRules and also allows you to leverage all available $PLACEHOLDERS such as $SERVICE,$STAGE,$PROJECT,$LABEL.YOURLABEL, ... - here is one example: It will instruct the *dynatrace-service* to send its events to a monitored Dynatrace Service that holds a tag with the key that matches your Keptn Service name ($SERICE) as well as holds an additional auto-tag that defines the enviornment to be pulled from a label that has been sent to Keptn.
```
---
spec_version: '0.1.0'
attachRules:
  tagRule:
    meTypes:
    - SERVICE
    tags:
    - context: CONTEXTLESS
      key: $SERVICE
    - context: CONTEXTLESS
      key: environment
      value: $LABEL.environment
```

### Enriching Events sent to Dynatrace with more context

The *dynatrace-service* sends CUSTOM_DEPLOYMENT, CUSTOM_INFO and CUSTOM_ANNOTATION events when it handles Keptn events such as deployment-finished, test-finished or evaluation-done. The *dynatrace-service* will parse all labels in the Keptn event and will pass them on to Dynatrace as custom properties. This gives you more flexiblity in passing more context to Dynatrace, e.g: ciBackLink for a CUSTOM_DEPLOYMENT or things like Jenkins Job ID, Jenkins Job URL ... that will show up in Dynatrace as well.
Here is a sample Deployment Finished Event:
```
{
  "type": "sh.keptn.events.deployment-finished",
  "contenttype": "application/json",
  "specversion": "0.2",
  "source": "jenkins",
  "id": "f2b878d3-03c0-4e8f-bc3f-454bc1b3d79d",
  "shkeptncontext": "08735340-6f9e-4b32-97ff-3b6c292bc509",
  "data": {
    "project": "simpleproject",
    "stage": "staging",
    "service": "simplenode",
    "testStrategy": "performance",
    "deploymentStrategy": "direct",
    "tag": "0.10.1",
    "image": "grabnerandi/simplenodeservice:1.0.0",
    "labels": {
      "testid": "12345",
      "buildnr": "build17",
      "runby": "grabnerandi",
      "environment" : "testenvironment",
      "ciBackLink" : "http://myjenkinsserver/job/12345"
    },
    "deploymentURILocal": "http://carts.sockshop-staging.svc.cluster.local",
    "deploymentURIPublic":  "https://carts.sockshop-staging.my-domain.com"
  }
}
```

It will result in the following events in Dynatrace:
![](./assets/deployevent.png)

### Sending Events to different Dynatrace Environments per Project, Stage or Service

Many Dynatrace user have different Dynatrace environments for e.g: Pre-Production vs Production. By default the *dynatrace-service* gets the Dynatrace Tenant URL & Token from the k8s secret stored in keptn/dynatrace (see installation instructions for details).
If you have multiple Dynatrace environment and want to have the *dynatrace-service* send events to a specific Dynatrace Environment for a specific Keptn Project, Stage or Service you can now specify the name of the secret that should be used in the *dynatrace.conf.yaml* which was introduced earlier. Here is a sample file:
```
---
spec_version: '0.1.0'
dtCreds: dynatrace-production
attachRules:
  tagRule:
    meTypes:
    - SERVICE
    tags:
    - context: CONTEXTLESS
      key: $SERVICE
    - context: CONTEXTLESS
      key: environment
      value: $LABEL.environment
```

The *dtCreds* value references your k8s secret where you store your Tenant and Token information. If you do not specify dtCreds it defaults to *dynatrace* which means it is the default behavior that we had for this service since the beginning!



# (STANDARD DOC) Dynatrace Service and Dynatrace OneAgent Operator
![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-contrib/dynatrace-service)
[![Build Status](https://travis-ci.org/keptn-contrib/dynatrace-service.svg?branch=master)](https://travis-ci.org/keptn-contrib/dynatrace-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-contrib/dynatrace-service)](https://goreportcard.com/report/github.com/keptn-contrib/dynatrace-service)

The *dynatrace-service* is a [Keptn](https://keptn.sh) service that sends information about the current state of a 
 pipeline run for a service to Dynatrace by sending events for the correlating detected service. 
 
The service is subscribed to the following Keptn CloudEvents:

- sh.keptn.events.deployment-finished
- sh.keptn.events.evaluation-done
- sh.keptn.events.tests-finished
- sh.keptn.internal.event.project.create
- sh.keptn.event.monitoring.configure

## Installation of Dynatrace Service and Dynatrace OneAgent Operator

1. Define your credentials by executing the following script:
    ```console
    kubectl -n keptn create secret generic dynatrace --from-literal="DT_API_TOKEN=<DT_API_TOKEN>" --from-literal="DT_TENANT=<DT_TENANT>" --from-literal="DT_PAAS_TOKEN=<DT_PAAS_TOKEN>"
    ```
    The $DT_TENANT has to be set according to the appropriate pattern:
    - Dynatrace SaaS tenant: `{your-environment-id}.live.dynatrace.com`
    - Dynatrace-managed tenant: `{your-domain}/e/{your-environment-id}`

1. Deploy the `dynatrace-service` using `kubectl apply`:

    ```console
    kubectl apply -f deploy/manifests/dynatrace-service/dynatrace-service.yaml
    ```
   
    When the service is deployed, use the following command to let the `dynatrace-service` install Dynatrace on your cluster. If Dynatrace is already deployed, the current deployment of Dynatrace will not be modified.

    ```console
    keptn configure monitoring dynatrace
    ```
   
 NOTE: If you're rolling out Dynatrace OneAgent to Container-Optimized OS(cos) based GKE clusters, you'll need to edit the `oneagent` Custom Resource in the `dynatrace` namespace and 
 add the following entry to the `env` section in the custom resource.
 
 First, edit the `OneAgent` Custom Resource:
  ```console
  kubectl edit oneagent -n dynatrace
  ```
 And then add this entry to the `env` section in the custom resource
 
  ```console
  env:
    - name: ONEAGENT_ENABLE_VOLUME_STORAGE
      value: "true"
  ```

  When the next event is sent to any of the keptn channels you see an event in Dynatrace for the correlating service:
![Dynatrace events](assets/events.png?raw=true "Dynatrace Events")

## Set up Dynatrace monitoring for already existing Keptn projects

If you already have created a project using Keptn and would like to enable Dynatrace monitoring for that project afterwards, please execute the following command:

    ```console
    keptn configure monitoring dynatrace --project=<PROJECT_NAME>
    ```

## Uninstall dynatrace-service

To uninstall the dynatrace service and remove the subscriptions to keptn channels execute this command.

```console
kubectl delete -f ./deploy/manifests/dynatrace-service/dynatrace-service.yaml
```


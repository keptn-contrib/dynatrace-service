# Dynatrace Service and Dynatrace OneAgent Operator
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
    kubectl apply -f https://raw.githubusercontent.com/keptn-contrib/dynatrace-service/master/deploy/manifests/dynatrace-service/dynatrace-service.yaml
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

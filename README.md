# Dynatrace Service and Dynatrace OneAgent Operator

The dynatrace-service sends information about the current state of a pipeline run for a service to Dynatrace by sending events for the correlating detected service.
The service is subscribed to the following keptn events:

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

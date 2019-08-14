# Dynatrace Service and Dynatrace OneAgent Operator

The dynatrace-service sends information about the current state of a pipeline run for a service to Dynatrace by sending events for the correlating detected service.
The service is subscribed to the following keptn events:

- sh.keptn.events.deployment-finished
- sh.keptn.events.evaluation-done
- sh.keptn.events.tests-finished

## Installation of Dynatrace Service and Dynatrace OneAgent Operator

1. Define your credentials by executing the following script:
    ```console
    cd ./deploy/scripts
    ./defineDynatraceCredentials.sh
    ```
    When the  script asks for your Dynatrace tenant, please enter your tenant according to the appropriate pattern:
    - Dynatrace SaaS tenant: `{your-environment-id}.live.dynatrace.com`
    - Dynatrace-managed tenant: `{your-domain}/e/{your-environment-id}`

1. Execute the installation script depending on your platform.
    - For AKS
    ```console
    ./deployDynatraceOnAKS.sh
    ```
    - For EKS
    ```console
    ./deployDynatraceOnEKS.sh
    ```    
    - For GKE
    ```console
    ./deployDynatraceOnGKE.sh
    ```    
    - For OpenShift
    ```console
    ./deployDynatraceOnGKE.sh
    ```
    When this script is finished, the Dynatrace OneAgent and the dynatrace-service are deployed in your cluster. Execute the following commands to verify the deployment of the dynatrace-service.

    ```console
    kubectl get ksvc dynatrace-service -n keptn
    ```

    ```console
    NAME                DOMAIN                                      LATESTCREATED             LATESTREADY               READY
    dynatrace-service   dynatrace-service.keptn.svc.cluster.local   dynatrace-service-26sm4   dynatrace-service-26sm4   True
    ```

    ```console
    kubectl get subscription -n keptn | grep dynatrace-subscription
    ```

    ```console
    dynatrace-subscription-deployment-finished          True
    dynatrace-subscription-evaluation-done              True
    dynatrace-subscription-tests-finished               True
    ```

  When the next event is sent to any of the keptn channels you see an event in Dynatrace for the correlating service:
![Dynatrace events](assets/events.png?raw=true "Dynatrace Events")

## Installation of dynatrace-service only

To use this service, you must have set up Dynatrace monitoring, as described in the [documentation](https://keptn.sh/docs/0.2.1/monitoring/dynatrace/).
Afterwards, apply the `dynatrace-service.yaml` using `kubectl` to create the dynatrace-service and the subscriptions to the keptn channels.

```console
kubectl apply -f ./deploy/manifests/dynatrace-service/dynatrace-service.yaml
```

```console
service.serving.knative.dev/dynatrace-service created
subscription.eventing.knative.dev/dynatrace-subscription-deployment-finished created
subscription.eventing.knative.dev/dynatrace-subscription-tests-finished created
subscription.eventing.knative.dev/dynatrace-subscription-evaluation-done created
```

## Uninstall of dynatrace-service only

To uninstall the dynatrace service and remove the subscriptions to keptn channels execute this command.

```console
kubectl delete -f ./deploy/manifests/dynatrace-service/dynatrace-service.yaml
```

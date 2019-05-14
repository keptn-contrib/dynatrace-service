# Dynatrace Service

This service sends information about the current state of a pipeline run for a service to Dynatrace by sending events for the correlating detected service.
The service is subscribed to the following keptn events:

- sh.keptn.events.deployment-finished
- sh.keptn.events.evaluation-done
- sh.keptn.events.tests-finished

## Installation

To use this service, you must have set up Dynatrace monitoring, as described in the [documentation](https://keptn.sh/docs/0.2.1/monitoring/dynatrace/).
Afterwards, to install the service in your keptn installation checkout or copy the `dynatrace-service.yaml`.

Then apply the `dynatrace-service.yaml` using `kubectl` to create the Dynatrace service and the subscriptions to the keptn channels.

```
kubectl apply -f dynatrace-service.yaml
```

Expected output:

```
service.serving.knative.dev/dynatrace-service created
subscription.eventing.knative.dev/dynatrace-subscription-new-artefact created
subscription.eventing.knative.dev/dynatrace-subscription-configuration-changed created
subscription.eventing.knative.dev/dynatrace-subscription-deployment-finished created
subscription.eventing.knative.dev/dynatrace-subscription-tests-finished created
subscription.eventing.knative.dev/dynatrace-subscription-evaluation-done created
```

## Verification of installation

```
$ kubectl get ksvc dynatrace-service -n keptn
NAME            DOMAIN                               LATESTCREATED         LATESTREADY           READY     REASON
dynatrace-service   dynatrace-service.keptn.x.x.x.x.xip.io   dynatrace-service-dd9km   dynatrace-service-dd9km   True
```

```
$ kubectl get subscription -n keptn | grep dynatrace-subscription
dynatrace-subscription-configuration-changed        True
dynatrace-subscription-deployment-finished          True
dynatrace-subscription-evaluation-done              True
dynatrace-subscription-keptn                        True
dynatrace-subscription-new-artefact                 True
dynatrace-subscription-tests-finished               True
$
```

When the next event is sent to any of the keptn channels you should see an event in Dynatrace for the correlating service:
![Dynatrace events](assets/events.png?raw=true "Dynatrace Events")

## Uninstall service

To uninstall the dynatrace service and remove the subscriptions to keptn channels execute this command.

```
kubectl delete -f dynatrace-service.yaml
```
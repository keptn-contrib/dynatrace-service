# Restore Keptn wildcard subscriptions

By default, the dynatrace-service subscribes to all `sh.keptn` events using the wildcard subscription `sh.keptn.*`. This allows the service to react to events as detailed in [Feature overview](feature-overview.md#keptn-events).

![default dynatrace-service subscription](images/subscriptions.png)

If you have changed this via Keptn Uniform and want to restore the default of `sh.keptn.*`, this can be achieved using the [Keptn API](https://keptn.sh/docs/0.13.x/reference/api/). In the future Keptn might introduce a way to restore the default via the Keptn Uniform UI.

## Get all subscriptions for the dynatrace-service

Go to Keptn API / controlPlane and query the `GET /uniform/registration` endpoint with the parameter `name=dynatrace-service`. This should return a payload like the one below:

```json
[
  {
    "id": "4b4d28a8f3bf811aa735f9531bae6bebf0df0f78",
    "name": "dynatrace-service",
    "metadata": {
      "hostname": "<some hostname>",
      "integrationversion": "0.20.0",
      "distributorversion": "0.9.0",
      "location": "control-plane",
      "kubernetesmetadata": {
        "namespace": "<some namespace>",
        "podname": "dynatrace-service-f645c8db5-bg4rr",
        "deploymentname": "dynatrace-service"
      },
      "lastseen": "2022-01-18T12:00:43.744Z"
    },
    "subscription": {
      "topics": null,
      "status": "",
      "filter": {
        "project": "",
        "stage": "",
        "service": ""
      }
    },
    "subscriptions": null
  }
]
```

Find the service with the correct version (if there is more than one) and copy the **integrationID** - in this case it would be `4b4d28a8f3bf811aa735f9531bae6bebf0df0f78` for version `0.20.0`.

Here, no subscriptions are available (`subscriptions: null`) because they were deleted previously.

### Optional: Query available subscriptions for an integration

In case you see at least one subscription, these can be queried via `GET /uniform/registration/{integrationID}/subscription` with the correct **integrationID** from the last step.

This will return any subscriptions set up via Keptn Uniform, such as:

```json
[
  {
    "id": "6ea2d86e-d6fa-4b52-a487-28cf4e4b2161",
    "event": "sh.keptn.event.get-sli.triggered",
    "filter": {
      "projects": [],
      "stages": [],
      "services": []
    }
  }
]
```

### Optional: Delete a subscription for an integration

If you have an invalid subscription, then delete it via `DELETE /uniform/registration/{integrationID}/subscription/{subscriptionID}` using the correct **integrationID** from above, as well as the correct **subscriptionID** from the last step.

## Add a new default subscription

If you want to restore the default subscription for the dynatrace-service, use `POST /uniform/registration/{integrationID}/subscription` together with the correct **integrationID** from above and payload below:

```json
{
  "filter": {
      "projects": [],
      "stages": [],
      "services": []
  },
  "event": "sh.keptn.>"
}
```

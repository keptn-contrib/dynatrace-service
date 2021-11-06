## Configurations of Dashboard SLI/SLO queries through `dynatrace.conf.yaml`

The `dynatrace.conf.yaml` provides the `dashboard` option to configure whether the *dynatrace-service* should use the metric queries defined in `sli.yaml`, whether it should pull data from a specific dashboard or whether it should query the data from a Dynatrace Dashboard whose name matches the Keptn project, stage and service. 

Here is an example `dynatrace.conf.yaml` including the `dashboard` parameter:

```yaml
---
spec_version: '0.1.0'
dtCreds: dynatrace-prod
dashboard: query
```

This file should be uploaded into the `dynatrace` subfolder, e.g. using the Keptn CLI:

```console
keptn add-resource --project=yourproject --stage=yourstage --resource=./dynatrace.conf.yaml --resourceUri=dynatrace/dynatrace.conf.yaml
```

The `dashboard` parameter provides 3 options:

* blank (default): If `dashboard` is not specified at all or if you do not even have a `dynatrace.conf.yaml` then the *dynatrace-service* will simply execute the metric query as defined in `slo.yaml`
* `query`: This value means that the *dynatrace-service* will look for a dashboard on your Dynatrace Tenant (dynatrace-prod in the example above) which has the following dashboard naming format: `KQG;project=<YOURKEPTNPROJECT>;service=<YOURKEPTNSERVICE>;stage=<YOURKEPTNSTAGE>`. If such a dashboard exists it will use the definition of that dashboard for SLIs as well as SLOs. If no dashboard was found an error will be returned.
* DASHBOARD-UUID: If you specify the UUID of a Dynatrace dashboard the *dynatrace-service* will query this dashboard on the specified Dynatrace Tenant. If it exists it will use the definition of this dashboard for SLIs as well as SLOs. If the dashboard was not found the *dynatrace-service* will return an error.

Here is an example of a `dynatrace.conf.yaml` specifying the UUID of a Dynatrace Dashboard:

```yaml
---
spec_version: '0.1.0'
dtCreds: dynatrace-prod
dashboard: 311f4aa7-5257-41d7-abd1-70420500e1c8
```

**Tip:** You can easily find the dashboard id for an existing dashboard by navigating to it in your Dynatrace Web interface. The ID is then part of the URL.
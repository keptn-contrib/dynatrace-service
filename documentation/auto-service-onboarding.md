# Automatic onboarding of monitored service entities

The dynatrace-service can automatically add service entities monitored by a single Dynatrace tenant to a Keptn project named `dynatrace`. To enable this feature, please see [Additional installation options](additional-installation-options.md#configuring-automatic-onboarding-of-services-monitored-by-dynatrace). 

To set up this feature, create a Keptn project named `dynatrace` containing at least a `quality-gate` stage and ensure access to the Dynatrace tenant is configured as described in [project setup](project-setup.md). For example, the following `shipyard.yaml` file could be used:

```yaml
apiVersion: "spec.keptn.sh/0.2.2"
kind: "Shipyard"
metadata:
  name: "dynatrace"
spec:
  stages:
    - name: "quality-gate"
      test_strategy: "performance"
```

By default, the dynatrace-service checks for new services every 60 seconds. It then adds any monitored services tagged with `keptn_managed` and `keptn_service:<service-name>` that are not already contained in the Keptn project. For each new service, the dynatrace-service adds the service to Keptn and creates the following default SLIs and SLOs:

SLIs (`dynatrace/sli.yaml` file):

```yaml
spec_version: '1.0'
indicators:
  throughput: "metricSelector=builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  error_rate: "metricSelector=builtin:service.errors.total.rate:merge(\"dt.entity.service\"):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p50: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p90: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(90)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p95: "metricSelector=builtin:service.response.time:merge(\"dt.entity.service\"):percentile(95)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
```

SLOs (`slo.yaml` file):

```yaml
spec_version: "1.0"
comparison:
  aggregate_function: "avg"
  compare_with: "single_result"
  include_result_with_score: "pass"
  number_of_comparison_results: 1
filter:
objectives:
  - sli: "response_time_p95"
    key_sli: false
    pass:             
      - criteria:
          - "<600"    
    warning:        
      - criteria:
          - "<=800"
    weight: 1
  - sli: "error_rate"
    key_sli: false
    pass:
      - criteria:
          - "<5"
  - sli: throughput
total_score:
  pass: "90%"
  warning: "75%"
```

After service onboarding, you should be able to see the newly created services within the Bridge:

![Services in Keptn Bridge](images/keptn_services_imported.png "Services in Keptn Bridge")


## Adding service entities to Keptn

After the project has been created, you can import service entities detected by Dynatrace by applying the tags `keptn_managed` and `keptn_service: <service_name>`:

![Keptn tags applied to a service](images/service_tags.png "Keptn tags applied to a service")

To set the `keptn_managed` tag, you can use the Dynatrace UI: First, in the **Transactions and services** menu, open the service entity you would like to tag, and add the `keptn_managed` tag as shown in the screenshot below:

![Adding a keptn_managed tag](images/keptn_managed_tag.png "Adding a keptn_managed tag")
 
The `keptn_service` tag can be set in two ways: 

1. Using an automated tagging rule, which can be set up in the menu **Settings > Tags > Automatically applied tags**. Within this section, add a new rule with the settings shown below:

    ![Adding an automated tagging rule](images/keptn_service_tag.png "Adding an automated tagging rule")

1. [A POST API call to the `v2/tags` endpoint](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/custom-tags/post-tags/):
    ```console
    curl -X POST "${DYNATRACE_TENANT}/api/v2/tags?entitySelector=${ENTITY_ID}" -H "accept: application/json; charset=utf-8" -H "Authorization: Api-Token ${API_TOKEN}" -H "Content-Type: application/json; charset=utf-8" -d "{\"tags\":[{\"key\":\"keptn_service\",\"value\":\"test\"}]}"
    ```


## Removing an onboarded service from Keptn

If you would like to remove an onboarded services from Keptn, remove the `keptn_managed` and `keptn_service` tags from the service entity in the Dynatrace tenant and then use the Keptn CLI to delete the service:

```console
keptn delete service <service-to-be-removed> --project=dynatrace

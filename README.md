# Dynatrace Service

![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-contrib/dynatrace-service)
[![Build Status](https://travis-ci.org/keptn-contrib/dynatrace-service.svg?branch=master)](https://travis-ci.org/keptn-contrib/dynatrace-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-contrib/dynatrace-service)](https://goreportcard.com/report/github.com/keptn-contrib/dynatrace-service)

The *dynatrace-service* is a [Keptn-service](https://keptn.sh) that forwards Keptn events - occurring during a delivery workflow - to Dynatrace. In addition, the service is responsible for configuring your Dynatrace tenant to fully interact with the Keptn installation.
 
The service is subscribed to the following [Keptn CloudEvents](https://github.com/keptn/spec/blob/master/cloudevents.md):

- sh.keptn.events.deployment-finished
- sh.keptn.events.evaluation-done
- sh.keptn.events.tests-finished
- sh.keptn.internal.event.project.create
- sh.keptn.event.monitoring.configure

## Compatibility Matrix

| Keptn Version    | [Dynatrace Service](https://hub.docker.com/r/keptncontrib/dynatrace-service/tags?page=1&ordering=last_updated) | Kubernetes Versions                      |
|:----------------:|:----------------------------------------:|:----------------------------------------:|
|       0.6.1      | keptn/dynatrace-service:0.6.2            | 1.13 - 1.15                              |
|       0.6.1      | keptncontrib/dynatrace-service:0.6.9     | 1.13 - 1.15                              |
|       0.6.2      | keptncontrib/dynatrace-service:0.7.1     | 1.13 - 1.15                              |
|       0.7.0      | keptncontrib/dynatrace-service:0.8.0     | 1.14 - 1.18                              |
|       0.7.1      | keptncontrib/dynatrace-service:0.9.0     | 1.14 - 1.18                              |
|       0.7.2      | keptncontrib/dynatrace-service:0.10.0     | 1.14 - 1.18                             |
|       0.7.3      | keptncontrib/dynatrace-service:0.10.1     | 1.14 - 1.18                            |
|       0.7.3      | keptncontrib/dynatrace-service:0.10.2     | 1.14 - 1.18                            |
|       0.7.3      | keptncontrib/dynatrace-service:0.10.3     | 1.14 - 1.18                            |
|       0.8.0, 0.8.1      | keptncontrib/dynatrace-service:0.11.0 (*)    | 1.14 - 1.19                            |
|       0.8.0, 0.8.1      | keptncontrib/dynatrace-service:0.12.0    | 1.14 - 1.19                            |
|       0.8.0 - 0.8.3     | keptncontrib/dynatrace-service:0.13.1    | 1.14 - 1.19                            |
|       0.8.0 - 0.8.3     | keptncontrib/dynatrace-service:0.14.0    | 1.14 - 1.19                            |
|       0.8.4             | keptncontrib/dynatrace-service:0.15.0    | 1.15 - 1.20                        |

(*) *Note:* 0.11.0 is feature-complete with 0.10.0. Changes and fixes made from 0.10.1 to 0.10.3 will be incorporated in 0.12.0

## Installation 


### 1. Gather Dynatrace and Keptn Credentials

To function correctly, the *dynatrace-service* requires access to a Dynatrace Tenant and to the Keptn API.

*  The credentials for the Dynatrace Tenant include `DT_API_TOKEN` and `DT_TENANT`: 

    * To create a Dynatrace API Token `DT_API_TOKEN`, log in to your Dynatrace tenant and go to **Settings > Integration > Dynatrace API**. Then, create a new API token with the following permissions:
      - Access problem and event feed, metrics, and topology
      - Read log content
      - Read configuration
      - Write configuration
      - Capture request data

    * The `DT_TENANT` has to be set according to the appropriate pattern:
      - Dynatrace SaaS tenant: `{your-environment-id}.live.dynatrace.com`
      - Dynatrace-managed tenant: `{your-domain}/e/{your-environment-id}` 

* The credentials for access to Keptn include `KEPTN_API_URL`, `KEPTN_API_TOKEN` and optionally `KEPTN_BRIDGE_URL`:

    * To determine the values for `KEPTN_API_URL` and `KEPTN_API_TOKEN` please refer to the [Keptn docs](https://keptn.sh/docs/0.8.x/operate/install/). 
   
    * If you would like to make use of the inclusion of backlinks to the Keptn Bridge, you `KEPTN_BRIDGE_URL` should also be provided. To find the URL of the bridge, please refer to the following section of the [Keptn docs](https://keptn.sh/docs/0.8.x/reference/bridge/#expose-lockdown-bridge). 

While setting up the service, it is recommended to gather these and set them as environment variables:

```console
DT_API_TOKEN=<DT_API_TOKEN>
DT_TENANT=<DT_TENANT>
KEPTN_API_URL=<KEPTN_API_URL>
KEPTN_API_TOKEN=<KEPTN_API_TOKEN>
KEPTN_BRIDGE_URL=<KEPTN_BRIDGE_URL> # optional
```

### 2. Create a Secret with Credentials

Create a secret (named `dynatrace` by default) containing the credentials for the Dynatrace Tenant (`DT_API_TOKEN` and `DT_TENANT`) and optionally for the Keptn API (`KEPTN_API_URL`, `KEPTN_API_TOKEN` and `KEPTN_BRIDGE_URL`).

```console
kubectl -n keptn create secret generic dynatrace \
--from-literal="DT_API_TOKEN=$DT_API_TOKEN" \
--from-literal="DT_TENANT=$DT_TENANT" \
--from-literal="KEPTN_API_URL=$KEPTN_API_URL" \
--from-literal="KEPTN_API_TOKEN=$KEPTN_API_TOKEN" \
--from-literal="KEPTN_BRIDGE_URL=$KEPTN_BRIDGE_URL" \
-oyaml --dry-run=client | kubectl replace -f -
```

 If the Keptn credentials are omitted from this main secret, `KEPTN_API_TOKEN` must be provided by the `keptn-api-token` secret. Furthermore, `dynatraceService.config.keptnApiUrl` and optionally `dynatraceService.config.keptnBridgeUrl` must be set when applying the helm chart (see below).

### 3. Deploy the Service

To deploy the current version of the *dynatrace-service* in your Kubernetes cluster, use the helm chart located in the `chart` directory.
Please use the same namespace for the *dynatrace-service* as you are using for Keptn, e.g: keptn.

```console
helm upgrade --install dynatrace-service -n keptn https://github.com/keptn-contrib/dynatrace-service/releases/download/$VERSION/dynatrace-service-$VERSION.tgz
```

The installation can then be verified using:

```console
kubectl -n keptn get deployment dynatrace-service -o wide
kubectl -n keptn get pods -l run=dynatrace-service
```

**Notes**: 
* Replace `$VERSION` with the desired version number (e.g. 0.15.0) you want to install.
* Variables may be set by appending key-value pairs with the syntax `--set key=value`
* If the `KEPTN_API_URL` and optionally `KEPTN_BRIDGE_URL` were not provided via a secret (see above) they should be provided using the variables `dynatraceService.config.keptnApiUrl` and `dynatraceService.config.keptnBridgeUrl`, i.e. by appending `--set dynatraceService.config.keptnApiUrl=$KEPTN_API_URL --set dynatraceService.config.keptnBridgeUrl=$KEPTN_BRIDGE_URL`.
* The `dynatrace-service` can automatically generate tagging rules, problem notifications, management zones, dashboards, and custom metric events in your Dynatrace tenant. You can configure whether these entities should be generated within your Dynatrace tenant by the environment variables specified in the provided [values.yml](https://raw.githubusercontent.com/keptn-contrib/dynatrace-service/$VERSION/chart/values.yaml),
 i.e. using the variables `dynatraceService.config.generateTaggingRules` (default `false`), `dynatraceService.config.generateProblemNotifications` (default `false`), `dynatraceService.config.generateManagementZones` (default `false`), `dynatraceService.config.generateDashboards` (default `false`), `dynatraceService.config.generateMetricEvents` (default `false`), and `dynatraceService.config.synchronizeDynatraceServices` (default `true`).
 
* The `dynatrace-service` by default validates the SSL certificate of the Dynatrace API. If your Dynatrace API only has a self-signed certificate, you can disable the SSL certificate check by setting the environment variable `dynatraceService.config.httpSSLVerify` (default `true`) specified in the [values.yml](https://raw.githubusercontent.com/keptn-contrib/dynatrace-service/$VERSION/chart/values.yaml) to `false`.

* The `dynatrace-service` can be configured to use a proxy server via the `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` environment variables  as described in [`httpproxy.FromEnvironment()`](https://golang.org/pkg/vendor/golang.org/x/net/http/httpproxy/#FromEnvironment). As the `dynatrace-service` connects to a `distributor`, a `NO_PROXY` entry including `127.0.0.1` should be used to prevent these from being proxied. The `HTTP_PROXY` and `HTTPS_PROXY` environment variables can be configured using the `dynatraceService.config.httpProxy` (default `""`) and `dynatraceService.config.httpsProxy` (default `""`) in [values.yml](https://raw.githubusercontent.com/keptn-contrib/dynatrace-service/$VERSION/chart/values.yaml), `NO_PROXY` is set to `127.0.0.1` by default. For example:


```console
helm upgrade --install dynatrace-service -n keptn https://github.com/keptn-contrib/dynatrace-service/releases/download/$VERSION/dynatrace-service.tgz --set dynatraceService.config.httpProxy=http://mylocalproxy:1234 --set dynatraceService.config.httpsProxy=https://mylocalproxy:1234
```

* When an event is sent out by Keptn, you see an event in Dynatrace for the correlating service:

![Dynatrace events](assets/events.png?raw=true "Dynatrace Events")


### Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by the `$VERSION` placeholder):

```console
helm upgrade dynatrace-service -n keptn https://github.com/keptn-contrib/dynatrace-service/releases/download/$VERSION/dynatrace-service-$VERSION.tgz
```

### Uninstall

To delete a deployed *dynatrace-service*, use the `helm` CLI to uninstall the installed release of the service:

```console
helm delete -n keptn dynatrace-service
```

## Set up Dynatrace monitoring for already existing Keptn projects

If you already have created a project using Keptn and would like to enable Dynatrace monitoring for that project afterwards, please execute the following command:

```console
keptn configure monitoring dynatrace --project=<PROJECT_NAME>
```

**ATTENTION:** If you have different Dynatrace Tenants (or Managed Environments) and want to make sure a Keptn project is linked to the correct Dynatrace Tenant/Environment please have a look at the dynatrace.conf.yaml file option as explained further down in this readme. It allows you on a project level to specify which Dynatrace Tenant/Environment to use. Whats needed is that you first upload dynatrace.conf.yaml on project level before calling keptn configure monitoring!

## Usage Information

### Sending Events to Dynatrace Monitored Entities

By default, the *dynatrace-service* assumes that all events it sends to Dynatrace, e.g: Deployment or Test Start/Stop Events are sent to a monitored Dynatrace SERVICE entity that has the following attachRule definition:
```
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

If your services are deployed with Keptn's *helm-service*, chances are that your services are automatically tagged like this. Here is a screenshot of how these tags show up in Dynatrace for a service deployed with Keptn:

![](./assets/keptn_tags_in_dynatrace.png)

If your services are however not tagged with these but other tags - or if you want the *dynatrace-service* to send the events not to a service but rather an application, process group or host then you can overwrite the default behavior by providing a *dynatrace/dynatrace.conf.yaml* file. This file can either be located on project, stage or service level. This file allows you to define your own attachRules and also allows you to leverage all available $PLACEHOLDERS such as $SERVICE,$STAGE,$PROJECT,$LABEL.YOURLABEL, etc. - here is one example: It will instruct the *dynatrace-service* to send its events to a monitored Dynatrace Service that holds a tag with the key that matches your Keptn Service name ($SERVICE) as well as holds an additional auto-tag that defines the enviornment to be pulled from a label that has been sent to Keptn.
```
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

Now - once you have this file - make sure you add it as a resource to your Keptn Project. As mentioned above - the dynatrace/dynatrace.conf.yaml can be uploaded either on project, service or stage level. Here is an example on how to define it for the whole project!
```
keptn add-resource --project=yourproject --resource=dynatrace/dynatrace.conf.yaml --resourceUri=dynatrace/dynatrace.conf.yaml
```

### Enriching Events sent to Dynatrace with more context

The *dynatrace-service* sends CUSTOM_DEPLOYMENT, CUSTOM_INFO and CUSTOM_ANNOTATION events when it handles Keptn events such as deployment-finished, test-finished or evaluation-done. The *dynatrace-service* will parse all labels in the Keptn event and will pass them on to Dynatrace as custom properties. This gives you more flexiblity in passing more context to Dynatrace, e.g: ciBackLink for a CUSTOM_DEPLOYMENT or things like Jenkins Job ID, Jenkins Job URL, etc. that will show up in Dynatrace as well. 

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
  - meTypes:
    - SERVICE
    tags:
    - context: CONTEXTLESS
      key: $SERVICE
    - context: CONTEXTLESS
      key: environment
      value: $LABEL.environment
```

The *dtCreds* value references your k8s secret where you store your Tenant and Token information. If you do not specify dtCreds it defaults to *dynatrace* which means it is the default behavior that we had for this service since the beginning!

As a reminder - here is the way how to upload this to your Keptn Configuration Repository. In case you have two separate dynatrace.conf.yaml for your different dynatrace tenants you can even upload them to your different stages in your Keptn project in case your different stages are monitored by different dynatrace enviornments. Here are some examples on how to upload these files:
```
keptn add-resource --project=yourproject --stage=preprod --resource=dynatrace/dynatrace-preprod.conf.yaml --resourceUri=dynatrace/dynatrace.conf.yaml
keptn add-resource --project=yourproject --stage=production --resource=dynatrace/dynatrace-production.conf.yaml --resourceUri=dynatrace/dynatrace.conf.yaml
```


### Synchronizing Service Entities detected by Dynatrace

The Dynatrace service allows Service Entities detected by Dynatrace to be automatically imported into Keptn. To enable this feature, the environment variable `SYNCHRONIZE_DYNATRACE_SERVICES`
needs to be set to `true`. Once enabled, the service will by default scan Dynatrace for Service Entities every 60 seconds. This interval can be configured by changing the environment variable `SYNCHRONIZE_DYNATRACE_SERVICES_INTERVAL_SECONDS`.

To import a Service Entity into Keptn, a project with the name `dynatrace`, containing the stage `quality-gate` has to be available within Keptn. To create the project, create a `shipyard.yaml` file with the following content:

```
stages:
  - name: "quality-gate"
    test_strategy: "performance"
```

Afterwards, create the project using the following command:

```
keptn create project dynatrace --shipyard=shipyard.yaml
```

After the project has been created, you can import Service Entities detected by Dynatrace by applying the tags `keptn_managed` and `keptn_service: <service_name>`:

![](./assets/service_tags.png)

To set the `keptn_managed` tag, you can use the Dynatrace UI: First, in the **Transactions and services** menu, open the Service Entity you would like to tag, and add the `keptn_managed` tag as shown in the screenshot below:

![](./assets/keptn_managed_tag.png)
 
The `keptn_service` tag can be set in two ways. 

1. Using an automated tagging rule, which can be set up in the menu **Settings > Tags > Automatically applied tags**. Within this section, add a new rule with the settings shown below:
    ![](./assets/keptn_service_tag.png)

1. Sending a POST API call to the `v2/tags` endpoint; [see here](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/custom-tags/post-tags/)
    ```console
    curl -X POST "${DYNATRACE_TENANT}/api/v2/tags?entitySelector=${ENTITY_ID}" -H "accept: application/json; charset=utf-8" -H "Authorization: Api-Token ${API_TOKEN}" -H "Content-Type: application/json; charset=utf-8" -d "{\"tags\":[{\"key\":\"keptn_service\",\"value\":\"test\"}]}"
    ```

The Dynatrace Service will then periodically check for services containing those tags and create correlating services within the `dynatrace` project in Keptn. After the service synchronization, you should be able to see the newly created services within the Bridge:

![](./assets/keptn_services_imported.png)

Note that if you would like to remove one of the imported services from Keptn, you will need to use the Keptn CLI to delete the service after removing the `keptn_managed` and `keptn_service` tags:

```
keptn delete service <service-to-be-removed> --project=dynatrace
```

In addition to creating the service, the dynatrace-service will also upload the following default `slo.yaml` to enable the quality-gates feature for the service:

```
---
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

To enable queries against the SLIs specified in the SLO.yaml file, the following configuration is created for the SLI configuration for the `dynatrace-sli-service`:

```
---
spec_version: '1.0'
indicators:
  throughput: "metricSelector=builtin:service.requestCount.total:merge(0):sum&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  error_rate: "metricSelector=builtin:service.errors.total.rate:merge(0):avg&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p50: "metricSelector=builtin:service.response.time:merge(0):percentile(50)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p90: "metricSelector=builtin:service.response.time:merge(0):percentile(90)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"
  response_time_p95: "metricSelector=builtin:service.response.time:merge(0):percentile(95)&entitySelector=type(SERVICE),tag(keptn_managed),tag(keptn_service:$SERVICE)"`
```

This file will be stored in the `dynatrace/sli.yaml` config file for the created service. See the [dynatrace-sli-service docs](https://github.com/keptn-contrib/dynatrace-sli-service/tree/update/test-coverage-and-doc#overwrite-sli-configuration--custom-sli-queries) for a detailed description of how this file is used 
to configure the retrieval af metrics for a service 

### Sending Dynatrace Problems to Keptn for Auto-Remediation

One major use case of Keptn is Auto-Remediation. This is where Keptn receives a problem event which then triggers a remediation workflow.
External tools such as Dynatrace can send a `sh.keptn.events.problem` event to Keptn but first need to be mapped to a Keptn Project, Service and Stage. Depending on the alerting tool this might be done differently. 

The *dynatrace-service* provides the capabilty to receive such a `sh.keptn.events.problem` - analyzes its content and sends a `sh.keptn.event.problem.open` to the matching keptn project, service and stage including all relevent problem details such as PID, ProblemTitle, Problem URL, ...

**Setting Up Problem Notification for Problems detected on Keptn Deployed Services**

If you use Keptn to deploy your microservices and follow our tagging practices Dynatrace will tag your monitored services with keptn_project, keptn_service and keptn_stage. If Dynatrace then detects a problem in one of these deployed services, e.g: High Failure Rate, Slow response time, ... you can let Dynatrace send these problems back to Keptn and map the problem directly to the correct Kept Project, Stage and Service.

To setup this integration you just need to setup a Custo Problem Notification that looks like this: 
```json
{
    "specversion":"1.0",
    "shkeptncontext":"{PID}",
    "type":"sh.keptn.events.problem",
    "source":"dynatrace",
    "id":"{PID}",
    "time":"",
    "contenttype":"application/json",
    "data": {
        "State":"{State}",
        "ProblemID":"{ProblemID}",
        "PID":"{PID}",
        "ProblemTitle":"{ProblemTitle}",
        "ProblemURL":"{ProblemURL}",
        "ProblemDetails":{ProblemDetailsJSON},
        "Tags":"{Tags}",
        "ImpactedEntities":{ImpactedEntities},
        "ImpactedEntity":"{ImpactedEntity}",
    }
}
```

The *dynatrace-service* will parse the "Tags" field and tries to find keptn_project, keptn_service and keptn_stage tags that come directly from the impacted entities that Dynatrace detected. If the problem was in fact detected on a Keptn deployed service the `{Tags}` string should contain the correct information and the mapping will work.

*Best practice:* if you setup this type of integration we suggest that you use a Dynatrace Alerting Profile that only includes problems on services that have the Keptn tags. Otherwise problems will be sent to Keptn that cant be mapped through this capability!


**Setting Up Problem Notification for ANY type of detected problem, e.g: Infrastructure, ...**

So - what if you want to send any type of problem for a specific Alerting Profile to Keptn and use Keptn to orchestrate auto-remediation workflows? In that case we allow you specify Keptn Project, Stage and Service as properties in the data structure that is sent to Keptn.

Here the custom payload for a Custom Notification Integration that will send all problems to a Keptn project called `dynatrace`, stage called `production` and service called `allproblems`:
```json
{
    "specversion":"1.0",
    "shkeptncontext":"{PID}",
    "type":"sh.keptn.events.problem",
    "source":"dynatrace",
    "id":"{PID}",
    "time":"",
    "contenttype":"application/json",
    "data": {
        "State":"{State}",
        "ProblemID":"{ProblemID}",
        "PID":"{PID}",
        "ProblemTitle":"{ProblemTitle}",
        "ProblemURL":"{ProblemURL}",
        "ProblemDetails":{ProblemDetailsJSON},
        "Tags":"{Tags}",
        "ImpactedEntities":{ImpactedEntities},
        "ImpactedEntity":"{ImpactedEntity}",
        "KeptnProject" : "demo-remediation",
        "KeptnService" : "allproblem",
        "KeptnStage" : "production"
    }
}
``` 

When the *dynatrace-service* receives this `sh.keptn.events.problem` it will parse the fields KeptnProject, KeptnService and KeptnStage and will then send a `sh.keptn.event.problem.open` to Keptn including the rest of the problem details!
This allows you to send any type of Dynatrace detected problem to Keptn and let Keptn execute a remediation workflow

*Best Practice:* We suggest that you use Dynatrace Alerting Profiles to filter on certain problem types, e.g: Infrastructure problems in production, Slow Performance in Developer Environment ...  We then also suggest that you create a Keptn project on Dynatrace to handle these remediation workflows and create a Keptn Service for each alerting profile. With this you have a clear match of Problems per Alerting Profile and a Keptn Remediation Workflow that will be executed as it matches your Keptn Project and Service. For stage I suggest you also go with the environment names you have, e.g: Pre-Prod or Production.

Here is a screenshot of a workflow triggered by a Dynatrace problem and how it then executes in Keptn:

![](./assets/remediation_workflow.png)

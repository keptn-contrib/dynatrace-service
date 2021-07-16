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
* Replace `$VERSION` with the desired version number (e.g. 0.15.1) you want to install.
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

![Dynatrace events](images/events.png?raw=true "Dynatrace Events")


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

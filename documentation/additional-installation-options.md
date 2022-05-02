# Additional installation options

This section describes additional installation options for the dynatrace-service. These can be specified by appending key-value pairs (syntax `--set key=value`) to `helm upgrade` commands. All default values are listed in [`chart/values.yaml`](https://github.com/keptn-contrib/dynatrace-service/blob/master/chart/values.yaml), while the chart's [README.md](https://github.com/keptn-contrib/dynatrace-service/blob/master/chart/README.md) provides a summarized list.


## Configuring automatic onboarding of services monitored by Dynatrace

The dynatrace-service can periodically check for new monitored services and add these to Keptn. This feature may be customized using the following Helm chart values:

| Value name | Description | Default |
|---|---|---|
| `dynatraceService.config.synchronizeDynatraceServices` | Automatically add newly detected service entities to Keptn | `true` |
| `dynatraceService.config.synchronizeDynatraceServicesIntervalSeconds` | Interval between checks | `60` |

Further details are provided in [Automatic onboarding of monitored service entities](auto-service-onboarding.md)


## Configuring automatic Dynatrace tenant configuration

The dynatrace-service can automatically generate basic tagging rules, problem notifications, management zones, dashboards, and custom metric events in the Dynatrace tenant associated with a Keptn project. These may be enabled using the following Helm chart values:

| Value name | Description | Default |
|---|---|---|
| `dynatraceService.config.generateTaggingRules` | Generate standard tagging rules in the Dynatrace tenant | `false` |
| `dynatraceService.config.generateProblemNotifications` | Generate a standard problem notification configuration in the Dynatrace tenant | `false` |
| `dynatraceService.config.generateManagementZones` | Generate standard management zones in the Dynatrace tenant | `false` |
| `dynatraceService.config.generateDashboards` | Generate a standard dashboard in the Dynatrace tenant | `false` |
| `dynatraceService.config.generateMetricEvents` | Generate standard metric events in Dynatrace tenant | `false` |

The actual configuration is carried out in response to a `sh.keptn.event.monitoring.configure` event. Further details are provided in [Automatic configuration of a Dynatrace tenant](auto-tenant-configuration.md).


## Configuring Dynatrace tenant API SSL certificate validation

By default, the dynatrace-service validates the SSL certificate of the Dynatrace tenant's API. If the Dynatrace API only has a self-signed certificate, you can disable the SSL certificate check by setting the Helm chart value `dynatraceService.config.httpSSLVerify` to `false`.

| Value name | Description | Default |
|---|---|---|
| `dynatraceService.config.httpSSLVerify` | Verify Dynatrace tenant's API HTTPS SSL certificates | `true` |


## Configuring the dynatrace-service to use a proxy

In certain instances where the dynatrace-service is installed behind a firewall, it may need to use a proxy to access a Dynatrace tenant. This can be configured using the `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` environment variables as described in [`httpproxy.FromEnvironment()`](https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment). The environment variables are exposed through the `dynatraceService.config.httpProxy`, `dynatraceService.config.httpsProxy` and `dynatraceService.config.noProxy` Helm values.

Due to the large variety of configurations, the dynatrace-service no longer provides defaults for `dynatraceService.config.noProxy`. In general, entries should be added to prevent requests to other Keptn services as well as Kubernetes services operating within the cluster. For example, this may be done by setting `dynatraceService.config.noProxy` to `"127.0.0.1,mongodb-datastore,configuration-service,shipyard-controller,kubernetes.default.svc.cluster.local"`, however the exact values depend on the specific setup.

| Value name | Description | Default |
|---|---|---|
| `dynatraceService.config.httpProxy` | Proxy for HTTP requests | `""` |
| `dynatraceService.config.httpsProxy` | Proxy for HTTPS requests | `""` |
| `dynatraceService.config.noProxy` | Proxy exceptions for HTTP and HTTPS requests | `""` |


## Setting the log output level

The minimum log level of messages emitted by the service may be set via `dynatraceService.config.logLevel`. The following levels are supported: `panic`, `fatal`, `error`,`warn` (or `warning`), `info`, `debug` and `trace`. By default the minimum level is set to `info`, meaning that info, warning, error, fatal and panic messages are emitted.

| Value name | Description | Default |
|---|---|---|
| `dynatraceService.config.logLevel`| Minimum log level to log | `info` |


## Setting the termination grace period

The termination grace period of the pod may be set via `terminationGracePeriodSeconds`.

| Value name | Description | Default |
|---|---|---|
| `terminationGracePeriodSeconds` | Termination grace period (in seconds) | `30` |
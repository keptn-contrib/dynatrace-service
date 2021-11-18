
Dynatrace-service
===========

Helm Chart for the *keptn-contrib* *dynatrace-service*


## Configuration

The following table lists the configurable parameters of the *dynatrace-service* chart and their default values.

| Parameter                | Description             | Default        |
| ------------------------ | ----------------------- | -------------- |
| `dynatraceService.image.repository` | Container image name | `"docker.io/keptncontrib/dynatrace-service"` |
| `dynatraceService.image.pullPolicy` | Kubernetes image pull policy | `"IfNotPresent"` |
| `dynatraceService.image.tag` | Container tag | `""` |
| `dynatraceService.service.enabled` | Creates a kubernetes service for the *dynatrace-service* | `true` |
| `dynatraceService.config.generateTaggingRules` | Generate Tagging Rules in Dynatrace Tenant | `false` |
| `dynatraceService.config.generateProblemNotifications` | Generate Problem Notifications in Dynatrace Tenant | `false` |
| `dynatraceService.config.generateManagementZones` | Generate Management Zones in Dynatrace Tenant | `false` |
| `dynatraceService.config.generateDashboards` | Generate Dashboards in Dynatrace Tenant | `false` |
| `dynatraceService.config.generateMetricEvents` | Generate Metric Events in Dynatrace Tenant | `false` |
| `dynatraceService.config.synchronizeDynatraceServices` | Synchronize Service Entities between Dynatrace and Keptn | `true` |
| `dynatraceService.config.synchronizeDynatraceServicesIntervalSeconds` | Synchronization Interval | `300` |
| `dynatraceService.config.httpSSLVerify` | Verify HTTPS SSL certificates | `true` |
| `dynatraceService.config.httpProxy` | Proxy for HTTP requests | `""` |
| `dynatraceService.config.httpsProxy` | Proxy for HTTPS requests | `""` |
| `dynatraceService.config.noProxy` | Proxy exceptions for HTTP and HTTPS requests | `""` |
| `dynatraceService.config.logLevel`| Minimum log level to log | `info` |
| `distributor.stageFilter` | Sets the stage this *dynatrace-service* belongs to | `""` |
| `distributor.serviceFilter` | Sets the service this *dynatrace-service* belongs to | `""` |
| `distributor.projectFilter` | Sets the project this *dynatrace-service* belongs to | `""` |
| `distributor.image.repository` | Container image name | `"docker.io/keptn/distributor"` |
| `distributor.image.pullPolicy` | Kubernetes image pull policy | `"IfNotPresent"` |
| `distributor.image.tag` | Container tag | `""` |
| `remoteControlPlane.enabled` | Enables remote execution plane mode | `false` |
| `remoteControlPlane.api.protocol` | Used protocol (http, https | `"https"` |
| `remoteControlPlane.api.hostname` | Hostname of the control plane cluster (and port) | `""` |
| `remoteControlPlane.api.apiValidateTls` | Defines if the control plane certificate should be validated | `true` |
| `remoteControlPlane.api.token` | Keptn api token | `""` |
| `imagePullSecrets` | Secrets to use for container registry credentials | `[]` |
| `serviceAccount.create` | Enables the service account creation | `true` |
| `serviceAccount.annotations` | Annotations to add to the service account | `{}` |
| `podAnnotations` | Annotations to add to the created pods | `{}` |
| `podSecurityContext` | Set the pod security context (e.g. `fsgroups`) | `{}` |
| `securityContext` | Set the security context (e.g. `runasuser`) | `{}` |
| `resources` | Resource limits and requests | `{}` |
| `nodeSelector` | Node selector configuration | `{}` |
| `tolerations` | Tolerations for the pods | `[]` |
| `affinity` | Affinity rules | `{}` |






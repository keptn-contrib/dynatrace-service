# Dynatrace API token scopes

To interact with a Dynatrace tenant, the dynatrace-service requires access to the Dynatrace API. Different scopes are required depending on the features you would like to use.

## Scopes required for features

|Feature | Required scope(s)|
|:--------|:-----------------|
| [SLIs via `dynatrace/sli.yaml` files](slis-via-files.md) | - |
| [SLIs via a Dynatrace dashboard](slis-via-dashboard.md) | Read configuration (`ReadConfig`)|
| [Forwarding events from Keptn to Dynatrace](event-forwarding-to-dynatrace.md) | Access problem and event feed, metrics, and topology (`DataExport`) |
| [Forwarding problem notifications from Dynatrace to Keptn](problem-forwarding-to-keptn.md) | - |
| [Automatic onboarding of monitored service entities](auto-service-onboarding.md) | Read entities (`entities.read`) |
| [Automatic configuration of a Dynatrace tenant](auto-tenant-configuration.md) | Read configuration (`ReadConfig`), Write configuration (`WriteConfig`) |

## Scopes required for SLIs

When functioning as an SLI provider for Keptn, additional scopes are required depending on the type of SLI:

|SLI type| Required scope(s) |
|:--|:--|
| Metrics | Read metrics (`metrics.read`) |
| SLOs (`SLO`) | Read SLO (`slo.read`) |
| Problems (`PV2`) | Read problems (`problems.read`) |
| Security problems (`SECPV2`) | Read security problems (`securityProblems.read`) |
| User sessions (`USQL`) | User sessions (`DTAQLAccess`) |
| Converted metrics (`MV2`) | Read metrics (`metrics.read`) |
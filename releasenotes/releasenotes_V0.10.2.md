# Release Notes 0.10.2

This release enables better Auto Remediation integrations as the tags from a Dynatrace problem are also added to the Keptn Problem Event which is then accessible by any remediation action handler, e.g: its now possible to use these tags when creating a ticket in Jira, OpsGenie, VictorOps, ServiceNow or other incident management systems

This release also includes a fix when creating Dynatrace Custom Alerts as well as improved logging of problem event handling

## New Features

- Dynatrace Problem Event Tags added to Keptn Problem Event [#221](https://github.com/keptn-contrib/dynatrace-service/issues/221)

## Fixes

- Use correct default Dynatrace service error rate metric when setting up alerts [#229](https://github.com/keptn-contrib/dynatrace-service/issues/229)
- Improved logging when dynatrace-service processes incoming problem events from Dynatrace [#223](https://github.com/keptn-contrib/dynatrace-service/issues/223)

# Release Notes 0.6.0

## New Features
- Use Dynatrace Service to set up Dynatrace OneAgent, Management Zones, Dashboards, Auto Tagging Rules and Problem Notifications [#443](https://github.com/keptn/keptn/issues/443)
- Automatically create Keptn-specific alerting profile in Dynatrace [#1281](https://github.com/keptn/keptn/issues/1281)
- Receive Dynatrace problem notifications and send convert them to a Keptn-internal `sh.keptn.event.problem.open` event () [#1185](https://github.com/keptn/keptn/issues/1185)

## Fixed Issues

## Known Limitations
- When using Container-Optimized OS (COS) based GKE clusters, the deployed OneAgent has to be updated after the installation of Dynatrace

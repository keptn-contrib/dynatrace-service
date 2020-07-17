# Release Notes 0.8.0

This release addresses RBAC issues of the dynatrace-service on a Kubernetes cluster and is adapted to work with Keptn 0.7.

**Attention:** In this release, the Dynatrace OneAgent Operator will not be installed automatically.

## New Features

- Send problem comments when actions are triggered/finished [#142](https://github.com/keptn-contrib/dynatrace-service/pull/142)
- Added problemUrl to problem payload [#144](https://github.com/keptn-contrib/dynatrace-service/pull/144)
- Created and now uses RBAC objects [#145](https://github.com/keptn-contrib/dynatrace-service/pull/145)
- Removed dependency from keptn-domain ConfigMap [#148](https://github.com/keptn-contrib/dynatrace-service/pull/148)
- Removed the installation of Dynatrace OneAgent when configuring monitoring [#156](https://github.com/keptn-contrib/dynatrace-service/pull/156)

## Fixed Issues

- Correct WebSocket URL to /websocket [#152](https://github.com/keptn-contrib/dynatrace-service/pull/152)

## Known Limitations

- For old limitations, please see [Release 0.7.0](https://github.com/keptn-contrib/dynatrace-service/releases/tag/0.7.0).
# Release Notes 0.18.1

## New Features

- Support DIMENSION filter type in Data Explorer tiles [#577](https://github.com/keptn-contrib/dynatrace-service/issues/577)

## Fixed Issues

- Management zone is not correctly applied to SLIs generated from Data Explorer tiles if they have no entity selector [#599](https://github.com/keptn-contrib/dynatrace-service/issues/599)
- Requests to in-cluster Kubernetes services should not be proxied by default [#555](https://github.com/keptn-contrib/dynatrace-service/issues/555)
- Dynatrace-service serviceAccount value defaults to release-name [#587](https://github.com/keptn-contrib/dynatrace-service/issues/587)
- Errors accessing individual dynatrace secrets are not reported by DynatraceCredentialsProviderFallbackDecorator [#583](https://github.com/keptn-contrib/dynatrace-service/issues/583)
- Metrics selectors generated from Custom Charting tiles should apply splitBy after applying filter [#581](https://github.com/keptn-contrib/dynatrace-service/issues/581)
- Metrics selectors generated from Data Explorer tiles should apply splitBy after applying filter [#579](https://github.com/keptn-contrib/dynatrace-service/issues/579)

## Other Changes

- Remove all proxy configuration defaults, improve proxy configuration documentation [#597](https://github.com/keptn-contrib/dynatrace-service/issues/597)
- Remove fallback if a specified Dynatrace credentials secret cannot be found [#442](https://github.com/keptn-contrib/dynatrace-service/issues/442)
- Metrics queries generated from Data Explorer and Custom Charting tiles should use splitBy rather than merge to implement split by functionality [#578](https://github.com/keptn-contrib/dynatrace-service/issues/578)

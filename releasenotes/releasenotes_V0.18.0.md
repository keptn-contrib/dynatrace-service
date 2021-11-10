# Release Notes 0.18.0

## New Features

-  Support filters on SERVICE_KEY_REQUEST in Custom Charting tiles [#565](https://github.com/keptn-contrib/dynatrace-service/issues/565)

## Fixed issues

-  Plus sign gets removed from SLI [#537](https://github.com/keptn-contrib/dynatrace-service/issues/537)
-  Data Explorer tile processing ignores filterType [#564](https://github.com/keptn-contrib/dynatrace-service/issues/564)
-  Data Explorer tile processing ignores spaceAggregation [#563](https://github.com/keptn-contrib/dynatrace-service/issues/563)
-  Custom Charting tile processing ignores series entityType and selects filter based on metric definition entityType [#566](https://github.com/keptn-contrib/dynatrace-service/issues/566) 
-  Dashboard generated SLI not including filter [#369](https://github.com/keptn-contrib/dynatrace-service/issues/369)
-  Dashboard processing without results will not return an error [#553](https://github.com/keptn-contrib/dynatrace-service/issues/553)
-  No default SLO definitions [#551](https://github.com/keptn-contrib/dynatrace-service/issues/551)

## Other Changes

-  Improve parsing and validation of custom SLI definitions [#571](https://github.com/keptn-contrib/dynatrace-service/issues/571)
-  Security Hardening: Remove the role which allows to get, list, watch all secrets [#485](https://github.com/keptn-contrib/dynatrace-service/issues/485)
-  Add securityContext, resource limits and requests [#568](https://github.com/keptn-contrib/dynatrace-service/pull/568)
-  Bump k8s.io/client-go from 0.22.2 to 0.22.3 [#559](https://github.com/keptn-contrib/dynatrace-service/pull/559)
-  Refactor KeptnCredentials and their usage [#543](https://github.com/keptn-contrib/dynatrace-service/issues/543)
-  Remove Dashboard caching in Keptn resources [#535](https://github.com/keptn-contrib/dynatrace-service/issues/535) 
-  Refactor and improve handling of Dynatrace and Keptn credentials [#540](https://github.com/keptn-contrib/dynatrace-service/issues/540)
-  Bump github.com/cloudevents/sdk-go/v2 from 2.6.0 to 2.6.1 [#548](https://github.com/keptn-contrib/dynatrace-service/pull/548)
-  Bump github.com/cloudevents/sdk-go/v2 from 2.5.0 to 2.6.0 [#546](https://github.com/keptn-contrib/dynatrace-service/pull/546)
-  Send errors to Keptn Uniform if dynatrace secret could not be found [#533](https://github.com/keptn-contrib/dynatrace-service/issues/533)
-  Remove fallbacks in dashboard processing [#531](https://github.com/keptn-contrib/dynatrace-service/pull/531) 
-  Remove fallbacks if dashboard is not found [#433](https://github.com/keptn-contrib/dynatrace-service/issues/433)
-  Bump github.com/go-test/deep from 1.0.7 to 1.0.8 [#529](https://github.com/keptn-contrib/dynatrace-service/pull/529) 
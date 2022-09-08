# Release Notes 0.17.1

## New Features

-  No features added

## Fixed Issues

-  QueryProcessing is overly restrictive on metric query results [#525](https://github.com/keptn-contrib/dynatrace-service/issues/525)  
-  MetricsQueryProcessing applies MV2 prefix regardless of metricUnit [#520](https://github.com/keptn-contrib/dynatrace-service/issues/520) 
-  MetricsQueryProcessing is overly restrictive on metric query results despite only querying a single metric id [#514](https://github.com/keptn-contrib/dynatrace-service/issues/514)
-  Dynatrace-service crashes with custom sli.yaml [#515](https://github.com/keptn-contrib/dynatrace-service/issues/515) 
-  Result of get-sli should be failing if SLI retrieval fails [#507](https://github.com/keptn-contrib/dynatrace-service/issues/507) 
-  No error log when sli.yaml is invalid - instead uses default slis [#413](https://github.com/keptn-contrib/dynatrace-service/issues/413) 
-  Config result produced by dashboard creation includes a URL with extra https prefix [#504](https://github.com/keptn-contrib/dynatrace-service/issues/504) 

## Other Changes

-  Remove fallbacks to sli.yaml file when dashboard processing fails [#521](https://github.com/keptn-contrib/dynatrace-service/pull/521) 

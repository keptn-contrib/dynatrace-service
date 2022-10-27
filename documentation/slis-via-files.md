# SLIs via `dynatrace/sli.yaml` files

To specify SLIs via files, add one or more `dynatrace/sli.yaml` files to Keptn project's Git repository on the project, stage or service level. Each `dynatrace/sli.yaml` file must be a well-formed [YAML file](https://yaml.org/) and contain an `indicators` element which in turn contains key-value pairs for each SLI. 


**Notes:**

Definitions can target any type of metric available in Dynatrace and any entity type (`APPLICATION`, `SERVICE`, `PROCESS_GROUP`, `HOST`, `CUSTOM_DEVICE`, etc.).

- As users would commonly like the `builtin:service.response.time` metric to be specified in milliseconds, the dynatrace-service automatically converts SLIs using this metric from microseconds to milliseconds. To convert other metrics, see [Converted metrics](#converted-metrics-prefix-mv2))

- This service uses the Dynatrace Metrics v2 API by default but can also parse v1 metrics query. If you use the v1 query language you will see warning log outputs in the *dynatrace-service* which encourages you to update your queries to v2. More information about Metrics v2 API can be found in the [Dynatrace documentation](https://www.dynatrace.com/support/help/extend-dynatrace/dynatrace-api/environment-api/metric-v2/)


## Example `dynatrace/sli.yaml` file

To assist you in getting started, consult [the example `dynatrace/sli.yaml` file](assets/sli.yaml) which contains definitions for `throughput`, `error_rate`, `response_time_p50`, `response_time_p90` and `response_time_p95`:

```yaml
spec_version: "1.0"
indicators:
 throughput: "metricSelector=builtin:service.requestCount.total:splitBy():sum&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)"
 error_rate: "metricSelector=builtin:service.errors.total.rate:splitBy():avg&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)"
 response_time_p50: "metricSelector=builtin:service.response.time:splitBy():percentile(50)&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)"
 response_time_p90: "metricSelector=builtin:service.response.time:splitBy():percentile(90)&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)"
 response_time_p95: "metricSelector=builtin:service.response.time:splitBy():percentile(95)&entitySelector=tag(keptn_project:$PROJECT),tag(keptn_stage:$STAGE),tag(keptn_service:$SERVICE),tag(keptn_deployment:$DEPLOYMENT),type(SERVICE)"
```

These SLIs may then be used by Keptn in conjunction with service-level objectives (SLOs) to evaluate quality gates. For example, the following sample SLOs could be defined in a [`slo.yaml` file](assets/slo.yaml): 

```yaml
spec_version: "1.0"
comparison:
  aggregate_function: "avg"
  compare_with: "single_result"
  include_result_with_score: "pass"
  number_of_comparison_results: 1
filter:
objectives:
  - sli: "response_time_p95"
    key_sli: false
    pass:             
      - criteria:
          - "<600"    
    warning:        
      - criteria:
          - "<=800"
    weight: 1
  - sli: "error_rate"
    key_sli: false
    pass:
      - criteria:
          - "<5"
  - sli: throughput
total_score:
  pass: "90%"
  warning: "75%"
```

## Using placeholders in SLI definitions

Queries can contain placeholders such as `$SERVICE`, `$STAGE`, `$PROJECT`, `$DEPLOYMENT` as well as `$LABEL.yourlabel1`, `$LABEL.yourlabel2` which are substituted using values from the `sh.keptn.event.get-sli.triggered` event. Further details are outlined in the section [Keptn placeholders](keptn-placeholders.md).

For example, `throughput` could be defined such that the tag name is retrieved from a label that is passed to Keptn:

```yaml
spec_version: "1.0"
indicators:
    throughput:  "metricSelector=builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum&entitySelector=tag($LABEL.dttag),type(SERVICE)"
```

If an event was then sent to the dynatrace-service including a label with the name `dttag` and a value e.g. `evaluateforsli`, it will match an entity that has this tag on it.

By using multiple labels it is possible to define SLIs that span multiple layers of your stack, e.g., services, process groups and host metrics. For example, the following `dynatrace/sli.yaml` would query one metric from a service, one from a process group and one from a host:

```yaml
spec_version: "1.0"
indicators:
    throughput:  "metricSelector=builtin:service.requestCount.total:merge(\"dt.entity.service\"):sum&entitySelector=tag($LABEL.dtservicetag),type(SERVICE)"
    gcheapuse:   "metricSelector=builtin:tech.nodejs.v8heap.gcHeapUsed:merge(\"dt.entity.process_group_instance\"):sum&entitySelector=tag($LABEL.dtpgtag),type(PROCESS_GROUP_INSTANCE)"
    hostmemory:  "metricSelector=builtin:host.mem.usage:merge(\"dt.entity.host\"):avg&entitySelector=tag($LABEL.dthosttag),type(HOST)"
```


## Supported SLI definition types

By default, the dynatrace-service queries metrics using the [Metrics v2 API](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/metric-v2/). However, by prefixing the SLI definition, other Dynatrace endpoints may be targeted. These SLI definitions have the form `<PREFIX>;<QUERY>` where `<QUERY>` is the set of parameters that should be passed to the endpoint.


### Dynatrace Metrics v2

Metrics v2 queries must specify a [`metricSelector`](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/metric-v2/metric-selector) which can also include transformations. In addition, `entitySelector`, `resolution` and `mzSelector` parameters are supported. If a query returns multiple values, the dynatrace-service will attempt to set `resolution=Inf` or apply a `:fold()` transformation to obtain a single value.

SLIs targeting the Metrics v2 API are returned in their original unit. To convert to a different unit, apply a `:toUnit(...,...)` transformation. For example, the SLI definition:

```
service_response_time: metricSelector=builtin:service.response.time:splitBy():avg:toUnit(microSecond,milliSecond)
```

will retrieve a `service_response_time` SLI in milliseconds rather than  microseconds (the default for the metric).


### Dynatrace SLO definitions (prefix: `SLO`)

With Dynatrace Version 207, Dynatrace introduced native support for SLO monitoring. The dynatrace-service is able to query these SLO definitions using the [SLO API](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/service-level-objectives/) by referencing them by SLO-ID using the syntax `SLO;<SLO_ID>`:

```yaml
spec_version: "1.0"
indicators:
    rt_faster_500ms: SLO;524ca177-849b-3e8c-8175-42b93fbc33c5
```

This queries the SLO using the `/api/v2/slo/<SLO_ID>` endpoint and will return the value of the `evaluatedPercentage` field.


### Open problems (prefix: `PV2`)

The dynatrace-service may query the [Problems API v2](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/problems-v2/) number of open problems in a particular environment, or those that match a particular problem type using the syntax `PV2;<query>` where `<query>` may include a `problemSelector` and / or `entitySelector`, e.g., `problemSelector=...&entitySelector=...`:

```yaml
spec_version: "1.0"
indicators:
    problems: PV2;problemSelector=status(open)&entitySelector=mzId(7030365576649815430)
```

This passes the `problemSelector` and `entitySelector` to the `/api/v2/problems` endpoint and will return the value of the `totalCount` field, i.e., the total number of problems matching the query, as the SLI value.


### Open security problems (prefix: `SECPV2`)

By using the syntax `SECPV2;securityProblemSelector=...` the dynatrace-service will query the [Security problems API](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/security-problems/):


```yaml
spec_version: "1.0"
indicators:
    security_problems: SECPV2;securityProblemSelector=status(open)
```

This passes the `securityProblemSelector` to the `/api/v2/securityProblems` endpoint and will return the value of the `totalCount` field, i.e., the total number of security problems matching the query, as the SLI value.


### User sessions (prefix: `USQL`)

With the syntax `USQL;<tile_type>;<dimension>;<query>`, the dynatrace-service can extract an SLI value from a user session query developed in the Dynatrace tenant. Internally, `<query>` is passed to the `/api/v1/userSessionQueryLanguage/table` endpoint as described in the [User sessions API](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/user-sessions). Parameters `tile_type` and `dimension` are then used to control how the SLI value is extracted from the query result:

| Tile type | Description |
|---|---|
| SINGLE_VALUE | Select the first column of the first row as the result; `<dimension>` should be empty. The type of the first column should be *number*. |
| PIE_CHART | Select the first row where the value in the first column equals `<dimension>`, take the value in the second column as the result. The type of the first column should be *string*, the type of the second column should be *number*. |
| COLUMN_CHART | Select the first row where the value in the first column equals `<dimension>`, take the value in the second column as the result. The type of the first column should be *string*, the type of the second column should be *number*.|
| LINE_CHART | Select the first row where the value in the first column equals `<dimension>`, take the value in the second column as the result. The type of the first column should be *string*, the type of the second column should be *number*.|
| TABLE | Select the first row where the value in the first column equals `<dimension>`, take the value of the last column as the result. The type of the first column should be *string*, the type of the last column should be *number*.|

Fox example, the following SLI definition will calculate the average duration of iPad mini user sessions in Austria:

```yaml
spec_version: "1.0"
indicators:
  ipad-mini-session-duration: USQL;COLUMN_CHART;iPad mini;SELECT device, AVG(duration) FROM usersession WHERE country IN('Austria') GROUP BY device
```


### Converted metrics (prefix: `MV2`)

To specify that a metrics query should be converted from microseconds to milliseconds or bytes to kilobytes, apply an `MV2` prefix. Currently, there are two possible prefixes for a regular query:

- `MV2;MicroSecond;`: convert the result of the query from microseconds to milliseconds
- `MV2;Byte;`: convert the result of the query from bytes to kilobytes

The following example demonstrates how to specify that a metric's unit is microseconds and should be converted to milliseconds:

```yaml
indicators:
 teststep_rt_Basic_Check: "MV2;MicroSecond;metricSelector=calc:service.teststepresponsetime:merge(\"dt.entity.service\"):avg:names:filter(eq(\"Test Step\",\"Basic Check\"))&entitySelector=type(SERVICE)"
```

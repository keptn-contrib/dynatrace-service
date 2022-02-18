# SLIs and SLOs based on a Dynatrace dashboard

The dynatrace-service can dynamically create SLIs and SLOs from a Dynatrace dashboard in response to a `sh.keptn.event.get-sli.triggered` event. To select this mode, set the `dashboard` property in the `dynatrace/dynatrace.conf.yaml` configuration file. Two options are available:

- `query`: the dynatrace-service will use the first dashboard found with a name beginning with `KQG;project=<project>;service=<service>;stage=<stage>`, where `<project>`, `<service>` and `<stage>` are taken from the `sh.keptn.event.get-sli.triggered` event. To further customize the name, append any additional description as `;<custom-description>` after the stage.
- `<dashboard-uuid>`: set the `dashboard` property to the UUID of a specific dashboard to use it.

In response to  a `sh.keptn.event.get-sli.triggered` event, the dynatrace-service will transform each supported tile into Dynatrace API queries. An SLI is created for each result together with a corresponding SLO. The SLOs are then stored in an `slo.yaml` file in the appropriate service and stage of the Keptn project, and values of the SLIs are queried and returned in the `sh.keptn.event.get-sli.finished` event.


## Defining SLIs and SLOs

The base name of the SLI as well as the properties of the SLO must be set by appending `;`-separated `<key>=<value>` pairs to the tile's title. The following keys are supported:

| Key | Description | Required | Example |
|---|---|---|---|
| `sli` | Use `<value>` as the base-name of the SLI | Yes | `sli=response_time` |
| `pass` | Add `<value>` as a pass criterion to the SLO | No | `pass=<200` |
| `warning` | Add `<value>` as a warning criterion to the SLO | No | `warning=<300` |
| `key` | Mark SLI as a key SLI | No | `key=true` |
| `weight` | Set the weight of the SLO to `<value>` | No | `weight=2` |

Consult [the Keptn documentation](https://keptn.sh/docs/0.11.x/quality_gates/slo/#objectives) for more details on configuring objectives.


## Supported tile types

The following dashboard tile types are supported:


### Data explorer tiles

Data explorer tiles must only include a single query (i.e., one metric) and include up to one *filter by* and up to one *split by* clause. Metric selectors provided via the code tab are currently not supported.


### Custom chart tiles

Each custom chart tile may only contain a single series. Furthermore, the series may only contain zero or one *dimensions* and optionally a single *filter*.


### Problems tiles

A problems tile on the dashboard is mapped to an SLI `problems` with the total count of open problems. A corresponding SLO specifies that `problems` is a key SLI with a pass criterion of `<=0`.

As dashboards currently do not offer a tile for security problems, an additional SLI `security_problems` is also added with the total count of open security problems. A corresponding SLO specifies that `security_problems` is a key SLI with a pass criterion of `<=0`.


### SLO tiles

An SLO tile will produce an SLI with the same name as the underlying SLO and the SLO status (or `evaluatedPercentage`) as the value. The SLO's pass and warning criteria are taken directly from the target and warning thresholds of the underlying SLO. Querying remote environments, or using custom management zones or timeframes is not supported.    


### USQL tiles

Depending on the query and visualization type, a USQL tile will produce one or more SLIs. Single value queries always produce a single SLI, whereas bar charts, line charts, pie charts and tables produce an SLI (and SLO) for each value of the selected dimension. The funnel visualization type is currently not supported.


## Automatic expansion of results including one or more dimensions

Results from queries created from Data Explorer, Custom Charting or USQL tiles that include one or more dimensions are automatically expanded into multiple SLIs and SLOs. In this case the SLI name specified in the tile's title is used as base and dimension values are concatenated to it to produce unique names.

For example, a Data Explorer query titled `sli=response_time;pass=<20` targeting the metric `builtin:service.response.time` and split by `dt.entity.service` that returns values for `journey service` and `account service` will result in an SLI `response_time_journey_service` and `response_time_account_service`.


## SLO Comparison and Scoring

By default, the dynatrace-service instructs Keptn to perform the evaluation of SLOs using the following comparison and scoring properties:

```yaml
comparison:
  compare_with: "single_result"
  include_result_with_score: "pass"
  aggregate_function: avg
total_score:
  pass: 90%
  warning: 75%
```

Further details about SLO comparison and scoring are provided in [the Keptn documentation](https://keptn.sh/docs/0.11.x/quality_gates/slo/).

To override these defaults, add a markdown tile to the dashboard with one or more of the following `;`-separated `<key>=<value>` pairs:

| Key | Description |
|---|---|
|`KQG.Compare.Results` | Use `<value>` as the `comparison: compare_with` value |
|`KQG.Compare.WithScore` | Use `<value>` as the `comparison: include_result_with_score` value |
|`KQG.Compare.Function` | Use `<value>` as the `comparison: aggregate_function` value |
|`KQG.Total.Pass` | Use `<value>` as the `total_score: pass` value |
|`KQG.Total.Warning` | Use `<value>` as the `total_score: warning` value |

For example, the defaults above could be specified using a markdown tile containing:

```
KQG.Total.Pass=90%;KQG.Total.Warning=75%;KQG.Compare.WithScore=pass;KQG.Compare.Results=1;KQG.Compare.Function=avg
```


## Limiting the scope of SLIs using management zones

The entities used for SLIs may be filtered either by setting a management zone for the entire dashboard or for individual tiles. In case both are specified, the management zone applied to a tile is used.

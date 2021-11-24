# SLI-provider

The dynatrace-service can support the evaluation of the quality gates by retrieving SLIs for a Keptn project, stage or service in response to a `sh.keptn.event.get-sli.triggered` event. Two modes are available: 

- [SLIs via a combination of `dynatrace/sli.yaml` files located on the Keptn service, stage and project](slis-via-files.md), or 
- [SLI and SLOs based on a Dynatrace dashboard](slis-via-dashboard.md).

The mode selected by the dynatrace-service depends on the value of the `dashboard` key in the `dynatrace/dynatrace.conf.yaml` used for a particular event as outlined in [Dashboard SLI-mode configuration (`dashboard`)`](dynatrace-conf-yaml-file.md#dashboard-sli-mode-configuration-dashboard)


## SLI evaluation in auto-remediation workflows

As part of its auto-remediation workflow, Keptn also evaluates SLOs after executing the remediation action. By default, the auto-remediation workflow can be terminated if and only if the problem has been closed in Dynatrace.

To support this, the dynatrace-service will automatically query the status of the problem that originally triggered the workflow using Dynatrace's Problem API v2. It will then append an SLI `problem_open` with the value `0` (=problem no longer open) or `1` (=problem still open). Furthermore, a default key SLO is added with a  pass criteria of `<=0` ensuring that the evaluation will only succeed if the problem is closed:

```yaml
objectives:
- sli: problem_open
  pass:
  - criteria:
    - <=0
  key_sli: true
```

Alternatively, if you'd like to add a custom SLO definition, simply override the default by defining an SLI named `problem_open` together with the appropriate pass and warning annotations.

**Note:** The Dynatrace problem associated with the remediation workflow is tracked via a label containing the Dynatrace Problem URL that is added to each Keptn event in the sequence.


## Known Limitations

- The Dynatrace Metrics API provides data with the "eventual consistency" approach. Therefore, the metrics data retrieved can be incomplete or even contain inconsistencies in case of timeframes that are within two hours of the current datetime. Usually, it takes a minute to catch up, but in extreme situations this might not be enough. The dynatrace-service tries to mitigate this issue by delaying calls to the metrics API by 60 seconds.


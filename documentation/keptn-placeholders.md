# Keptn placeholders

SLI queries (see [SLIs via `dynatrace/sli.yaml` files](slis-via-files.md)) as well as certain values in [`dynatrace/dynatrace.conf.yaml` files](dynatrace-conf-yaml-file.md) may include placeholders which are automatically replaced with values from the Keptn event or environment variables. This is very powerful as you can define generic sli.yaml files and leverage the dynamic data of a Keptn event. 


## Standard placeholders for Keptn event values

The following table provides an outline of supported placeholders:

| Placeholder | Description|
|---|---|
| `$CONTEXT` | Unique UUID value that connects various events together |
| `$EVENT` | The type of the Keptn event|
| `$SOURCE` | The source of the Keptn event|
| `$PROJECT` | The Keptn project  |
| `$STAGE` | The Keptn stage |
| `$SERVICE` | The Keptn service |
| `$DEPLOYMENT` | The Keptn deployment |
| `$TESTSTRATEGY` | The test strategy|


## Keptn event label placeholders

Keptn events may also include labels, a collection of key-value pairs passed with a Keptn event. These may be then referenced as placeholders using the syntax `$LABEL.<key>`. For example, if an evaluation is triggered from the CLI using:

```console
keptn trigger evaluation --project test-project --stage quality-gate --service=test-service --labels="my-label=my-value"
```

The label `my-label` may be referenced using the placeholder `$LABEL.my-label`.


## Environment variable placeholders

All environment variable key-value pairs available to the dynatrace-service may be referenced using the placeholder `$ENV.<key>`.

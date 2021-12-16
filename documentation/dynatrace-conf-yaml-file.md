# Configuring the dynatrace-service with `dynatrace/dynatrace.conf.yaml`

The dynatrace-service always reads its configuration from a `dynatrace/dynatrace.conf.yaml` file. While configuration files may be placed on a project, stage or service level, it is mandatory to have a `dynatrace/dynatrace.conf.yaml` file on the project level. 


## Overview

The configuration file must be a well-formed [YAML file](https://yaml.org/). The file may contain the following mappings:

| Key name| Description |
|---|---|
| `specVersion` |Specification version |
| `dtCreds` | Dynatrace API credentials secret name|
| `dashboard` | Dashboard SLI-mode configuration|
| `attachRules` | Attach rules for connecting Dynatrace entities with events |


## Dynatrace API credentials secret name (`dtCreds`)

The `dtCreds` property allows you to specify the  name of the Kubernetes secret containing Dynatrace API credentials. By default, the value `dynatrace` is used.  Further details about the structure of this secret and how to create it can be found the section [Create a Dynatrace API credentials secret](project-setup.md#1-create-a-dynatrace-api-credentials-secret).


## Dashboard SLI-mode configuration (`dashboard`)

The `dashboard` property allows you to specify if SLIs definitions should be retrieved from files or dynamically from a Dynatrace dashboard. By default this value is empty, selecting [file-based SLIs](slis-via-files.md). Alternatively, set it to a dashboard ID to target a particular dashboard, or to `query` to instruct the dynatrace-service to search for a dashboard named with the pattern `KQG;project=<project>;service=<service>;stage=<stage>`. For more details, see [SLIs and SLOs based on a Dynatrace dashboard](slis-via-dashboard.md).


## Attach rules for connecting Dynatrace entities with events (`attachRules`) 

A set of rules defining Dynatrace entities to be associated with event pushed from Keptn. Each rule consists of the types of the Dynatrace entities (for example hosts or services) to be picked as well as the tags required for matching. Further details about these can be found in the [Dynatrace events API documentation](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/events-v1/post-event/#events-post-parameter-tagmatchrule) The default attach rules used are:

```yaml
- meTypes:
  - SERVICE
  tags:
    - context: CONTEXTLESS
      key: keptn_project
      value: $PROJECT
    - context: CONTEXTLESS
      key: keptn_stage
      value: $STAGE
    - context: CONTEXTLESS
      key: keptn_service
      value: $SERVICE
```


## Customizing the configuration for a specific Keptn stage or service

When processing a Keptn event, the dynatrace-service first looks for a configuration on the service level, followed by the stage level and finally the project level. In other words, while configuration files on a service level have the highest priority, the dynatrace-service will ultimately look for a configuration file on the project level if no other `dynatrace/dynatrace.conf.yaml` can be found.

Thus if a different configuration is required on a stage or service level, create an additional `dynatrace/dynatrace.conf.yaml` file to override the configuration.

| Level | Branch | Location on branch|
|---|---|---|
| Project | default branch (main) | `dynatrace/dynatrace.conf.yaml` |
| Stage | { stage branch } | `dynatrace/dynatrace.conf.yaml` |
| Service| { stage branch } | `{service}/dynatrace/dynatrace.conf.yaml` |

**Note:** only a single configuration file is used for a given event. It is not possible to combine multiple `dynatrace/dynatrace.conf.yaml` files to override only individual configuration fields.


## Using placeholders in `dynatrace/dynatrace.conf.yaml` files

Placeholders may be used in values for `dtCreds` and `dashboard`, as well as `meTypes`, `context`, `key` and `value` values within attach rules. For more details about all available placeholders, see the topic [Keptn placeholders](keptn-placeholders.md).

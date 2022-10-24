# Dynatrace-service

![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-contrib/dynatrace-service)
![CI](https://github.com/keptn-contrib/dynatrace-service/workflows/CI/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-contrib/dynatrace-service)](https://goreportcard.com/report/github.com/keptn-contrib/dynatrace-service)

## Release validated with

||||
|---|---|---|
| Dynatrace-service: `0.23.0` | Keptn: `0.16.1` | Dynatrace: `1.245` |


## Overview

The dynatrace-service allows you to integrate Dynatrace monitoring in your Keptn sequences. It provides the following capabilities:

- [**SLI-provider**](documentation/sli-provider.md): To support the evaluation of the quality gates, the dynatrace-service can be configured to retrieve SLIs for a Keptn project, stage or service. 

- [**Forwarding events from Keptn to Dynatrace**](documentation/event-forwarding-to-dynatrace.md): The dynatrace-service can forward events such as remediation, deployment, test start/stop, evaluation or release events to Dynatrace using attach rules to ensure that the correct monitored entities are associated with the event.

- [**Forwarding problem notifications from Dynatrace to Keptn**](documentation/problem-forwarding-to-keptn.md): The dynatrace-service can support triggering remediation sequences by forwarding problem notifications from Dynatrace to a Keptn environment and ensuring that the `sh.keptn.events.problem` event is mapped to the correct project, service and stage.

- [**Automatic onboarding of monitored service entities**](documentation/auto-service-onboarding.md): The dynatrace-service can be configured to periodically check for new service entities detected by Dynatrace and automatically import these into Keptn.

### Compatibility

The dynatrace-service version `0.24.0` requires Keptn version `0.19.0` or newer.

### Upgrading to 0.18.0 or newer

If you are planning to upgrade to dynatrace-service version `0.18.0` or newer from version `0.17.1` or older, then please make sure to read and follow [these instructions on patching your secrets](documentation/patching-dynatrace-secrets.md) before doing the upgrade.

## Table of contents

- [Installation](documentation/installation.md)
  - [Downloading the latest Helm chart](documentation/installation.md#1-download-the-latest-dynatrace-service-helm-chart)
  - [Gathering Keptn credentials](documentation/installation.md#2-gather-keptn-credentials)
  - [Installing the dynatrace-service](documentation/installation.md#3-install-the-dynatrace-service )
- [Project setup](documentation/project-setup.md)
  - [Creating a Dynatrace API credentials secret](documentation/project-setup.md#1-create-a-dynatrace-api-credentials-secret)
  - [Creating a dynatrace-service configuration file](documentation/project-setup.md#2-create-a-dynatrace-service-configuration-file-dynatracedynatraceconfyaml)
  - [Configuring Dynatrace as the monitoring provider](documentation/project-setup.md#3-configure-dynatrace-as-the-monitoring-provider)
- [Feature overview](documentation/feature-overview.md)
  - [SLI provider](documentation/sli-provider.md)
    - [SLIs via `dynatrace/sli.yaml` files](documentation/slis-via-files.md)
    - [SLIs via a Dynatrace dashboard](documentation/slis-via-dashboard.md)
  - [Forwarding events from Keptn to Dynatrace](documentation/event-forwarding-to-dynatrace.md)
    - [Targeting specific entities for deployment, test, evaluation and release information](documentation/event-forwarding-to-dynatrace-to-specific-entities.md)
  - [Forwarding problem notifications from Dynatrace to Keptn](documentation/problem-forwarding-to-keptn.md)
  - [Automatic onboarding of monitored service entities](documentation/auto-service-onboarding.md)
- [Troubleshooting common problems](documentation/troubleshooting.md)
  - [No SLI provider configured](documentation/troubleshooting_no-sli-provider.md)
  - [Configure monitoring fails](documentation/troubleshooting_configure-monitoring-fails.md)
  - [Evaluation fails](documentation/troubleshooting_evaluation-fails.md)
  - [Restore Keptn wildcard subscriptions `sh.keptn.*`](documentation/troubleshooting_restore-keptn-wildcard-subscriptions.md)
- Other topics
  - [Additional installation options](documentation/additional-installation-options.md)
  - [Dynatrace API token scopes](documentation/dynatrace-api-token-scopes.md)
  - [Keptn placeholders](documentation/keptn-placeholders.md)
  - [Automatic configuration of a Dynatrace tenant](documentation/auto-tenant-configuration.md)
  - [Upgrading the dynatrace-service](documentation/other-topics.md#upgrading-the-dynatrace-service)
  - [Uninstalling the dynatrace-service](documentation/other-topics.md#uninstalling-the-dynatrace-service)
  - [Developing the dynatrace-service](documentation/other-topics.md#developing-the-dynatrace-service)

# Dynatrace-service

![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-contrib/dynatrace-service)
[![Build Status](https://travis-ci.org/keptn-contrib/dynatrace-service.svg?branch=master)](https://travis-ci.org/keptn-contrib/dynatrace-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-contrib/dynatrace-service)](https://goreportcard.com/report/github.com/keptn-contrib/dynatrace-service)

## Release validated with

||||
|---|---|---|
| Dynatrace-service `0.18.1` | Keptn `0.10.0` | Dynatrace API `1.232` |

## Overview

The dynatrace-service allows you to integrate Dynatrace monitoring in your Keptn workflow. It provides the following capabilities:

- [**SLI-provider**](documentation/sli-provider.md): To support the evaluation of the quality gates, the dynatrace-service can be configured to retrieve SLIs for a Keptn project, stage or service. 

- [**Forwarding deployment and test events from Keptn to Dynatrace**](documentation/event-forwarding-to-dynatrace.md): The dynatrace-service can forward events such deployment or test start/stop events to Dynatrace along with attach rules to ensure that the correct monitored entities are associated with the event.

- [**Forwarding problem notifications from Dynatrace to Keptn**](documentation/problem-forwarding-to-keptn.md): The dynatrace-service can support auto-remediation by forwarding problem notifications from Dynatrace to a Keptn environment and ensuring that the `sh.keptn.events.problem` event is mapped to the correct project, service and stage.

- [**Automatic onboarding of monitored service entities**](documentation/auto-service-onboarding.md): The dynatrace-service can be configured to periodically check for new service entities detected by Dynatrace and automatically import these into Keptn.

### Upgrading to 0.18.0 or newer

If you are planning to upgrade to dynatrace-service version `0.18.0` or newer from version `0.17.1` or older, then please make sure to read and follow [these instructions on patching your secrets](documentation/patching-dynatrace-secrets.md) before doing the upgrade.

## Table of contents

- [Installation](documentation/installation.md)
  - [Downloading the latest Helm chart](documentation/installation.md#download-the-latest-dynatrace-service-helm-chart)
  - [Gathering Keptn credentials](documentation/installation.md#gather-keptn-credentials)
  - [Installing the dynatrace-service](documentation/installation.md#install-the-dynatrace-service)
- [Project setup](documentation/project-setup.md)
  - [Creating a Dynatrace API credentials secret](documentation/project-setup.md#1-create-a-dynatrace-api-credentials-secret)
  - [Creating a dynatrace-service configuration file](documentation/project-setup.md#2-create-a-dynatrace-service-configuration-file)
  - [Configuring Dynatrace as the monitoring provider](documentation/project-setup.md#3-configure-dynatrace-as-the-monitoring-provider)
- [Feature overview](documentation/feature-overview.md)
  - [SLI provider](documentation/use-case-sli-proviider.md)
    - [SLIs via `dynatrace/sli.yaml` files](documentation/slis-via-files.md)
    - [SLIs via a Dynatrace dashboard](documentation/slis-via-dashboard.md)
  - [Forwarding deployment and test events from Keptn to Dynatrace](documentation/event-forwarding-to-dynatrace.md)
  - [Forwarding problem notifications from Dynatrace to Keptn](documentation/problem-forwarding-to-keptn.md)
  - [Automatic onboarding of monitored service entities](documentation/auto-service-onboarding.md)
- Other topics
  - [Additional installation options](documentation/additional-installation-options.md)
  - [Keptn placeholders](documentation/keptn-placeholders.md)
  - [Automatic creation of Dynatrace entities](documentation/generation-of-dynatrace-entities.md)
  - [Upgrading the dynatrace-service](documentation/other-topics.md#upgrading-the-dynatrace-service)
  - [Uninstalling the dynatrace-service](documentation/other-topics.md#uninstalling-the-dynatrace-service)
  - [Developing the dynatrace-service](documentation/other-topics.md#developing-the-dynatrace-service)

	
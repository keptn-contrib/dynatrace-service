# Installation

The dynatrace-service can be installed in three steps:


## 1. Download the latest dynatrace-service Helm chart

Download [the latest dynatrace-service Helm chart](https://github.com/keptn-contrib/dynatrace-service/releases/latest/) from GitHub. Please ensure that the version of the dynatrace-service is compatible with the version of Keptn you have installed by consulting the [Compatibility Matrix](compatibility.md). Details on installing or upgrading Keptn can be found on the [Keptn website](https://keptn.sh/docs/quickstart/).


## 2. Gather Keptn credentials

The dynatrace-service requires access to the Keptn API consisting of `KEPTN_ENDPOINT`, `KEPTN_API_TOKEN` and optionally `KEPTN_BRIDGE_URL`.

* To get the values for `KEPTN_ENDPOINT`, please see [Authenticate Keptn CLI](https://keptn.sh/docs/0.10.x/operate/install/#authenticate-keptn-cli).

* By default, the `KEPTN_API_TOKEN` is read from the `keptn-api-token` secret (i.e., the secret from the control-plane) and does not need to be set during installation.

* If you would like to use backlinks from your Dynatrace tenant to the Keptn Bridge, provide the service with `KEPTN_BRIDGE_URL`. For further details about this value, please see [Authenticate Keptn Bridge](https://keptn.sh/docs/0.10.x/operate/install/#authenticate-keptn-bridge).

If running on a Linux or Unix based system, you can assign these to environment variables to simplify the installation process: 

```console
KEPTN_ENDPOINT=<KEPTN_ENDPOINT>
KEPTN_BRIDGE_URL=<KEPTN_BRIDGE_URL> # optional
```

Alternatively, replace the variables with the actual values in the `helm upgrade` command in the following section.


## 3. Install the dynatrace-service

To install the dynatrace-service in the standard `keptn` namespace, execute:

```console
helm upgrade --install dynatrace-service -n keptn \
    <HELM_CHART_FILENAME> \
    --set dynatraceService.config.keptnApiUrl=$KEPTN_ENDPOINT \
    --set dynatraceService.config.keptnBridgeUrl=$KEPTN_BRIDGE_URL
```

**Notes:**
- You can select additional installation options by appending key-value pairs with the syntax `--set key=value`. Further details are provided in [Additional installation options](additional-installation-options.md).
- To target a different distributor version, set the Helm chart variable `distributor.image.tag`, i.e. by appending `--set distributor.image.tag=...`. 

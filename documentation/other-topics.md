# Other topics
 
 
## Upgrading the dynatrace-service

As outlined in [the installation guide](installation.md#download-the-latest-dynatrace-service-helm-chart), download [the latest dynatrace-service Helm chart](https://github.com/keptn-contrib/dynatrace-service/releases/latest/) from GitHub.

Then, to upgrade the dynatrace-service, execute:

```console
helm upgrade dynatrace-service -n keptn \
    <HELM_CHART_FILENAME>
```

**Note:** If you are upgrading to dynatrace-service version `0.18.0` or newer from version `0.17.1` or older, please make sure to read and follow [these instructions on patching your secrets](patching-dynatrace-secrets.md) before doing the upgrade.


## Uninstalling the dynatrace-service

To uninstall the dynatrace-service, use the Helm CLI to delete the release:

```console
helm delete -n keptn dynatrace-service
```

**Note:** This command only removes the dynatrace-service. Other components, such as the Dynatrace OneAgent on Kubernetes will be unaffected.


## Developing the dynatrace-service


### Building from source

* Get dependencies: `go mod download`
* Build locally: `go build -v -o dynatrace-service ./cmd/`
* Run tests: `go test -race -v ./...`


## Debugging

Remote debugging is supported using [Skaffold](https://skaffold.dev/) via `skaffold debug`, which starts a [Delve](https://github.com/go-delve/delve) instance prior to running the service.
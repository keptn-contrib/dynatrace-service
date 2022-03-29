# Configure monitoring fails

Keptn CLI reports:
> `Configure monitoring failed. dynatrace-service: cannot handle event: could not get configuration: could not find resource: 'dynatrace/dynatrace.conf.yaml' of project 'dashboard-config-test'`

![Configure monitoring failed](images/configure-monitoring-failed.png)

Likely cause:
- The CLI command `keptn configure monitoring dynatrace --project <project-name>` was run before adding a `dynatrace/dynatrace.conf.yaml` to the Keptn project

Suggested solution:
- Create a `dynatrace/dynatrace.conf.yaml` file on the project level. See [Configuring the dynatrace-service with `dynatrace/dynatrace.conf.yaml`](dynatrace-conf-yaml-file.md).
- Following this, re-run `keptn configure monitoring dynatrace --project <project-name>`

## Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by the `$VERSION` placeholder):

```console
helm upgrade dynatrace-service -n keptn https://github.com/keptn-contrib/dynatrace-service/releases/download/$VERSION/dynatrace-service-$VERSION.tgz
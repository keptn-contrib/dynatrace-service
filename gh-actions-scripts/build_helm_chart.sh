#!/bin/bash
# shellcheck disable=SC2181

VERSION=$1
IMAGE_TAG=$2
IMAGE=$3
DT_TAGS=$4

if [ $# -ne 3 ] && [ $# -ne 4 ]; then
  echo "Usage: $0 VERSION IMAGE_TAG IMAGE (DT_TAGS)"
  exit
fi

if [ -z "$VERSION" ]; then
  echo "No Version set, exiting..."
  exit 1
fi

if [ -z "$IMAGE_TAG" ]; then
  echo "No Image Tag set, defaulting to version"
  IMAGE_TAG=$VERSION
fi

if [ -z "$IMAGE" ]; then
  echo "No image set, exiting..."
  exit 1
fi

# set  "appVersion" and "version" in Chart.yaml file
find ./chart -name Chart.yaml -exec sed -i -- "s/appVersion: latest/appVersion: ${IMAGE_TAG}/g" {} \;
find ./chart -name Chart.yaml -exec sed -i -- "s/version: latest/version: ${VERSION}/g" {} \;

# optionally insert DT_TAGS as environment variable at the end of _dynatrace_service_environment.tpl
if [ -n "$DT_TAGS" ]; then
  find ./chart/templates -name _dynatrace_service_environment.tpl -exec sed -i -- "s/{{- end }}/- name: DT_TAGS\n  value: ${DT_TAGS}\n{{- end }}/g" {} \;
fi

mkdir installer/

# ####################
# HELM CHART
# ####################
BASE_PATH=.

helm package ${BASE_PATH}/chart --app-version "$VERSION" --version "$VERSION"
if [ $? -ne 0 ]; then
  echo "Error packaging installer, exiting..."
  exit 1
fi

mv "${IMAGE}-${VERSION}.tgz" "installer/${IMAGE}-${VERSION}.tgz"

#verify the chart
helm template --debug "installer/${IMAGE}-${VERSION}.tgz"

if [ $? -ne 0 ]; then
  echo "::error Helm Chart for ${IMAGE} has templating errors -exiting"
  exit 1
fi

echo "Generated files:"
echo " - installer/${IMAGE}-${VERSION}.tgz"

#!/bin/bash
# shellcheck disable=SC2181

VERSION=$1
IMAGE_TAG=$2
IMAGE=$3
DISTRIBUTOR_VERSION=$4

if [ $# -ne 4 ]; then
  echo "Usage: $0 VERSION IMAGE_TAG IMAGE DISTRIBUTOR_VERSION"
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

if [ -z "$DISTRIBUTOR_VERSION" ]; then
  echo "No Distributor version set, exiting..."
  exit 1
fi


# replace "appVersion: latest" with "appVersion: $VERSION" in all Chart.yaml files
IT="$IMAGE_TAG" yq e -i '.appVersion = strenv(IT)' ./chart/Chart.yaml
V="$VERSION" yq e -i '.version = strenv(V)' ./chart/Chart.yaml

# replace distributor version in 'values.yaml'
DV="$DISTRIBUTOR_VERSION" yq e -i '.distributor.image.tag = strenv(DV)' ./chart/values.yaml

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

#!/bin/bash

# ${IMAGE}=$1 ${FOLDER}=$2 ${GIT_SHA}=$3 ${VERSION}=$4 ${DATETIME}=$5
IMAGE=$1
GIT_SHA=$2
VERSION=$3
DATETIME=$4

echo "Building Docker Image ${IMAGE}:${VERSION}.${DATETIME}"
cp MANIFEST ./docker/MANIFEST
cp docker/entrypoint.sh ./docker/entrypoint.sh


# uncomment certain lines from Dockerfile that are for Travis builds only
sed -i '/#build-uncomment/s/^#build-uncomment //g' Dockerfile
cat MANIFEST
docker build . -t "${IMAGE}:${VERSION}.${DATETIME}" -t "${IMAGE}:${VERSION}" --build-arg version="${VERSION}"

if [[ $? -ne 0 ]]; then
  echo "Failed to build Docker Image ${IMAGE}:${VERSION}.${DATETIME}, exiting"
  echo "::error file=Dockerfile::Failed to build Docker Image"
  exit 1
fi

docker push "${IMAGE}"


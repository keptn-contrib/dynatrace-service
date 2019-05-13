#!/bin/sh
REGISTRY_URI=$(kubectl describe svc docker-registry -n keptn | grep IP: | sed 's~IP:[ \t]*~~')

# Deploy service
rm -f config/gen/service-build.yaml

cat config/service-build.yaml | \
  sed 's~REGISTRY_URI_PLACEHOLDER~'"$REGISTRY_URI"'~' >> config/gen/service-build.yaml

kubectl delete -f config/gen/service-build.yaml --ignore-not-found
kubectl apply -f config/gen/service-build.yaml

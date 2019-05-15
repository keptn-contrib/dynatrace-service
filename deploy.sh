#!/bin/sh

kubectl delete -f config/service.yaml --ignore-not-found
kubectl apply -f config/service.yaml

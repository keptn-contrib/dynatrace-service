# Clean up dynatrace namespace
kubectl delete services,deployments,pods --all -n dynatrace --ignore-not-found
kubectl delete namespace dynatrace --ignore-not-found
kubectl delete secret dynatrace -n keptn --ignore-not-found

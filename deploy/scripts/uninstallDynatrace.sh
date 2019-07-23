# Clean up dynatrace namespace
echo "Uninstalling Dynatrace from cluster"
kubectl delete services,deployments,pods --all -n dynatrace --ignore-not-found
kubectl delete namespace dynatrace --ignore-not-found
kubectl delete secret dynatrace -n keptn --ignore-not-found
echo "Dynatrace uninstalled"

# Deleting CRB
DT_OPERATOR_LATEST_RELEASE=$(curl -s https://api.github.com/repos/dynatrace/dynatrace-oneagent-operator/releases/latest | grep tag_name | cut -d '"' -f 4)
echo "Deleting Cluster Role Binding for Dynatrace Operator $DT_OPERATOR_LATEST_RELEASE"

kubectl delete -f https://raw.githubusercontent.com/Dynatrace/dynatrace-oneagent-operator/$DT_OPERATOR_LATEST_RELEASE/deploy/kubernetes.yaml --ignore-not-found
echo "Cluster Role Binding for Dynatrace Operator deleted"
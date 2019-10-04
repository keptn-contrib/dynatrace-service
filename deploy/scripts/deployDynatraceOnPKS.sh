#!/bin/bash

source ./utils.sh

DT_TENANT=$(cat creds_dt.json | jq -r '.dynatraceTenant')
DT_API_TOKEN=$(cat creds_dt.json | jq -r '.dynatraceApiToken')
DT_PAAS_TOKEN=$(cat creds_dt.json | jq -r '.dynatracePaaSToken')

# Deploy Dynatrace operator
DT_OPERATOR_LATEST_RELEASE=$(curl -s https://api.github.com/repos/dynatrace/dynatrace-oneagent-operator/releases/latest | grep tag_name | cut -d '"' -f 4)
print_info "Installing Dynatrace Operator $DT_OPERATOR_LATEST_RELEASE"

kubectl create namespace dynatrace
verify_kubectl $? "Creating namespace dynatrace for oneagent operator failed."

kubectl label namespace dynatrace istio-injection=disabled

kubectl apply -f https://raw.githubusercontent.com/Dynatrace/dynatrace-oneagent-operator/$DT_OPERATOR_LATEST_RELEASE/deploy/kubernetes.yaml
verify_kubectl $? "Applying Dynatrace operator failed."
wait_for_crds "oneagent"

# Create Dynatrace secret
kubectl -n dynatrace create secret generic oneagent --from-literal="apiToken=$DT_API_TOKEN" --from-literal="paasToken=$DT_PAAS_TOKEN"
verify_kubectl $? "Creating secret for Dynatrace OneAgent failed."

# Create Dynatrace OneAgent for PKS
rm -f ../manifests/dynatrace/gen/cr.yml

curl -o ../manifests/dynatrace/cr.yml https://raw.githubusercontent.com/Dynatrace/dynatrace-oneagent-operator/$DT_OPERATOR_LATEST_RELEASE/deploy/cr.yaml
cat ../manifests/dynatrace/cr.yml | sed 's~ENVIRONMENTID.live.dynatrace.com~'"$DT_TENANT"'~' >> ../manifests/dynatrace/gen/cr.yml

sed '/env:/a\
PLACEHOLDER_LINE1' ../manifests/dynatrace/gen/cr.yml > ../manifests/dynatrace/gen/cr_01.yml

sed '/PLACEHOLDER_LINE1/a\
PLACEHOLDER_LINE2' ../manifests/dynatrace/gen/cr_01.yml > ../manifests/dynatrace/gen/cr_02.yml

sed '/PLACEHOLDER_LINE2/a\
PLACEHOLDER_LINE3' ../manifests/dynatrace/gen/cr_02.yml > ../manifests/dynatrace/gen/cr_03.yml

sed '/PLACEHOLDER_LINE3/a\
PLACEHOLDER_LINE4' ../manifests/dynatrace/gen/cr_03.yml > ../manifests/dynatrace/gen/cr_04.yml

cat ../manifests/dynatrace/gen/cr_04.yml | sed 's~env:.*~env:~' > ../manifests/dynatrace/gen/cr_05.yml
cat ../manifests/dynatrace/gen/cr_05.yml | sed 's/PLACEHOLDER_LINE1/   - name: ONEAGENT_ENABLE_VOLUME_STORAGE/' > ../manifests/dynatrace/gen/cr_06.yml
cat ../manifests/dynatrace/gen/cr_06.yml | sed 's/PLACEHOLDER_LINE2/     value: "true"/' > ../manifests/dynatrace/gen/cr_07.yml
cat ../manifests/dynatrace/gen/cr_07.yml | sed 's/PLACEHOLDER_LINE3/   - name: ONEAGENT_CONTAINER_STORAGE_PATH/' > ../manifests/dynatrace/gen/cr_08.yml
cat ../manifests/dynatrace/gen/cr_08.yml | sed 's~PLACEHOLDER_LINE4~     value: /var/vcap/store~' > ../manifests/dynatrace/gen/cr_final.yml

kubectl apply -f ../manifests/dynatrace/gen/cr_final.yml
verify_kubectl $? "Deploying Dynatrace OneAgent failed."

# Apply auto tagging rules in Dynatrace
print_info "Applying auto tagging rules in Dynatrace."
./applyAutoTaggingRules.sh $DT_TENANT $DT_API_TOKEN
verify_install_step $? "Applying auto tagging rules in Dynatrace failed."
print_info "Applying auto tagging rules in Dynatrace done."

# Setup problem notification in Dynatrace
print_info "Set up problem notification in Dynatrace."
KEPTN_DNS=https://api.keptn.$(kubectl get cm -n keptn keptn-domain -ojsonpath={.data.app_domain})
KEPTN_API_TOKEN=$(kubectl get secret keptn-api-token -n keptn -ojsonpath={.data.keptn-api-token} | base64 --decode)
./setupProblemNotification.sh $DT_TENANT $DT_API_TOKEN $KEPTN_DNS $KEPTN_API_TOKEN
verify_install_step $? "Setup of problem notification in Dynatrace failed."
print_info "Setup of problem notification in Dynatrace done."

# Create secrets to be used by dynatrace-service
kubectl -n keptn create secret generic dynatrace --from-literal="DT_API_TOKEN=$DT_API_TOKEN" --from-literal="DT_TENANT=$DT_TENANT"
verify_kubectl $? "Creating dynatrace secret for keptn services failed."

# Create dynatrace-service
print_info "Deploying dynatrace-service"
kubectl apply -f ../manifests/dynatrace-service/dynatrace-service.yaml
verify_kubectl $? "Deploying dynatrace-service failed."
wait_for_deployment_in_namespace "dynatrace-service" "keptn"

kubectl apply -f ../manifests/dynatrace-service/dynatrace-service-distributors.yaml
verify_kubectl $? "Deploying dynatrace-service failed."

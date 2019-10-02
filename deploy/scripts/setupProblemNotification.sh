#!/bin/bash

source ./utils.sh

# API documentation - not detailed information yet. Please check your Dynatrace tenant Swagger API docs for more info
# https://www.dynatrace.com/support/help/extend-dynatrace/dynatrace-api/configuration-api/

DT_TENANT=$1
DT_API_TOKEN=$2
KEPTN_DNS=$3
KEPTN_TOKEN=$4
CLUSTERVERSION=$(curl -s https://$DT_TENANT/api/v1/config/clusterversion?api-token=$DT_API_TOKEN | jq -r .version[0:5])

# Check tenant is at least 1.169
if (( $(echo "$CLUSTERVERSION > 1.168" | bc -l) ))
then
  curl -X POST \
    "https://$DT_TENANT/api/config/v1/notifications?Api-Token=$DT_API_TOKEN" \
    -H 'accept: application/json; charset=utf-8' \
    -H 'Content-Type: application/json; charset=utf-8' \
    -d '{ 
      "type": "WEBHOOK", 
      "name": "keptn remediation", 
      "alertingProfile": "c21f969b-5f03-333d-83e0-4f8f136e7682", 
      "active": true, 
      "url": "'$KEPTN_DNS'/v1/event", 
      "acceptAnyCertificate": true, 
      "headers": [ 
        { "name": "x-token", "value": "'$KEPTN_TOKEN'" },
        { "name": "Content-Type", "value": "application/cloudevents+json" }
      ],
      "payload": "{\n    \"specversion\":\"0.2\",\n    \"type\":\"sh.keptn.events.problem\",\n    \"shkeptncontext\":\"{PID}\",\n    \"source\":\"dynatrace\",\n    \"id\":\"{PID}\",\n    \"time\":\"\",\n    \"contenttype\":\"application/json\",\n    \"data\": {\n        \"State\":\"{State}\",\n        \"ProblemID\":\"{ProblemID}\",\n        \"PID\":\"{PID}\",\n        \"ProblemTitle\":\"{ProblemTitle}\",\n        \"ProblemDetails\":{ProblemDetailsJSON},\n        \"ImpactedEntities\":{ImpactedEntities},\n        \"ImpactedEntity\":\"{ImpactedEntity}\"\n    }\n}\n" 

      }'

else
  echo "Cluster must be 1.169 or above, detected cluster version was $CLUSTERVERSION. Please configure notification manually"
fi


if [[ $? != '0' ]]; then
  echo ""
  print_error "Problem notification could not be created for Dynatrace tenant $DT_TENANT."
  exit 1
fi


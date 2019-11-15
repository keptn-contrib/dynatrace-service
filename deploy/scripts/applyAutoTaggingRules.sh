#!/bin/bash

source ./utils.sh

# API documentation
# https://www.dynatrace.com/support/help/extend-dynatrace/dynatrace-api/configuration/auto-tag-api/

DT_TENANT=$1
DT_API_TOKEN=$2

DT_RULE_NAME=keptn_stage
# check if rule already exists in Dynatrace tenant
export DT_ID=
export DT_ID=$(curl -X GET \
  "https://$DT_TENANT/api/config/v1/autoTags?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  | jq -r '.values[] | select(.name == "'$DT_RULE_NAME'") | .id')

# if exists, then delete it
if [ "$DT_ID" != "" ]
then
  echo "Removing $DT_RULE_NAME since exists. Replacing with new definition."
  curl -f -X DELETE \
  "https://$DT_TENANT/api/config/v1/autoTags/$DT_ID?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache'

  if [[ $? -ne 0 ]]; then
    print_error "Tagging rule: $DT_RULE_NAME could not be deleted in tenant $DT_TENANT_ID."
    exit 1
  fi
fi

curl -f -X POST \
  "https://$DT_TENANT/api/config/v1/autoTags?api-token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  -d '{
  "name": "'$DT_RULE_NAME'",
  "rules": [
    {
      "type": "SERVICE",
      "enabled": true,
      "valueFormat": "{ProcessGroup:Environment:keptn_stage}",
      "propagationTypes": [
        "SERVICE_TO_PROCESS_GROUP_LIKE"
      ],
      "conditions": [
        {
          "key": {
            "attribute": "PROCESS_GROUP_CUSTOM_METADATA",
            "dynamicKey": {
              "source": "ENVIRONMENT",
              "key": "keptn_stage"
            },
            "type": "PROCESS_CUSTOM_METADATA_KEY"
          },
          "comparisonInfo": {
            "type": "STRING",
            "operator": "EXISTS",
            "value": null,
            "negate": false,
            "caseSensitive": null
          }
        }
      ]
    }
  ]
}'

if [[ $? != '0' ]]; then
  echo ""
  print_error "Tagging rule for keptn_stage could not be created in tenant $DT_TENANT."
  exit 1
fi

DT_RULE_NAME=keptn_service
# check if rule already exists in Dynatrace tenant
export DT_ID=
export DT_ID=$(curl -X GET \
  "https://$DT_TENANT/api/config/v1/autoTags?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  | jq -r '.values[] | select(.name == "'$DT_RULE_NAME'") | .id')

# if exists, then delete it
if [ "$DT_ID" != "" ]
then
  echo "Removing $DT_RULE_NAME since exists. Replacing with new definition."
  curl -f -X DELETE \
  "https://$DT_TENANT/api/config/v1/autoTags/$DT_ID?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache'

  if [[ $? -ne 0 ]]; then
    print_error "Tagging rule: $DT_RULE_NAME could not be deleted in tenant $DT_TENANT_ID."
    exit 1
  fi
fi

curl -f -X POST \
  "https://$DT_TENANT/api/config/v1/autoTags?api-token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  -d '{
  "name": "'$DT_RULE_NAME'",
  "rules": [
    {
      "type": "SERVICE",
      "enabled": true,
      "valueFormat": "{ProcessGroup:Environment:keptn_service}",
      "propagationTypes": [
        "SERVICE_TO_PROCESS_GROUP_LIKE"
      ],
      "conditions": [
        {
          "key": {
            "attribute": "PROCESS_GROUP_CUSTOM_METADATA",
            "dynamicKey": {
              "source": "ENVIRONMENT",
              "key": "keptn_service"
            },
            "type": "PROCESS_CUSTOM_METADATA_KEY"
          },
          "comparisonInfo": {
            "type": "STRING",
            "operator": "EXISTS",
            "value": null,
            "negate": false,
            "caseSensitive": null
          }
        }
      ]
    }
  ]
}'

if [[ $? != '0' ]]; then
  echo ""
  print_error "Tagging rule for keptn_service could not be created in tenant $DT_TENANT."
  exit 1
fi

DT_RULE_NAME=keptn_project
# check if rule already exists in Dynatrace tenant
export DT_ID=
export DT_ID=$(curl -X GET \
  "https://$DT_TENANT/api/config/v1/autoTags?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  | jq -r '.values[] | select(.name == "'$DT_RULE_NAME'") | .id')

# if exists, then delete it
if [ "$DT_ID" != "" ]
then
  echo "Removing $DT_RULE_NAME since exists. Replacing with new definition."
  curl -f -X DELETE \
  "https://$DT_TENANT/api/config/v1/autoTags/$DT_ID?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache'

  if [[ $? -ne 0 ]]; then
    print_error "Tagging rule: $DT_RULE_NAME could not be deleted in tenant $DT_TENANT_ID."
    exit 1
  fi
fi

curl -f -X POST \
  "https://$DT_TENANT/api/config/v1/autoTags?api-token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  -d '{
  "name": "'$DT_RULE_NAME'",
  "rules": [
    {
      "type": "SERVICE",
      "enabled": true,
      "valueFormat": "{ProcessGroup:Environment:keptn_project}",
      "propagationTypes": [
        "SERVICE_TO_PROCESS_GROUP_LIKE"
      ],
      "conditions": [
        {
          "key": {
            "attribute": "PROCESS_GROUP_CUSTOM_METADATA",
            "dynamicKey": {
              "source": "ENVIRONMENT",
              "key": "keptn_project"
            },
            "type": "PROCESS_CUSTOM_METADATA_KEY"
          },
          "comparisonInfo": {
            "type": "STRING",
            "operator": "EXISTS",
            "value": null,
            "negate": false,
            "caseSensitive": null
          }
        }
      ]
    }
  ]
}'

if [[ $? != '0' ]]; then
  echo ""
  print_error "Tagging rule for keptn_project could not be created in tenant $DT_TENANT."
  exit 1
fi


DT_RULE_NAME=keptn_deployment
# check if rule already exists in Dynatrace tenant
export DT_ID=
export DT_ID=$(curl -X GET \
  "https://$DT_TENANT/api/config/v1/autoTags?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  | jq -r '.values[] | select(.name == "'$DT_RULE_NAME'") | .id')

# if exists, then delete it
if [ "$DT_ID" != "" ]
then
  echo "Removing $DT_RULE_NAME since exists. Replacing with new definition."
  curl -f -X DELETE \
  "https://$DT_TENANT/api/config/v1/autoTags/$DT_ID?Api-Token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache'

  if [[ $? -ne 0 ]]; then
    print_error "Tagging rule: $DT_RULE_NAME could not be deleted in tenant $DT_TENANT_ID."
    exit 1
  fi
fi



curl -f -X POST \
  "https://$DT_TENANT/api/config/v1/autoTags?api-token=$DT_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'cache-control: no-cache' \
  -d '{
  "name": "'$DT_RULE_NAME'",
  "rules": [
    {
      "type": "SERVICE",
      "enabled": true,
      "valueFormat": "{ProcessGroup:Environment:keptn_deployment}",
      "propagationTypes": [
        "SERVICE_TO_PROCESS_GROUP_LIKE"
      ],
      "conditions": [
        {
          "key": {
            "attribute": "PROCESS_GROUP_CUSTOM_METADATA",
            "dynamicKey": {
              "source": "ENVIRONMENT",
              "key": "keptn_deployment"
            },
            "type": "PROCESS_CUSTOM_METADATA_KEY"
          },
          "comparisonInfo": {
            "type": "STRING",
            "operator": "EXISTS",
            "value": null,
            "negate": false,
            "caseSensitive": null
          }
        }
      ]
    }
  ]
}'

if [[ $? != '0' ]]; then
  echo ""
  print_error "Tagging rule for keptn_deployment could not be created in tenant $DT_TENANT."
  exit 1
fi
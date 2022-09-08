#!/usr/bin/env bash

if [ "$1" = "--dry-run" ]; then
  DRY_RUN=1
else
  DRY_RUN=0
fi

secret_patch="
metadata:
  labels:
    app.kubernetes.io/managed-by: keptn-secret-service
    app.kubernetes.io/scope: dynatrace-service
"

# Loop through all namespaces
for ns in $(kubectl get ns | awk 'FNR > 1 {print $1}'); do

        kubectl -n $ns get role dynatrace-service-secrets > /dev/null 2>&1
        STATUS=$?
        echo ======================================================================================
        if [ "$STATUS" = "0" ]; then
            PATCH=""
            echo Found dynatrace-service in namespace $ns.

            for secret in $(kubectl -n $ns get secret | awk 'FNR > 1 {print $1}'); do
                # Check if secret is a dynatrace-service secret
                ISDTSECRET=$(kubectl -n $ns get secret $secret -o jsonpath="{.data.DT_TENANT}")

                if [ ! -z "$ISDTSECRET" ]; then
                    echo $secret is a dynatrace secret

                    # Backup existing Secret with a timestamp
                    kubectl -n $ns get secret $secret -o yaml > $secret-$ns-secret-$(date +%s).yaml

                    # Patch the secret with secret-service annotations
                    if [ "$DRY_RUN" = "0" ]; then
                      kubectl -n $ns patch secret $secret --patch "$secret_patch"
                    else
                      echo dry-run: Please execute the following command to apply changes
                      echo kubectl -n $ns patch secret $secret --patch "$secret_patch"
                    fi

                    # Add secret name to the resourceNames list to add apply later
                    if [ -z "$PATCH" ]; then
                        PATCH="- $secret
"
                    else
                        PATCH="$PATCH      - $secret
"
                    fi
                fi
            done # END LOOP FOR SECRETS

            # If we need to patch and add the role and rolebinding
            if [ ! -z "$PATCH" ]; then
                cat >keptn-dynatrace-svc-read-role-$ns.yaml <<EOL
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: keptn-dynatrace-svc-read
  labels:
    app.kubernetes.io/managed-by: keptn-secret-service
    app.kubernetes.io/scope: dynatrace-service
rules:
  - verbs:
      - get
    apiGroups:
      - ''
    resources:
      - secrets
    resourceNames:
      ${PATCH}
EOL

                echo "Add new role keptn-dynatrace-svc-read to $ns"
                if [ "$DRY_RUN" = "0" ]; then
                  kubectl -n $ns apply -f keptn-dynatrace-svc-read-role-$ns.yaml
                else
                  echo dry-run: Please execute the following command to apply changes
                  echo kubectl -n $ns apply -f keptn-dynatrace-svc-read-role-$ns.yaml
                fi

                cat >dynatrace-service-rolebinding-$ns.yaml <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: dynatrace-service-rolebinding
  labels:
    app.kubernetes.io/managed-by: keptn-secret-service
    app.kubernetes.io/scope: dynatrace-service
subjects:
  - kind: ServiceAccount
    name: dynatrace-service
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: keptn-dynatrace-svc-read
EOF

                echo "Add new rolebinding dynatrace-service-rolebinding to $ns"
                if [ "$DRY_RUN" = "0" ]; then
                  kubectl -n $ns apply -f dynatrace-service-rolebinding-$ns.yaml
                else
                  echo dry-run: Please execute the following command to apply changes
                  echo kubectl -n $ns apply -f dynatrace-service-rolebinding-$ns.yaml
                fi
            else
                echo "No dynatrace-service secrets detected to add the role and rolebinding."
            fi

        else # Namespace does not contain the role dynatrace-service-secret
            echo "No dynatrace-service found in namespace $ns."
        fi

done

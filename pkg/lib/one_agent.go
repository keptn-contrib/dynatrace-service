package lib

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dtOperatorReleaseInfo struct {
	TagName string `json:"tag_name"`
}

func (dt *DynatraceHelper) CheckDTIsInstalled() error {
	_, err := dt.KubeApi.AppsV1().Deployments("dynatrace").Get("dynatrace-oneagent-operator", metav1.GetOptions{})
	if err != nil {
		dt.Logger.Info(`
Dynatrace OneAgent Operator is not installed on cluster
# ATTENTION # ------------------------------------------------------------------------------------
The behavior has changed and Dynatrace OneAgent Operator will NOT be installed automatically.
If you want to roll-out the Dynatrace OneAgent Operator: 
1.) Please follow the instructions as provided here: 
    https://www.dynatrace.com/support/help/technology-support/cloud-platforms/kubernetes/deploy-oneagent-k8/
2.) Then, re-deploy the dynatrace-service: 
    kubectl apply -f https://raw.githubusercontent.com/keptn-contrib/dynatrace-service/<VERSION>/deploy/service.yaml
--------------------------------------------------------------------------------------------------`)
	} else {
		dt.Logger.Info("Dynatrace OneAgent Operator is installed on cluster")
	}
	return nil
}


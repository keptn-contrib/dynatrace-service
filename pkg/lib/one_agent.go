package lib

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dtOperatorReleaseInfo struct {
	TagName string `json:"tag_name"`
}

func (dt *DynatraceHelper) EnsureDTIsInstalled() error {
	if dt.isDynatraceDeployed() {
		dt.Logger.Info("Dynatrace OneAgent Operator is installed on cluster")
	} else {
		dt.Logger.Info("Dynatrace OneAgent Operator is not installed on cluster")
		dt.Logger.Info("# ATTENTION # ------------------------------------------------------------------------------------------")
		dt.Logger.Info("The behavior has changed and Dynatrace OneAgent Operator will NOT be installed.")
		dt.Logger.Info("If you want to roll-out the OneAgent, please follow the instructions as provided here: ")
		dt.Logger.Info("https://www.dynatrace.com/support/help/technology-support/cloud-platforms/kubernetes/deploy-oneagent-k8/")
		dt.Logger.Info("--------------------------------------------------------------------------------------------------------")
	}
	return nil
}

func (dt *DynatraceHelper) isDynatraceDeployed() bool {
	_, err := dt.KubeApi.AppsV1().Deployments("dynatrace").Get("dynatrace-oneagent-operator", metav1.GetOptions{})
	if err != nil {
		dt.Logger.Error(err.Error())
		return false
	}
	return true
}


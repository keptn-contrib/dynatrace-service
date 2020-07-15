package lib

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dtOperatorReleaseInfo struct {
	TagName string `json:"tag_name"`
}

func (dt *DynatraceHelper) EnsureDTIsInstalled() error {
	if dt.isDynatraceDeployed() {
		dt.Logger.Info("Dynatrace OneAgent is installed on cluster")
	} else {
		dt.Logger.Info("Dynatrace OneAgent is not installed on cluster")
	}
	return nil
}

func (dt *DynatraceHelper) isDynatraceDeployed() bool {
	_, err := dt.KubeApi.AppsV1().Deployments("dynatrace").Get("dynatrace-oneagent-operator", metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}


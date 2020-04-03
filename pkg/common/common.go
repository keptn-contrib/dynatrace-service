package common

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

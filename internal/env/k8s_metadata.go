package env

import (
	"fmt"
	"os"
	"strings"
)

const (
	deploymentNameEnvironmentVariable      = "K8S_DEPLOYMENT_NAME"
	deploymentVersionEnvironmentVariable   = "K8S_DEPLOYMENT_VERSION"
	deploymentComponentEnvironmentVariable = "K8S_DEPLOYMENT_COMPONENT"
	podNameEnvironmentVariable             = "K8S_POD_NAME"
	namespaceEnvironmentVariable           = "K8S_NAMESPACE"
	nodeNameEnvironmentVariable            = "K8S_NODE_NAME"
)

// K8sMetadata holds K8s metadata.
type K8sMetadata struct {
	deploymentName      string
	deploymentVersion   string
	deploymentComponent string
	podName             string
	namespace           string
	nodeName            string
}

// DeploymentName gets the K8s deployment name.
func (m *K8sMetadata) DeploymentName() string {
	return m.deploymentName
}

// DeploymentVersion gets the K8s deployment version.
func (m *K8sMetadata) DeploymentVersion() string {
	return m.deploymentVersion
}

// DeploymentComponent gets the K8s deployment component.
func (m *K8sMetadata) DeploymentComponent() string {
	return m.deploymentComponent
}

// PodName gets the K8s pod name.
func (m *K8sMetadata) PodName() string {
	return m.podName
}

// Namespace gets the K8s namespace.
func (m *K8sMetadata) Namespace() string {
	return m.namespace
}

// NodeName gets the K8s node name.
func (m *K8sMetadata) NodeName() string {
	return m.nodeName
}

// GetK8sMetadata gets K8s metadata from environment variables or returns an error if it is incomplete.
func GetK8sMetadata() (*K8sMetadata, error) {
	var missingEnvironmentVariables []string

	deploymentName, found := os.LookupEnv(deploymentNameEnvironmentVariable)
	if !found {
		missingEnvironmentVariables = append(missingEnvironmentVariables, deploymentNameEnvironmentVariable)
	}

	deploymentVersion, found := os.LookupEnv(deploymentVersionEnvironmentVariable)
	if !found {
		missingEnvironmentVariables = append(missingEnvironmentVariables, deploymentVersionEnvironmentVariable)
	}

	deploymentComponent, found := os.LookupEnv(deploymentComponentEnvironmentVariable)
	if !found {
		missingEnvironmentVariables = append(missingEnvironmentVariables, deploymentComponentEnvironmentVariable)
	}

	podName, found := os.LookupEnv(podNameEnvironmentVariable)
	if !found {
		missingEnvironmentVariables = append(missingEnvironmentVariables, podNameEnvironmentVariable)
	}

	namespace, found := os.LookupEnv(namespaceEnvironmentVariable)
	if !found {
		missingEnvironmentVariables = append(missingEnvironmentVariables, namespaceEnvironmentVariable)
	}

	nodeName, found := os.LookupEnv(nodeNameEnvironmentVariable)
	if !found {
		missingEnvironmentVariables = append(missingEnvironmentVariables, nodeNameEnvironmentVariable)
	}

	if len(missingEnvironmentVariables) > 0 {
		return nil, fmt.Errorf("missing environment variables: %s", strings.Join(missingEnvironmentVariables, ","))
	}

	return &K8sMetadata{
		deploymentName:      deploymentName,
		deploymentVersion:   deploymentVersion,
		deploymentComponent: deploymentComponent,
		podName:             podName,
		namespace:           namespace,
		nodeName:            nodeName,
	}, nil
}

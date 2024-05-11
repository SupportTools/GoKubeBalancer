package k8sutils

import (
	"context"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetConfig retrieves the Kubernetes configuration from Rancher
func GetConfig(ctx context.Context) (*rest.Config, error) {
	logger.Info("Retrieving cluster ID...")
	clusterID, err := GetClusterID()
	if err != nil {
		logger.Errorf("Failed to get cluster ID: %v", err)
		return nil, err
	}
	logger.Infof("Cluster ID obtained: %s", clusterID)

	logger.Info("Generating kubeconfig for the cluster...")
	kubeconfigString, err := GenerateKubeconfig(ctx, clusterID)
	if err != nil {
		logger.Errorf("Failed to generate kubeconfig: %v", err)
		return nil, err
	}

	logger.Info("Creating Kubernetes client configuration from kubeconfig...")
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigString))
	if err != nil {
		logger.Errorf("Failed to create Kubernetes client config from kubeconfig string: %v", err)
		return nil, err
	}

	logger.Infof("Successfully retrieved and configured Kubernetes client for cluster ID: %s", clusterID)
	return config, nil
}

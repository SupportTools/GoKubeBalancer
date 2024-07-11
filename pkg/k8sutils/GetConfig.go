package k8sutils

import (
	"context"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetConfig retrieves the Kubernetes configuration from Rancher
func GetConfig(ctx context.Context) (*rest.Config, error) {
	log.Info("Retrieving cluster ID...")
	clusterID, err := GetClusterID()
	if err != nil {
		log.Errorf("Failed to get cluster ID: %v", err)
		return nil, err
	}
	log.Infof("Cluster ID obtained: %s", clusterID)

	log.Info("Generating kubeconfig for the cluster...")
	kubeconfigString, err := GenerateKubeconfig(ctx, clusterID)
	if err != nil {
		log.Errorf("Failed to generate kubeconfig: %v", err)
		return nil, err
	}

	log.Info("Creating Kubernetes client configuration from kubeconfig...")
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigString))
	if err != nil {
		log.Errorf("Failed to create Kubernetes client config from kubeconfig string: %v", err)
		return nil, err
	}

	log.Infof("Successfully retrieved and configured Kubernetes client for cluster ID: %s", clusterID)
	return config, nil
}

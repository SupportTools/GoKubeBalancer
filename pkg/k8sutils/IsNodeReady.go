package k8sutils

import (
	"context"
	"fmt"

	"github.com/supporttools/GoKubeBalancer/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// IsNodeReady checks if the node is in a ready state.
func IsNodeReady(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) (bool, error) {
	if config.CFG.Debug {
		logger.Printf("Checking readiness for node %s", nodeName)
	}

	// Get the current status of the node from Kubernetes
	node, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get node %s: %v", nodeName, err)
	}

	for _, condition := range node.Status.Conditions {
		// Log each condition found in the node status
		if config.CFG.Debug {
			logger.Printf("Node %s condition type: %s, status: %s", nodeName, condition.Type, condition.Status)
		}

		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			// Log the positive readiness condition
			logger.Printf("Node %s is ready.", nodeName)
			return true, nil
		}
	}

	// Log the negative outcome if no ready condition is met
	if config.CFG.Debug {
		logger.Printf("Node %s is not ready. Conditions: %v", nodeName, node.Status.Conditions)
	}
	return false, nil
}

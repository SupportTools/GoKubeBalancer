package k8sutils

import (
	"time"

	"github.com/supporttools/GoKubeBalancer/pkg/config"
	v1 "k8s.io/api/core/v1"
)

// IsNewNode checks if a node is considered "new" based on a configurable age threshold.
func IsNewNode(node *v1.Node) bool {
	nodeAge := time.Since(node.CreationTimestamp.Time)
	logger.Printf("Checking if node %s is new. Age: %s, Threshold: %s", node.Name, nodeAge, config.CFG.NewNodeThreshold)

	if nodeAge < config.CFG.NewNodeThreshold {
		logger.Printf("Node %s is considered new (age %s).", node.Name, nodeAge)
		return true
	}

	logger.Printf("Node %s is not considered new (age %s).", node.Name, nodeAge)
	return false
}

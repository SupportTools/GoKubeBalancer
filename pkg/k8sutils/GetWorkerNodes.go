package k8sutils

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/supporttools/GoKubeBalancer/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const defaultPort = "80"

// NodeDetails holds the necessary details for backend nodes
type NodeDetails struct {
	Name string
	IP   string
}

// GetWorkerNodes retrieves a list of node details for nodes based on the configured node selector
func GetWorkerNodes(ctx context.Context, clientset *kubernetes.Clientset) ([]NodeDetails, error) {
	nodeSelector := config.CFG.NodeSelector
	logrus.Debugf("Retrieving nodes with selector: %s", nodeSelector)

	var nodes *v1.NodeList
	var err error
	for retries := 0; retries < 5; retries++ {
		nodes, err = clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
			LabelSelector: nodeSelector,
		})
		if err != nil {
			logrus.Errorf("Failed to list nodes (attempt %d): %v", retries+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}

	var details []NodeDetails
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" {
				details = append(details, NodeDetails{
					Name: node.Name,
					IP:   address.Address + ":" + defaultPort,
				})
			}
		}
	}
	return details, nil
}

package main

import (
	"context"
	"flag"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/supporttools/GoKubeBalancer/pkg/backend"
	"github.com/supporttools/GoKubeBalancer/pkg/config"
	"github.com/supporttools/GoKubeBalancer/pkg/k8sutils"
	"github.com/supporttools/GoKubeBalancer/pkg/logging"
	"github.com/supporttools/GoKubeBalancer/pkg/metrics"
	"github.com/supporttools/GoKubeBalancer/pkg/network"
	"k8s.io/client-go/kubernetes"
)

func main() {
	flag.Parse()
	if err := config.LoadConfiguration(); err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}
	logger := logging.SetupLogging()
	logger.Debug("Debug logging enabled")

	ctx := context.Background()
	var clientset *kubernetes.Clientset

	for {
		logger.Info("Connecting to Kubernetes cluster...")
		kubeConfig, err := k8sutils.GetConfig(ctx)
		if err != nil {
			logger.Errorf("Failed to get Kubernetes config: %v", err)
			time.Sleep(10 * time.Second) // Retry after 10 seconds
			continue
		}
		clientset, err = kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			logger.Errorf("Failed to create Kubernetes clientset: %v", err)
			time.Sleep(10 * time.Second) // Retry after 10 seconds
			continue
		}
		break
	}

	logger.Info("Retrieving worker nodes...")
	workerNodes, err := k8sutils.GetWorkerNodes(ctx, clientset)
	if err != nil {
		logger.Fatalf("Failed to retrieve worker nodes: %v", err)
	}

	logger.Info("Starting GoKubeBalancer...")
	backendManager := backend.NewManager(workerNodes, config.CFG.RescanInterval, clientset)
	go backendManager.HealthChecker(ctx) // Start health checking

	go func() {
		logger.Println("Starting metrics server...")
		metrics.StartMetricsServer()
	}()

	// Assuming config.CFG defines BackendHttpPort and BackendHttpsPort
	tcpBalancer80 := network.NewTCPBalancer(config.CFG.FrontendHttpPort, 80, backendManager)    // Use 80 for backend health checks
	tcpBalancer443 := network.NewTCPBalancer(config.CFG.FrontendHttpsPort, 443, backendManager) // Use 443 for backend health checks
	go tcpBalancer80.Start()
	go tcpBalancer443.Start()

	select {} // Block forever
}

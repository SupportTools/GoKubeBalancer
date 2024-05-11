package main

import (
	"flag"

	"github.com/supporttools/GoKubeBalancer/pkg/backend"
	"github.com/supporttools/GoKubeBalancer/pkg/config"
	"github.com/supporttools/GoKubeBalancer/pkg/logging"
	"github.com/supporttools/GoKubeBalancer/pkg/metrics"
	"github.com/supporttools/GoKubeBalancer/pkg/network"
)

func main() {
	flag.Parse()
	config.LoadConfiguration()
	logger := logging.SetupLogging()
	logger.Debug("Debug logging enabled")

	logger.Info("Starting GoKubeBalancer...")
	backendManager := backend.NewManager(config.CFG.BackendMembers, config.CFG.RescanInterval)
	go backendManager.HealthChecker() // Start health checking

	metrics.StartMetricsServer()

	tcpBalancer80 := network.NewTCPBalancer(config.CFG.HttpPort, backendManager)
	tcpBalancer443 := network.NewTCPBalancer(config.CFG.HttpPort, backendManager)
	go tcpBalancer80.Start()
	go tcpBalancer443.Start()

	select {}
}

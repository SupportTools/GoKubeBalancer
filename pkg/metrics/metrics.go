package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporttools/GoKubeBalancer/pkg/config"
	"github.com/supporttools/GoKubeBalancer/pkg/health"
	"github.com/supporttools/GoKubeBalancer/pkg/logging"
)

var logger = logging.SetupLogging()

var (
	totalRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "load_balancer_requests_total",
		Help: "Total number of requests handled by the load balancer.",
	}, []string{"backend"})
	healthyBackendsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "load_balancer_healthy_backends",
		Help: "Number of healthy backends.",
	}, []string{"backend"})
)

func StartMetricsServer() {
	if config.CFG.MetricsPort == 0 {
		logger.Fatalf("Metrics server port not configured")
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/healthz", health.HealthzHandler())
	mux.Handle("/readyz", health.ReadyzHandler())
	mux.Handle("/version", health.VersionHandler())
	mux.HandleFunc("/node-states", health.NodeStatesHandler)

	serverPortStr := strconv.Itoa(config.CFG.MetricsPort)
	logger.Infof("Metrics server starting on port %s", serverPortStr)

	if err := http.ListenAndServe(":"+serverPortStr, mux); err != nil {
		logger.Fatalf("Metrics server failed to start: %v", err)
	}
}

package backend

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/supporttools/GoKubeBalancer/pkg/k8sutils"
	"github.com/supporttools/GoKubeBalancer/pkg/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var log = logging.SetupLogging()

// BackendManager encapsulates backend management
type BackendManager struct {
	currentIndex        uint32
	backendList         map[string]k8sutils.NodeDetails
	ipMap               map[string]string
	healthMap           map[string]bool
	clientset           *kubernetes.Clientset
	mutex               sync.Mutex
	healthMutex         sync.Mutex
	healthCheckInterval time.Duration
}

// NewManager creates a new backend Manager
func NewManager(backends []k8sutils.NodeDetails, interval time.Duration, cs *kubernetes.Clientset) *BackendManager {
	log.Println("[Backend Manager] Initializing BackendManager with provided node details and interval.")
	backendManager := &BackendManager{
		backendList:         make(map[string]k8sutils.NodeDetails),
		ipMap:               make(map[string]string),
		healthMap:           make(map[string]bool),
		clientset:           cs,
		healthCheckInterval: interval,
	}

	for _, detail := range backends {
		// Ensure detail.IP does not include the port here
		ipWithoutPort := strings.Split(detail.IP, ":")[0]
		backendManager.backendList[detail.Name] = k8sutils.NodeDetails{Name: detail.Name, IP: ipWithoutPort}
		backendManager.healthMap[ipWithoutPort] = false
		log.Debugf("[Backend Manager] Added backend: %s with IP: %s to management pool.", detail.Name, ipWithoutPort)
	}

	return backendManager
}

// HealthChecker runs a loop to check the health of all backends periodically
func (bm *BackendManager) HealthChecker(ctx context.Context) {
	log.Println("[Health Checker] Starting HealthChecker.")
	ticker := time.NewTicker(bm.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[Health Checker] Context cancelled, stopping health checks.")
			return
		case <-ticker.C:
			log.Println("[Health Checker] Performing scheduled health checks on all backends.")
			bm.checkAllBackends(ctx)
		}
	}
}

// checkAllBackends iterates over all backends and checks their health
func (bm *BackendManager) checkAllBackends(ctx context.Context) {
	bm.healthMutex.Lock()
	defer bm.healthMutex.Unlock()
	for name, detail := range bm.backendList {
		log.Debugf("[Health Checker] Initiating health check for backend %s.", name)
		go bm.checkHealth(ctx, detail)
	}
}

// checkHealth performs a health check by making an HTTP request to the backend's health endpoint and checking Kubernetes node status
func (bm *BackendManager) checkHealth(ctx context.Context, detail k8sutils.NodeDetails) {
	// Health check should always be on port 80
	healthCheckURL := "http://" + detail.IP + ":80/healthz"
	log.Debugf("[Health Checker] Checking HTTP health for backend %s at %s.", detail.Name, healthCheckURL)
	resp, err := http.Get(healthCheckURL)
	if err != nil || resp.StatusCode != 200 {
		log.Debugf("[Health Checker] HTTP health check failed for backend %s (%s): %v", detail.Name, healthCheckURL, err)
		bm.setBackendHealth(detail.IP, false)
		return
	}
	resp.Body.Close()
	log.Debugf("[Health Checker] HTTP health check passed for backend %s (%s).", detail.Name, healthCheckURL)

	// Check Kubernetes node state if cluster connection is available
	if bm.clientset != nil {
		node, err := bm.clientset.CoreV1().Nodes().Get(ctx, detail.Name, metav1.GetOptions{})
		if err != nil {
			log.Debugf("[Health Checker] Failed to retrieve node details for backend %s: %v", detail.Name, err)
			bm.setBackendHealth(detail.IP, true) // Fallback to HTTP health check
			return
		}

		if k8sutils.IsNewNode(node) {
			log.Debugf("[Health Checker] Backend %s (%s) is new and not ready for traffic.", detail.Name, detail.IP)
			bm.setBackendHealth(detail.IP, false)
			return
		}

		ready, err := k8sutils.IsNodeReady(ctx, bm.clientset, detail.Name)
		if err != nil || !ready {
			log.Debugf("[Health Checker] Kubernetes node readiness check failed for backend %s: %v", detail.Name, err)
			bm.setBackendHealth(detail.IP, false)
			return
		}

		log.Debugf("[Health Checker] Backend %s (%s) is healthy and ready to handle traffic.", detail.Name, detail.IP)
		bm.setBackendHealth(detail.IP, true)
	} else {
		log.Warnf("[Health Checker] Skipping Kubernetes node check for backend %s (%s) due to missing clientset.", detail.Name, detail.IP)
		bm.setBackendHealth(detail.IP, true) // Fallback to HTTP health check
	}
}

// setBackendHealth updates the health status of a specific backend
func (bm *BackendManager) setBackendHealth(backendIP string, isHealthy bool) {
	bm.healthMutex.Lock()
	defer bm.healthMutex.Unlock()
	oldStatus, exists := bm.healthMap[backendIP]
	if exists {
		log.Debugf("[Backend Manager] Backend %s health status changed from %t to %t.", backendIP, oldStatus, isHealthy)
	} else {
		log.Debugf("[Backend Manager] Setting health status for new backend %s to %t.", backendIP, isHealthy)
	}
	bm.healthMap[backendIP] = isHealthy
}

// IsBackendHealthy returns the health status of the specified backend
func (bm *BackendManager) IsBackendHealthy(backendIP string) bool {
	log.Debugf("[Backend Manager] Checking health status for backend %s.", backendIP)
	bm.healthMutex.Lock()
	defer bm.healthMutex.Unlock()
	isHealthy, exists := bm.healthMap[backendIP]
	if !exists {
		log.Debugf("[Backend Manager] Backend %s not found in health map.", backendIP)
	}
	log.Debugf("[Backend Manager] Backend %s health status: %t.", backendIP, isHealthy)
	return exists && isHealthy
}

// GetBackendByIP returns the IP of the backend associated with the given client IP
func (bm *BackendManager) GetBackendByIP(ip string) string {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	backendIP, exists := bm.ipMap[ip]
	if exists && bm.IsBackendHealthy(backendIP) {
		log.Debugf("[Backend Manager] Found healthy backend %s for client IP %s.", backendIP, ip)
		return backendIP
	}

	log.Warnf("[Backend Manager] No healthy backend found for client IP %s, reselecting.", ip)
	newBackendIP := bm.selectNewBackend(ip)

	bm.ipMap[ip] = newBackendIP
	return newBackendIP
}

// selectNewBackend performs a round-robin selection to find a healthy backend
func (bm *BackendManager) selectNewBackend(ip string) string {
	log.Debugf("[Backend Manager] Selecting new backend for IP %s using round-robin method.", ip)
	nodeNames := make([]string, 0, len(bm.backendList))

	for name := range bm.backendList {
		nodeNames = append(nodeNames, name)
	}

	if len(nodeNames) == 0 {
		log.Debugf("[Backend Manager] No backends available for selection.")
		return ""
	}

	totalBackends := uint32(len(nodeNames))

	for i := uint32(0); i < totalBackends; i++ {
		currentIndex := (atomic.LoadUint32(&bm.currentIndex) + i) % totalBackends
		backendName := nodeNames[currentIndex]

		if bm.IsBackendHealthy(bm.backendList[backendName].IP) {
			atomic.StoreUint32(&bm.currentIndex, (currentIndex+1)%totalBackends)
			log.Debugf("[Backend Manager] New healthy backend assigned: %s for IP %s", backendName, ip)
			return bm.backendList[backendName].IP
		}
	}

	log.Debugf("[Backend Manager] No healthy backends available for IP %s after round-robin selection.", ip)
	return ""
}

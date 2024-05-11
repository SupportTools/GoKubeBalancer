package backend

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// BackendManager encapsulates backend management
type BackendManager struct {
	currentIndex        uint32
	backendList         []string
	ipMap               map[string]string
	healthMap           map[string]bool
	mutex               sync.Mutex
	healthMutex         sync.Mutex
	healthCheckInterval time.Duration
}

// NewManager creates a new backend Manager with specified backends and health check interval
func NewManager(backends []string, interval time.Duration) *BackendManager {
	mgr := &BackendManager{
		backendList:         backends,
		ipMap:               make(map[string]string),
		healthMap:           make(map[string]bool),
		healthCheckInterval: interval,
	}
	for _, backend := range backends {
		mgr.setBackendHealth(backend, false) // Initialize all as unhealthy
	}
	return mgr
}

// HealthChecker runs a loop to check the health of all backends periodically
func (bm *BackendManager) HealthChecker() {
	for {
		bm.checkAllBackends()
		time.Sleep(bm.healthCheckInterval) // Use the configurable interval
	}
}

// checkAllBackends iterates over all backends and checks their health
func (bm *BackendManager) checkAllBackends() {
	bm.healthMutex.Lock()
	defer bm.healthMutex.Unlock()
	for _, backend := range bm.backendList {
		go bm.checkHealth(backend)
	}
}

// checkHealth performs a health check by making an HTTP request to the backend's health endpoint
func (bm *BackendManager) checkHealth(backend string) {
	resp, err := http.Get("http://" + backend + "/healthz")
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Health check failed for backend %s: %v", backend, err)
		bm.setBackendHealth(backend, false)
	} else {
		defer resp.Body.Close()
		bm.setBackendHealth(backend, true)
		log.Printf("Backend %s is healthy", backend)
	}
}

// setBackendHealth updates the health status of a specific backend
func (bm *BackendManager) setBackendHealth(backend string, isHealthy bool) {
	bm.healthMutex.Lock()
	defer bm.healthMutex.Unlock()
	bm.healthMap[backend] = isHealthy
}

// IsBackendHealthy returns the health status of the specified backend
func (bm *BackendManager) IsBackendHealthy(backend string) bool {
	bm.healthMutex.Lock()
	defer bm.healthMutex.Unlock()
	isHealthy, exists := bm.healthMap[backend]
	return exists && isHealthy
}

// GetBackendByIP selects a backend for the given IP address, using round-robin selection and ensuring the backend is healthy
func (bm *BackendManager) GetBackendByIP(ip string) string {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	if backend, exists := bm.ipMap[ip]; exists && bm.IsBackendHealthy(backend) {
		log.Printf("Returning existing healthy backend %s for IP %s", backend, ip)
		return backend
	}
	return bm.selectNewBackend(ip)
}

// selectNewBackend performs a round-robin selection to find a healthy backend
func (bm *BackendManager) selectNewBackend(ip string) string {
	originalIndex := atomic.LoadUint32(&bm.currentIndex)
	currentIndex := originalIndex
	totalBackends := uint32(len(bm.backendList))

	for {
		backend := bm.backendList[currentIndex]
		if bm.IsBackendHealthy(backend) {
			atomic.StoreUint32(&bm.currentIndex, (currentIndex+1)%totalBackends)
			bm.ipMap[ip] = backend
			log.Printf("New healthy backend assigned: %s for IP %s", backend, ip)
			return backend
		}

		currentIndex = (currentIndex + 1) % totalBackends
		if currentIndex == originalIndex {
			log.Printf("No healthy backends available for IP %s", ip)
			break
		}
	}
	return ""
}

// RemoveBackendIPMapping removes the mapping from IP to backend, typically called when a session ends or a backend becomes unhealthy
func (bm *BackendManager) RemoveBackendIPMapping(ip string) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()
	delete(bm.ipMap, ip)
}

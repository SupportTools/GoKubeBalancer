package network

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/supporttools/GoKubeBalancer/pkg/backend"
)

// TCPBalancer manages TCP connections and routes them to backends
type TCPBalancer struct {
	Port           int
	backendManager *backend.BackendManager
}

// NewTCPBalancer creates a new instance of TCPBalancer with a BackendManager
func NewTCPBalancer(port int, bm *backend.BackendManager) *TCPBalancer {
	return &TCPBalancer{
		Port:           port,
		backendManager: bm, // Initialize the BackendManager field
	}
}

// Start listens on the specified port and handles incoming connections
func (tb *TCPBalancer) Start() {

	listenAddr := "0.0.0.0:" + strconv.Itoa(tb.Port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", tb.Port, err)
	}
	defer listener.Close()

	log.Printf("TCP Load Balancer started on port %d", tb.Port)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go tb.handleConnection(clientConn)
	}
}

// handleConnection manages a single client connection, routing it to an appropriate backend
func (tb *TCPBalancer) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	ip, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())
	backendAddr := tb.backendManager.GetBackendByIP(ip)
	if backendAddr == "" {
		log.Printf("No healthy backend available for client %s", ip)
		return
	}

	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("Failed to connect to backend %s for client %s: %v", backendAddr, ip, err)
		return
	}
	defer backendConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if _, err := io.Copy(backendConn, clientConn); err != nil {
			log.Printf("Error while copying from client to backend: %v", err)
		}
		backendConn.Close()
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(clientConn, backendConn); err != nil {
			log.Printf("Error while copying from backend to client: %v", err)
		}
		clientConn.Close()
	}()

	wg.Wait()
}

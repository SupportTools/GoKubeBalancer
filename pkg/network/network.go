package network

import (
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/supporttools/GoKubeBalancer/pkg/backend"
	"github.com/supporttools/GoKubeBalancer/pkg/logging"
)

var log = logging.SetupLogging()

// TCPBalancer manages TCP connections and routes them to backends
type TCPBalancer struct {
	frontendPort   int // Port to listen for incoming client connections
	backendPort    int // Default port for connecting to the backend servers
	backendManager *backend.BackendManager
}

// NewTCPBalancer creates a new instance of TCPBalancer with a BackendManager
func NewTCPBalancer(frontendPort int, backendPort int, bm *backend.BackendManager) *TCPBalancer {
	return &TCPBalancer{
		frontendPort:   frontendPort,
		backendPort:    backendPort,
		backendManager: bm,
	}
}

// Start listens on the specified frontend port and handles incoming connections
func (tb *TCPBalancer) Start() {
	listenAddr := "0.0.0.0:" + strconv.Itoa(tb.frontendPort) // Listen on the frontend port
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("[TCPBalancer] Failed to listen on port %d: %v", tb.frontendPort, err)
	}
	defer listener.Close()

	log.Printf("[TCPBalancer] TCP Load Balancer started on port %d", tb.frontendPort)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("[TCPBalancer] Failed to accept connection: %v", err)
			continue
		}
		log.Debugf("[TCPBalancer] Accepted new connection from %s", clientConn.RemoteAddr().String())
		go tb.handleConnection(clientConn)
	}
}

// handleConnection manages a single client connection, routing it to an appropriate backend
func (tb *TCPBalancer) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	clientIP, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())

	backendIP := tb.backendManager.GetBackendByIP(clientIP)
	if backendIP == "" {
		log.Printf("[Connection] No healthy backend available for client %s", clientIP)
		return
	}

	backendAddr := backendIP + ":" + strconv.Itoa(tb.backendPort)
	if strings.Contains(backendIP, ":") {
		backendAddr = backendIP // Use the IP:Port directly if it's already formatted
	}

	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("[Connection] Failed to connect to backend %s for client %s: %v", backendAddr, clientIP, err)
		return
	}
	defer backendConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	clientToBackendBytes := make(chan int64)
	backendToClientBytes := make(chan int64)
	errorChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		bytesCopied, err := io.Copy(backendConn, clientConn)
		if err != nil {
			errorChan <- err
			log.Debugf("[Connection] Error while copying from client %s to backend %s: %v", clientIP, backendAddr, err)
		}
		clientToBackendBytes <- bytesCopied
		backendConn.Close()
	}()

	go func() {
		defer wg.Done()
		bytesCopied, err := io.Copy(clientConn, backendConn)
		if err != nil {
			errorChan <- err
			log.Debugf("[Connection] Error while copying from backend %s to client %s: %v", backendAddr, clientIP, err)
		}
		backendToClientBytes <- bytesCopied
		clientConn.Close()
	}()

	clientToBackend := <-clientToBackendBytes
	backendToClient := <-backendToClientBytes

	log.Infof("[Connection] client=%s backend=%s port=%d client-to-backend=%dBytes backend-to-client=%dBytes total=%dBytes",
		clientIP, backendIP, tb.backendPort, clientToBackend, backendToClient, clientToBackend+backendToClient)

	wg.Wait()
}

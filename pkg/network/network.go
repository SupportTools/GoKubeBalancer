package network

import (
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/supporttools/GoKubeBalancer/pkg/backend"
	"github.com/supporttools/GoKubeBalancer/pkg/logging"
)

var log = logging.SetupLogging()

// TCPBalancer manages TCP connections and routes them to backends
type TCPBalancer struct {
	frontendPort   int // Port to listen for incoming client connections
	backendPort    int // Default port for connecting to the backend servers
	backendManager *backend.BackendManager
	pool           map[string][]net.Conn
	poolMutex      sync.Mutex
}

// NewTCPBalancer creates a new instance of TCPBalancer with a BackendManager
func NewTCPBalancer(frontendPort int, backendPort int, bm *backend.BackendManager) *TCPBalancer {
	return &TCPBalancer{
		frontendPort:   frontendPort,
		backendPort:    backendPort,
		backendManager: bm,
		pool:           make(map[string][]net.Conn),
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

func (tb *TCPBalancer) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	clientIP, _, _ := net.SplitHostPort(clientConn.RemoteAddr().String())

	// Set a read timeout to detect if the connection is idle and should be closed
	err := clientConn.SetReadDeadline(time.Now().Add(30 * time.Second)) // Adjust as needed
	if err != nil {
		log.Printf("[Connection] Failed to set read deadline for client %s: %v", clientIP, err)
		return
	}

	// Acquire a connection from the pool or create a new one if not available
	tb.poolMutex.Lock()
	backendIP := tb.backendManager.GetBackendByIP(clientIP)
	if backendIP == "" {
		log.Printf("[Connection] No healthy backend available for client %s", clientIP)
		return
	}

	backendAddr := backendIP + ":" + strconv.Itoa(tb.backendPort)
	if strings.Contains(backendIP, ":") {
		backendAddr = backendIP // Use the IP:Port directly if it's already formatted
	}

	var backendConn net.Conn
	if conn, ok := tb.pool[backendAddr]; ok && len(conn) > 0 {
		backendConn = conn[0]
		tb.pool[backendAddr] = conn[1:] // Remove the used connection from the pool
	} else {
		var err error
		backendConn, err = net.Dial("tcp", backendAddr)
		if err != nil {
			log.Printf("[Connection] Failed to connect to backend %s for client %s: %v", backendAddr, clientIP, err)
			return
		}
	}
	tb.poolMutex.Unlock()

	// Use the acquired or newly created connection for data transfer
	var wg sync.WaitGroup
	wg.Add(2)

	clientToBackendBytes := make(chan int64)
	backendToClientBytes := make(chan int64)

	go tb.copyAndClose(clientConn, backendConn, &wg, clientToBackendBytes)
	go tb.copyAndClose(backendConn, clientConn, &wg, backendToClientBytes)

	clientDataSize := <-clientToBackendBytes
	backendDataSize := <-backendToClientBytes

	log.Debugf("[Connection] Transfered %d bytes from client %s to backend and back", clientDataSize+backendDataSize, clientIP)
}

func (tb *TCPBalancer) copyAndClose(src net.Conn, dst net.Conn, wg *sync.WaitGroup, transferBytes chan<- int64) {
	defer src.Close()
	defer dst.Close()
	bytesWritten, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("[Connection] Failed to copy data between client and backend: %v", err)
	} else {
		transferBytes <- bytesWritten
	}
	wg.Done()
}

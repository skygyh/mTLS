package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	certFile       = "cert.pem"       // Path to the original certificate file
	keyFile        = "key.pem"        // Path to the original private key file
	newCertFile    = "new_cert.pem"   // Path to the new certificate file
	newKeyFile     = "new_key.pem"    // Path to the new private key file
	reloadInterval = 10 * time.Second // Interval for checking file changes
)

// ConnectionState holds the connection state information
type ConnectionState struct {
	tlsConfig *tls.Config
}

// Server holds the server state information
type Server struct {
	mu            sync.RWMutex
	tlsConfig     *tls.Config
	connectionMap map[*http.Conn]struct{}
}

// HandleConnection handles the incoming connections
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	tlsConfig := s.tlsConfig // Use the current TLS configuration
	s.mu.RUnlock()

	tlsConn, err := tls.ListenerForConfig(&tls.Config{
		Certificates:       tlsConfig.Certificates,
		InsecureSkipVerify: true, // Skip verification for simplicity
	})
	if err != nil {
		log.Printf("Failed to create TLS connection: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tlsConn.Close()

	// Handle the connection logic
	// ...
}

// Start starts the server
func (s *Server) Start() {
	http.HandleFunc("/", s.HandleConnection)

	server := &http.Server{
		Addr:    ":443",
		Handler: nil, // Use the default handler
		TLSConfig: &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				// Return the current certificate for all connections
				s.mu.RLock()
				defer s.mu.RUnlock()
				return s.tlsConfig.Certificates[0], nil
			},
		},
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Server started")

	// Wait for the interrupt signal to gracefully shutdown the server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	// Gracefully shutdown the server
	server.Shutdown(nil)
}

// ReloadTLSConfig reloads the TLS configuration
func (s *Server) ReloadTLSConfig() error {
	cert, err := tls.LoadX509KeyPair(newCertFile, newKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load new certificate: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tlsConfig.Certificates = []tls.Certificate{cert}

	log.Println("TLS configuration reloaded successfully")

	return nil
}

// MonitorFiles continuously monitors the certificate files for changes
func (s *Server) MonitorFiles() {
	for {
		select {
		case <-time.After(reloadInterval):
			// Check if the certificate files have changed
			if hasChanged(certFile, keyFile) {
				log.Println("Certificate files have changed. Reloading TLS configuration...")

				err := s.ReloadTLSConfig()
				if err != nil {
					log.Printf("Failed to reload TLS configuration: %v", err)
				}
			}
		}
	}
}

// hasChanged checks if the file has changed
func hasChanged(filename ...string) bool {
	// Implement your file change detection logic here
	// ...
	return false
}

func main() {
	// Create a server instance
	server := &Server{
		tlsConfig:     &tls.Config{},
		connectionMap: make(map[*http.Conn]struct{}),
	}

	// Load the initial certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load initial certificate: %v", err)
	}
	server.tlsConfig.Certificates = []tls.Certificate{cert}

	// Start the server
	go server.Start()

	// Start monitoring the certificate files
	go server.MonitorFiles()

	// Wait for the interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

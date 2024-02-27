package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const ca_cert_dir string = "./certs.bk"

func main() {

	go waitForShutdown()

	name := flag.String("c", "a", "client name")
	flag.Parse()

	clientCaCert := fmt.Sprintf("%s/ca.crt", ca_cert_dir)
	log.Println("Load CA- ", clientCaCert)
	cert, err := ioutil.ReadFile(clientCaCert)
	if err != nil {
		log.Fatalf("could not open certificate file: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)

	clientCert := fmt.Sprintf("%s/client.%s.crt", ca_cert_dir, *name)
	clientKey := fmt.Sprintf("%s/client.%s.key", ca_cert_dir, *name)
	log.Println("Load key pairs - ", clientCert, clientKey)
	certificate, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatalf("could not load certificate: %v", err)
	}

	client := http.Client{
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
				GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
					// Always get latest localhost.crt and localhost.key
					// ex: keeping certificates file somewhere in global location where created certificates updated and this closure function can refer that
					log.Printf("GetCertificate reloading")
					const ca_cert_dir string = "./certs"
					clientCert := fmt.Sprintf("%s/client.%s.crt", ca_cert_dir, *name)
					clientKey := fmt.Sprintf("%s/client.%s.key", ca_cert_dir, *name)
					log.Println("Load key pairs - ", clientCert, clientKey)

					cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
					if err != nil {
						return nil, err
					}
					return &cert, nil
				},
				Certificates: []tls.Certificate{certificate},
			},
		},
	}

	// Request /hello over port 8443 via the GET method
	// Using curl the verfiy it :
	// curl --trace trace.log -k \
	//   --cacert ./ca.crt  --cert ./client.b.crt --key ./client.b.key  \
	//     https://localhost:8443/hello

	for {

		r, err := client.Get("https://localhost:9443/hello")
		if err != nil {
			log.Printf("error making get request: %v", err)
		}

		if tlsErr, ok := err.(tls.RecordHeaderError); ok {
			fmt.Println("TLS handshake error, reloading ca root:", tlsErr)
			const ca_cert_dir1 string = "./certs"

			clientCaCert := fmt.Sprintf("%s/ca.crt", ca_cert_dir1)
			log.Println("Load CA- ", clientCaCert)
			cert, err := ioutil.ReadFile(clientCaCert)
			if err != nil {
				log.Fatalf("could not open certificate file: %v", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(cert)

			client.Transport.(*http.Transport).TLSClientConfig.RootCAs = caCertPool

			continue
		}

		// Read the response body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("error reading response: %v", err)
		}

		// Print the response body to stdout
		fmt.Printf("%s\n", body)
		r.Body.Close()
		time.Sleep(5 * time.Second)
	}

}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	// Perform cleanup and shutdown tasks here

	os.Exit(0)
}

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const ca_cert_dir string = "./certs.bk"

func main() {

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
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{certificate},
			},
		},
	}

	// Request /hello over port 8443 via the GET method
	// Using curl the verfiy it :
	// curl --trace trace.log -k \
	//   --cacert ./ca.crt  --cert ./client.b.crt --key ./client.b.key  \
	//     https://localhost:8443/hello

	r, err := client.Get("https://localhost:9443/hello")
	if err != nil {
		log.Fatalf("error making get request: %v", err)
	}

	// Read the response body
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}

	// Print the response body to stdout
	fmt.Printf("%s\n", body)
}

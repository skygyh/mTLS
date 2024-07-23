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

const CertsDir string = "/etc/tls"

func main() {

	address := flag.String("link", "", "link address")
	flag.Parse()

	// load CA certificate file and add it to list of client CAs
	caCertFile, err := ioutil.ReadFile(CertsDir + "/ca.crt")
	if err != nil {
		log.Fatalf("error reading CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertFile)

	serverCert, err := tls.LoadX509KeyPair(CertsDir+"/tls.crt", CertsDir+"/tls.key")
	if err != nil {
		log.Fatalf("error loading server certificate and key: %v", err)
	}

	client := http.Client{
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{serverCert},
			},
		},
	}

	// Request /hello over port 8443 via the GET method
	// Using curl the verfiy it :
	// curl --trace trace.log -k \
	//   --cacert ./ca.crt  --cert ./client.b.crt --key ./client.b.key  \
	//     https://localhost:8443/hello

	r, err := client.Get(*address)
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

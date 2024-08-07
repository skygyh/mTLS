package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

func printHeader(r *http.Request) {
	log.Print(">>>>>>>>>>>>>>>> Header <<<<<<<<<<<<<<<<")
	// Loop over header names
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			log.Printf("%v:%v", name, value)
		}
	}
}

func printConnState(state *tls.ConnectionState) {
	log.Print(">>>>>>>>>>>>>>>> State <<<<<<<<<<<<<<<<")

	log.Printf("Version: %x", state.Version)
	log.Printf("HandshakeComplete: %t", state.HandshakeComplete)
	log.Printf("DidResume: %t", state.DidResume)
	log.Printf("CipherSuite: %x", state.CipherSuite)
	log.Printf("NegotiatedProtocol: %s", state.NegotiatedProtocol)
	log.Printf("NegotiatedProtocolIsMutual: %t", state.NegotiatedProtocolIsMutual)

	log.Print("Certificate chain:")
	for i, cert := range state.PeerCertificates {
		subject := cert.Subject
		issuer := cert.Issuer
		log.Printf(" %d s:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", i, subject.Country, subject.Province, subject.Locality, subject.Organization, subject.OrganizationalUnit, subject.CommonName)
		log.Printf("   i:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", issuer.Country, issuer.Province, issuer.Locality, issuer.Organization, issuer.OrganizationalUnit, issuer.CommonName)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	printHeader(r)
	if r.TLS != nil {
		printConnState(r.TLS)
	}
	log.Print(">>>>>>>>>>>>>>>>> End <<<<<<<<<<<<<<<<<<")
	fmt.Println("")
	// Write "Hello, world!" to the response body
	io.WriteString(w, "Hello, world!\n")
}

const ca_cert string = "./certs/ca.crt"
const server_cert string = "./certs/server.crt"
const server_key string = "./certs/server.key"

func main() {
	port := 9080
	sslPort := 9443

	// Set up a /hello resource handler
	handler := http.NewServeMux()
	handler.HandleFunc("/hello", helloHandler)

	// Listen to port 8080 and wait
	go func() {

		server := http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler,
		}
		fmt.Printf("(HTTP) Listen on :%d\n", port)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("(HTTP) error listening to port: %v", err)
		}
	}()

	//	newCert, err := tls.LoadX509KeyPair(server_cert, server_key)
	//	if err != nil {
	//		log.Fatalf("error reading server certificate: %v", err)
	//	}

	// load CA certificate file and add it to list of client CAs
	caCertFile, err := ioutil.ReadFile(ca_cert)
	if err != nil {
		log.Fatalf("error reading CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertFile)

	//	caCertFilebk, err := ioutil.ReadFile("./certs.bk/ca.crt")
	//	if err != nil {
	//		log.Fatalf("error reading CA certificate: %v", err)
	//	}
	//	caCertPool.AppendCertsFromPEM(caCertFilebk)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			// Always get latest localhost.crt and localhost.key
			// ex: keeping certificates file somewhere in global location where created certificates updated and this closure function can refer that
			log.Printf("tlsconfig reloading")
			caCertFile, err := ioutil.ReadFile(ca_cert)
			if err != nil {
				log.Fatalf("error reading CA certificate: %v", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCertFile)

			tls_config := &tls.Config{
				ClientCAs: caCertPool,
				GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
					// Always get latest localhost.crt and localhost.key
					// ex: keeping certificates file somewhere in global location where created certificates updated and this closure function can refer that
					log.Printf("GetCertificate reloading")
					cert, err := tls.LoadX509KeyPair(server_cert, server_key)
					if err != nil {
						return nil, err
					}
					return &cert, nil
				},
				ClientAuth:               tls.RequireAndVerifyClientCert,
				MinVersion:               tls.VersionTLS13,
				CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				},
			}

			if err != nil {
				return nil, err
			}
			return tls_config, nil
		},
		//GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		//	return &newCert, nil
		//},
		ClientAuth:               tls.RequireAndVerifyClientCert,
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	tlsConfig.BuildNameToCertificate()

	// serve on port 8443 of local host
	server := http.Server{
		Addr:      fmt.Sprintf(":%d", sslPort),
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	//go reload_TLSConfig(server.TLSConfig)

	fmt.Printf("(HTTPS) Listen on :%d\n", sslPort)

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("(HTTP) error listening to port: %v", err)
	}

	tlsListener := tls.NewListener(ln, server.TLSConfig)

	if err := server.Serve(tlsListener); err != nil {
		log.Fatalf("(HTTPS) error listening to port: %v", err)
	}

}

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"

	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/dtls"
)

var CERT_NAME = "certs/client_cert.pem"
var KEY_NAME = "certs/client_key.pem"
var ROOT_CA = "certs/root_ca_cert.pem"

func main() {
	config, err := createClientConfig(context.Background())
	if err != nil {
		log.Fatalln(err)
		return
	}
	co, err := dtls.Dial("localhost:5688", config)
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	path := "/a"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := co.Get(ctx, path)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	log.Printf("Response payload: %+v", resp)
}

func createClientConfig(ctx context.Context) (*piondtls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(CERT_NAME, KEY_NAME)
	if err != nil {
		log.Fatalf("Error loading server key pair: %v", err)
	}
	rootBytes, err := os.ReadFile(ROOT_CA)
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(rootBytes) {
		log.Fatalf("Failed to append CA certificate to pool")
	}

	return &piondtls.Config{
		Certificates:         []tls.Certificate{certificate},
		ExtendedMasterSecret: piondtls.RequireExtendedMasterSecret,
		RootCAs:              certPool,
		InsecureSkipVerify:   false,
	}, nil
}

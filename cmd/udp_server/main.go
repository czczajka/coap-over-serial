package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/dtls"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp/client"
)

var CERT_NAME = "certs/server_cert.pem"
var KEY_NAME = "certs/server_key.pem"

var ROOT_CA = "certs/root_ca_cert.pem"

// var ROOT_CA = "certs/rsa/myCA.crt"

func onNewConn(cc *client.Conn) {
	dtlsConn, ok := cc.NetConn().(*piondtls.Conn)
	if !ok {
		log.Fatalf("invalid type %T", cc.NetConn())
	}
	clientCert, err := x509.ParseCertificate(dtlsConn.ConnectionState().PeerCertificates[0])
	if err != nil {
		log.Fatal(err)
	}
	cc.SetContextValue("client-cert", clientCert)
	cc.AddOnClose(func() {
		log.Println("closed connection")
	})
}

func toHexInt(n *big.Int) string {
	return fmt.Sprintf("%x", n) // or %X or upper case
}

func handleA(w mux.ResponseWriter, r *mux.Message) {
	clientCert := r.Context().Value("client-cert").(*x509.Certificate)
	log.Println("Serial number:", toHexInt(clientCert.SerialNumber))
	log.Println("Subject:", clientCert.Subject)
	log.Println("Email:", clientCert.EmailAddresses)

	log.Printf("got message in handleA:  %+v from %v\n", r, w.Conn().RemoteAddr())
	err := w.SetResponse(codes.GET, message.TextPlain, bytes.NewReader([]byte("A hello world")))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func main() {
	m := mux.NewRouter()
	m.Handle("/a", mux.HandlerFunc(handleA))

	config, err := createServerConfig(context.Background())
	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Fatal(listenAndServeDTLS("udp", ":5688", config, m))
}

func listenAndServeDTLS(network string, addr string, config *piondtls.Config, handler mux.Handler) error {
	l, err := net.NewDTLSListener(network, addr, config)
	if err != nil {
		return err
	}
	defer l.Close()
	s := dtls.NewServer(options.WithMux(handler), options.WithOnNewConn(onNewConn))
	return s.Serve(l)
}

func createServerConfig(ctx context.Context) (*piondtls.Config, error) {
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
		ClientCAs:            certPool,
		ClientAuth:           piondtls.RequireAndVerifyClientCert,
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(ctx, 30*time.Second)
		},
	}, nil
}

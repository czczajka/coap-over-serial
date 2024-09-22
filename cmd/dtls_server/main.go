package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"

	"github.com/czczajka/enrollment_app/common"
	"github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp/coder"
	"github.com/tarm/serial"
)

var CERT_NAME = "certs/server_cert.pem"
var KEY_NAME = "certs/server_key.pem"
var ROOT_CA = "certs/root_ca_cert.pem"

// RSA certificates
// var CERT_NAME = "certs/rsa/server.crt"
// var KEY_NAME = "certs/rsa/server.key"
// var ROOT_CA = "certs/rsa/myCA.crt"

type HandlerFunc func(conn *dtls.Conn, req *pool.Message)

type Router struct {
	routes map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]HandlerFunc),
	}
}

func (r *Router) Handle(path string, handler HandlerFunc) {
	r.routes[path] = handler
}

func (r *Router) ServeCOAP(conn *dtls.Conn, req *pool.Message) {
	path, err := req.Path()
	if err != nil {
		log.Printf("Error getting path: %v", err)
		return
	}

	handler, ok := r.routes[path]
	if !ok {
		log.Printf("No handler for path: %v", path)
		return
	}

	// Pass the DTLS connection to the handler
	handler(conn, req)
}

func main() {
	log.Printf("Starting CoAP server over dtls tutorial\n")

	// Set up the serial port
	c := &serial.Config{Name: "/dev/ttyGS0", Baud: 115200}
	serialConn, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("Error opening serial port: %v", err)
	}
	defer serialConn.Close()

	log.Println("Serial port opened successfully")

	config, err := createServerConfig(context.Background())
	if err != nil {
		log.Fatalln(err)
		return
	}

	// Create a DTLS server using the custom SerialPacketConn
	log.Println("Attempting to create DTLS server")
	dtlsConn, err := dtls.Server(common.NewSerialPacketConn(serialConn), nil, config)
	if err != nil {
		log.Fatalf("Error setting up DTLS server: %v", err)
	}
	defer dtlsConn.Close()

	log.Println("DTLS server set up successfully")

	router := NewRouter()
	router.Handle("/a", handleRequest)

	for {
		// Buffer to read incoming data
		buf := make([]byte, common.SERIAL_BUFFER_SIZE)

		// Read from DTLS connection
		n, err := dtlsConn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from DTLS connection: %v", err)
		}

		log.Printf("Server: received %d bytes: %x", n, buf[:n])

		// Decode the incoming CoAP message
		req := pool.NewMessage(context.Background())
		len, err := req.UnmarshalWithDecoder(coder.DefaultCoder, buf[:n])
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue // Handle the error gracefully
		}
		if len == 0 {
			log.Println("Failed to unmarshal message, len = 0")
			continue // Handle the error gracefully
		}

		log.Printf("Received CoAP request for path")

		router.ServeCOAP(dtlsConn, req)
	}
}

func handleRequest(conn *dtls.Conn, req *pool.Message) {
	// Prepare response
	resp := pool.NewMessage(context.Background())

	resp.SetCode(codes.Content)
	resp.SetMessageID(req.MessageID())
	resp.SetBody(bytes.NewReader([]byte("Hello World")))
	resp.SetType(message.Acknowledgement)

	// Encode the response message
	data, err := resp.MarshalWithEncoder(coder.DefaultCoder)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	// Send the response over serial
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Error sending response over serial: %v", err)
	}
}

func createServerConfig(ctx context.Context) (*dtls.Config, error) {
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

	return &dtls.Config{
		Certificates:         []tls.Certificate{certificate},
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		ClientCAs:            certPool,
		ClientAuth:           dtls.RequireAndVerifyClientCert,
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(ctx, 30*time.Second)
		},
		MTU: common.MTU,
	}, nil
}

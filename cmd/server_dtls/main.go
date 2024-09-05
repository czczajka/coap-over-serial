package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"log"
	"math/big"
	"time"

	"github.com/czczajka/enrollment_app/common"
	"github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp/coder"
	"github.com/tarm/serial"
)

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

	// DTLS Configuration
	// Updated DTLS Configuration on Server
	config := &dtls.Config{
		Certificates:         []tls.Certificate{generateSelfSignedCert()},
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		ClientAuth:           dtls.NoClientCert,
		FlightInterval:       time.Second * 5, // Increase retransmission interval
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(context.Background(), time.Minute*1)
		},
		InsecureSkipVerify: true, // For testing purposes; disable certificate verification.
		MTU:                common.MTU,
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

		log.Printf("Received %d bytes: %x", n, buf[:n])

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

// generateSelfSignedCert generates a self-signed certificate for DTLS.
func generateSelfSignedCert() tls.Certificate {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("failed to create certificate: %v", err)
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}
}

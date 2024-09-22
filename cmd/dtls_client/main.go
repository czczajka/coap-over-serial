package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"os"

	"github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp/coder"
	"github.com/tarm/serial"

	"github.com/czczajka/enrollment_app/common"
)

var SERIAL_PATH = "/dev/tty.usbmodem1201"

var CERT_NAME = "certs/client_cert.pem"
var KEY_NAME = "certs/client_key.pem"
var ROOT_CA = "certs/root_ca_cert.pem"

// RSA certificates
// var CERT_NAME = "certs/rsa/client.crt"
// var KEY_NAME = "certs/rsa/client.key"
// var ROOT_CA = "certs/rsa/myCA.crt"

func main() {
	log.Println("Starting CoAP client over dtls tutorial")

	// Set up the serial port connection
	c := &serial.Config{Name: SERIAL_PATH, Baud: 115200} // Change '/dev/tty.usbmodem1201' to the correct serial port for your setup.
	serialConn, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("Error opening serial port: %v", err)
	}
	defer serialConn.Close()

	log.Println("Serial port opened successfully")

	config, err := createClientConfig(context.Background())
	if err != nil {
		log.Fatalln(err)
		return
	}

	addr := &net.UDPAddr{IP: net.IPv4zero, Port: 0} // Dummy address
	dtlsConn, err := dtls.Client(common.NewSerialPacketConn(serialConn), addr, config)
	if err != nil {
		log.Fatalf("Error setting up DTLS client: %v", err)
	}
	defer func() {
		if dtlsConn != nil {
			dtlsConn.Close()
		}
	}()

	log.Println("DTLS handshake completed successfully")

	// Prepare a CoAP GET request message
	req := pool.NewMessage(context.Background())
	req.SetCode(codes.GET)
	req.SetPath("/a")
	req.SetMessageID(1234)
	req.SetType(message.Confirmable)

	// Encode the CoAP message
	data, err := req.MarshalWithEncoder(coder.DefaultCoder)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}
	log.Println("CoAP message marshaled successfully")

	// Send the CoAP message to the server over the DTLS connection
	_, err = dtlsConn.Write(data)
	if err != nil {
		log.Fatalf("Failed to send CoAP message: %v", err)
	}
	log.Println("CoAP message sent successfully")

	// Buffer to read the server's response
	buf := make([]byte, common.SERIAL_BUFFER_SIZE)
	n, err := dtlsConn.Read(buf)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// Decode the server's response
	resp := pool.NewMessage(context.Background())
	_, err = resp.UnmarshalWithDecoder(coder.DefaultCoder, buf[:n])
	if err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Print the response
	log.Printf("Response Code: %v", resp.Code())
	body, _ := resp.ReadBody()
	log.Printf("Response Body: %s", string(body))
}

func createClientConfig(ctx context.Context) (*dtls.Config, error) {
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
		RootCAs:              certPool,
		InsecureSkipVerify:   false,
		MTU:                  common.MTU,
	}, nil
}

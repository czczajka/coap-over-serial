package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp/coder"
	"github.com/tarm/serial"

	"github.com/czczajka/enrollment_app/common"
)

func main() {
	log.Println("Starting CoAP client over dtls tutorial")

	// Set up the serial port connection
	c := &serial.Config{Name: "/dev/tty.usbmodem1201", Baud: 115200} // Change '/dev/tty.usbmodem1201' to the correct serial port for your setup.
	serialConn, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("Error opening serial port: %v", err)
	}
	defer serialConn.Close()

	log.Println("Serial port opened successfully")

	// DTLS Configuration
	config := &dtls.Config{
		Certificates:         []tls.Certificate{generateSelfSignedCert()},
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		ClientAuth:           dtls.NoClientCert,
		FlightInterval:       time.Second * 5, // Increase retransmission interval
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(context.Background(), time.Minute*1)
		},
		// InsecureSkipVerify: true, // For testing purposes; disable certificate verification.
		MTU: common.MTU,
	}

	// Create a DTLS client using the custom SerialPacketConn
	log.Println("Attempting to create DTLS client")
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: 0} // Dummy address
	dtlsConn, err := dtls.Client(common.NewSerialPacketConn(serialConn), addr, config)
	log.Println("DTLS client created")
	if err != nil {
		log.Fatalf("Error setting up DTLS client: %v", err)
	}
	defer dtlsConn.Close()

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
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
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

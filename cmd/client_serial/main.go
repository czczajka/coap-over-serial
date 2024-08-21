package main

import (
	"context"
	"fmt"
	"log"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp/coder"
	"github.com/tarm/serial"
)

func main() {
	// Configure the serial port
	c := &serial.Config{Name: "/dev/tty.usbmodem1201", Baud: 9600} // Adjust the port as needed
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	req := pool.NewMessage(context.Background())
	req.SetCode(codes.GET)
	err = req.SetPath("/a")
	if err != nil {
		log.Fatal(err)
	}
	req.SetMessageID(1234)
	req.SetType(message.Confirmable)

	data, err := req.MarshalWithEncoder(coder.DefaultCoder)
	if err != nil {
		log.Fatal("Failed to marshal message:", err)
	}

	// Send the CoAP message to the server over the serial port
	_, err = s.Write(data)
	if err != nil {
		log.Fatal("Failed to send CoAP message:", err)
	}

	// Read the server's response
	readBuf := make([]byte, 1024)
	n, err := s.Read(readBuf)
	if err != nil {
		log.Fatal("Failed to read from serial port:", err)
	}

	msg := pool.NewMessage(context.Background())
	len, err := msg.UnmarshalWithDecoder(coder.DefaultCoder, readBuf[:n])
	if err != nil {
		log.Fatal("Failed to unmarshal message:", err)
	}
	if len == 0 {
		log.Fatal("Failed to unmarshal message, len =0")
	}

	body, err := msg.ReadBody()
	if err != nil {
		log.Fatal("Failed to read body:", err)
	}
	// Print the server's response
	fmt.Printf("Received response: code: %d, body: %s", msg.Code(), body)
}

package main

import (
	"bytes"
	"context"
	"log"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp/coder" // Using UDP coder for serial
	"github.com/tarm/serial"
)

type HandlerFunc func(conn *serial.Port, req *pool.Message)

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

func (r *Router) ServeCOAP(conn *serial.Port, req *pool.Message) {
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

	handler(conn, req)
}

func main() {
	// Set up the serial port
	c := &serial.Config{Name: "/dev/ttyGS0", Baud: 9600}
	serialConn, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("Error opening serial port: %v", err)
	}
	defer serialConn.Close()

	router := NewRouter()
	router.Handle("/a", handleRequest)

	for {
		// Buffer to read incoming data
		buf := make([]byte, 1024)

		// Read from serial port
		n, err := serialConn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from serial port: %v", err)
		}

		// Decode the incoming CoAP message
		req := pool.NewMessage(context.Background())
		len, err := req.UnmarshalWithDecoder(coder.DefaultCoder, buf[:n])
		if err != nil {
			log.Fatal("Failed to unmarshal message:", err)
		}
		if len == 0 {
			log.Fatal("Failed to unmarshal message, len =0")
		}

		router.ServeCOAP(serialConn, req)
	}
}

func handleRequest(conn *serial.Port, req *pool.Message) {
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

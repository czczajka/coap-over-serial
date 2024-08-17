package main

import (
	"bytes"
	"log"

	coap "github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

func handleA(w mux.ResponseWriter, req *mux.Message) {
	err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("hello world")))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func main() {
	log.Printf("Starting CoAP server tutorial\n")

	r := mux.NewRouter()
	r.Handle("/a", mux.HandlerFunc(handleA))

	log.Fatal(coap.ListenAndServe("udp", ":5688", r))
}

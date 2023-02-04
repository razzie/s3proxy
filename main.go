package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/razzie/s3proxy/pkg/s3proxy"
)

func main() {
	var addr, endpoint, encryptionKey string
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&endpoint, "endpoint", "", "Remote S3 endpoint URL")
	flag.StringVar(&encryptionKey, "encryption-key", "", "Optional key for transparent (weak) encryption")
	flag.Parse()
	srv, err := s3proxy.NewProxy(endpoint, encryptionKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting s3proxy on", addr)
	http.ListenAndServe(addr, srv)
}

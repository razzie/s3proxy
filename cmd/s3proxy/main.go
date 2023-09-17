package main

import (
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/razzie/s3proxy/pkg/s3proxy"
)

func main() {
	var addr, endpoint, encryptionKey string
	var logging bool
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&endpoint, "endpoint", "", "Remote S3 endpoint URL")
	flag.StringVar(&encryptionKey, "encryption-key", "", "Optional key for transparent (weak) encryption")
	flag.BoolVar(&logging, "logging", false, "Enable logging")
	flag.Parse()
	logger := log.Default()
	if !logging {
		logger = log.New(io.Discard, "", log.LstdFlags)
	}
	srv, err := s3proxy.NewProxy(endpoint, encryptionKey, logger)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting s3proxy on", addr)
	http.ListenAndServe(addr, srv)
}

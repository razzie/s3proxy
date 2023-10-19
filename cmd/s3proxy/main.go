package main

import (
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/razzie/s3proxy/pkg/s3proxy"
)

func main() {
	var addr, endpoint, encryptionKey, accessKey, secretKey string
	var logging bool
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&endpoint, "endpoint", "", "Remote S3 endpoint URL")
	flag.StringVar(&encryptionKey, "encryption-key", "", "Optional key for weak encryption")
	flag.BoolVar(&logging, "logging", false, "Enable logging")
	flag.StringVar(&accessKey, "access-key", "", "S3 access key")
	flag.StringVar(&secretKey, "secret-key", "", "S3 secret key")
	flag.Parse()
	logger := log.Default()
	if !logging {
		logger = log.New(io.Discard, "", log.LstdFlags)
	}
	o := &s3proxy.ProxyOptions{
		Endpoint:      endpoint,
		EncryptionKey: encryptionKey,
		Logger:        logger,
		AccessKeySecretMap: map[string]string{
			accessKey: secretKey,
		},
	}
	srv, err := s3proxy.NewProxy(o)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting s3proxy on", addr)
	http.ListenAndServe(addr, srv)
}

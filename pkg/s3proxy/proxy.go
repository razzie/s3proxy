package s3proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	Endpoint *url.URL
	LT       *LookupTable
}

func NewProxy(endpoint, encryptionKey string) (*Proxy, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	lt := NewLookupTable(encryptionKey)
	return &Proxy{
		Endpoint: u,
		LT:       lt,
	}, nil
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(p.Endpoint)
	r.Host = p.Endpoint.Host
	if r.Method == "PUT" {
		r.Body = &Reader{R: r.Body, LT: p.LT}
	} else if r.Method == "GET" && !strings.HasSuffix(r.URL.Path, "/") {
		w = &Writer{W: w, LT: p.LT}
	}
	proxy.ServeHTTP(w, r)
}

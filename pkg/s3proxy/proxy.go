package s3proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	endpoint *url.URL
	lt       *LookupTable
	proxy    *httputil.ReverseProxy
}

func NewProxy(endpoint, encryptionKey string) (*Proxy, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("missing endpoint")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	lt := NewLookupTable(encryptionKey)
	p := httputil.NewSingleHostReverseProxy(u)
	if len(encryptionKey) > 0 {
		origDirector := p.Director
		p.Director = func(r *http.Request) {
			origDirector(r)
			r.Header.Del("Content-MD5")
		}
		p.ModifyResponse = func(r *http.Response) error {
			r.Header.Del("Content-MD5")
			return nil
		}
	}
	return &Proxy{
		endpoint: u,
		lt:       lt,
		proxy:    p,
	}, nil
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Host = p.endpoint.Host
	if r.Method == "PUT" {
		r.Body = &Reader{R: r.Body, LT: p.lt}
	} else if r.Method == "GET" && !strings.HasSuffix(r.URL.Path, "/") {
		w = &Writer{W: w, LT: p.lt}
	}
	p.proxy.ServeHTTP(w, r)
}

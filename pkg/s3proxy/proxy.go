package s3proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	lt    *LookupTable
	proxy *httputil.ReverseProxy
}

func NewProxy(endpoint, encryptionKey string) (*Proxy, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("missing endpoint")
	}
	if !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://") {
		endpoint = "https://" + endpoint
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
			r.Host = u.Host
			r.Header.Del("Content-MD5")
		}
		p.ModifyResponse = func(r *http.Response) error {
			r.Header.Del("Content-MD5")
			return nil
		}
	}
	return &Proxy{
		lt:    lt,
		proxy: p,
	}, nil
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		r.Body = &ReadCloser{R: r.Body, F: p.lt.Encrypt}
	} else if r.Method == "GET" && !strings.HasSuffix(r.URL.Path, "/") {
		w = &ResponseWriter{W: w, F: p.lt.Decrypt}
	}
	p.proxy.ServeHTTP(w, r)
}

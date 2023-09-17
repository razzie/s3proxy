package s3proxy

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	lt     *LookupTable
	proxy  *httputil.ReverseProxy
	logger *log.Logger
}

func NewProxy(endpoint, encryptionKey string, logger *log.Logger) (*Proxy, error) {
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
	if logger == nil {
		logger = log.Default()
	}
	return &Proxy{
		lt:     lt,
		proxy:  p,
		logger: logger,
	}, nil
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		r.Body = &ReadCloser{R: r.Body, F: p.lt.Encrypt}
	} else if r.Method == "GET" && !strings.HasSuffix(r.URL.Path, "/") && !r.URL.Query().Has("uploadId") {
		w = &ResponseWriter{W: w, F: p.lt.Decrypt}
	}
	method := r.Method
	path := r.URL.RequestURI()
	ww := &responseWriterWrapper{w: w}
	p.proxy.ServeHTTP(ww, r)
	p.logger.Println(method, path, "->", ww.statusCode)
}

type responseWriterWrapper struct {
	w          http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriterWrapper) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.w.WriteHeader(statusCode)
}

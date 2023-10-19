package s3proxy

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type Proxy struct {
	lt     *LookupTable
	proxy  *httputil.ReverseProxy
	logger *log.Logger
}

type ProxyOptions struct {
	Endpoint           string
	EncryptionKey      string
	Logger             *log.Logger
	AccessKeySecretMap map[string]string
}

func NewProxy(o *ProxyOptions) (*Proxy, error) {
	endpoint := o.Endpoint
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
	logger := o.Logger
	if logger == nil {
		logger = log.Default()
	}
	lt := NewLookupTable(o.EncryptionKey)
	p := httputil.NewSingleHostReverseProxy(u)
	if len(o.EncryptionKey) > 0 {
		origDirector := p.Director
		p.Director = func(r *http.Request) {
			origDirector(r)
			accessKey, _ := GetAWSAccessKey(r)
			secretKey := o.AccessKeySecretMap[accessKey]
			r.Host = u.Host
			r.Header.Del("Content-MD5")
			r.Header.Del("x-amz-meta-sha256")
			updateAuthHeader(r, accessKey, secretKey)
			//dump, _ := httputil.DumpRequest(r, r.Header.Get("Content-Type") == "application/xml" || r.Method == "POST")
			//logger.Println("REQUEST", string(dump))
		}
		p.ModifyResponse = func(r *http.Response) error {
			r.Header.Del("Content-MD5")
			r.Header.Del("x-amz-meta-sha256")
			if r.StatusCode == 403 {
				dump, _ := httputil.DumpResponse(r, true)
				logger.Println("ERROR", string(dump))
				accessKey, _ := GetAWSAccessKey(r.Request)
				secretKey := o.AccessKeySecretMap[accessKey]
				stringToSign, _ := sigV2(r.Request, accessKey, secretKey)
				logger.Println("SIGNATURE", stringToSign)
			}
			//dump, _ := httputil.DumpResponse(r, r.Header.Get("Content-Type") == "application/xml" || r.Request.Method == "POST")
			//logger.Println("RESPONSE", string(dump))
			return nil
		}
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
	p.logger.Println("LOG", method, path, "->", ww.statusCode)
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

func sigV2(r *http.Request, accessKey, secretKey string) (string, string) {
	contentMD5 := r.Header.Get("Content-MD5")
	contentType := r.Header.Get("Content-Type")
	resource := getCanonicalizedResource(r)
	return CalculateSignatureV2(accessKey, secretKey, r.Method, contentType, contentMD5, "", resource, r.Header)
}

func updateAuthHeader(r *http.Request, accessKey, secretKey string) {
	date := time.Now().UTC().Format(http.TimeFormat)
	r.Header.Del("Date")
	r.Header.Set("X-Amz-Date", date)
	_, signature := sigV2(r, accessKey, secretKey)
	r.Header.Set("Authorization", "AWS "+accessKey+":"+signature)
}

func getCanonicalizedResource(r *http.Request) string {
	if r.Method == "GET" && !r.URL.Query().Has("uploads") && !r.URL.Query().Has("uploadId") {
		return r.URL.EscapedPath()
	}
	return strings.TrimSuffix(r.URL.RequestURI(), "=")
}

package s3proxy

import (
	"net/http"
)

type Writer struct {
	W  http.ResponseWriter
	LT *LookupTable
}

func (w *Writer) Header() http.Header {
	return w.W.Header()
}

func (w *Writer) Write(p []byte) (int, error) {
	w.LT.Encrypt(p)
	n, err := w.W.Write(p)
	return n, err
}

func (w *Writer) WriteHeader(statusCode int) {
	w.W.WriteHeader(statusCode)
}

package s3proxy

import (
	"io"
	"net/http"
)

type Writer struct {
	W io.Writer
	F func([]byte)
}

func (w *Writer) Write(p []byte) (int, error) {
	pp := append([]byte(nil), p...)
	w.F(pp)
	n, err := w.W.Write(pp)
	return n, err
}

type ResponseWriter struct {
	W http.ResponseWriter
	F func([]byte)
}

func (w *ResponseWriter) Header() http.Header {
	return w.W.Header()
}

func (w *ResponseWriter) Write(p []byte) (int, error) {
	pp := append([]byte(nil), p...)
	w.F(pp)
	n, err := w.W.Write(pp)
	return n, err
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.W.WriteHeader(statusCode)
}

package s3proxy

import (
	"io"
)

type Reader struct {
	R io.Reader
	F func([]byte)
}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.R.Read(p)
	r.F(p)
	return n, err
}

type ReadCloser struct {
	R io.ReadCloser
	F func([]byte)
}

func (r *ReadCloser) Read(p []byte) (int, error) {
	n, err := r.R.Read(p)
	r.F(p)
	return n, err
}

func (r *ReadCloser) Close() error {
	return r.R.Close()
}

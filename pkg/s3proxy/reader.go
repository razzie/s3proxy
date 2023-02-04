package s3proxy

import (
	"io"
)

type Reader struct {
	R  io.ReadCloser
	LT *LookupTable
}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.R.Read(p)
	r.LT.Decrypt(p)
	return n, err
}

func (r *Reader) Close() error {
	return r.R.Close()
}

package stream

import (
	"io"
)

// Share the amount of bytes read.
type CountReader struct {
	N uint64
	r io.Reader
}

func (s *CountReader) Read(b []byte) (n int, err error) {
	n, e := s.r.Read(b)
	s.N += uint64(n)
	return n, e
}

// TODO: racing?
func (s *CountReader) ReadReset() uint64 {
	n := s.N
	s.N = 0
	return n
}

func NewCountReader(r io.Reader) *CountReader {
	return &CountReader{
		r: r,
	}
}
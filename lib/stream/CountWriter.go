package stream

import (
	"io"
)

type CountWriter struct {
	w io.Writer
	n int64
}

func (c *CountWriter) Write(p []byte) (int, error) {
	n, e := c.w.Write(p)
	c.n += int64(n)
	return n, e
}

func (c *CountWriter) WriteString(p string) (int, error) {
	n, e := c.w.Write([]byte(p))
	c.n += int64(n)
	return n, e
}

func (c *CountWriter) Written() int64 {
	return c.n
}

func (c *CountWriter) ResetWritten() {
	c.n = 0
}

func NewCountWriter(w io.Writer) *CountWriter {
	return &CountWriter{w: w}
}
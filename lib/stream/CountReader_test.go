package stream

import (
	"strings"
	"testing"
	"io"
	"io/ioutil"
)

func TestCountReader(t *testing.T) {
	msg := `Hello world!`
	r := strings.NewReader(msg)
	n, e := io.Copy(ioutil.Discard, r)
	if e != nil {
		t.Fatal(e)
	}
	if n != int64(len(msg)) {
		t.Fatalf("Byte count mismatch, expect=%d found=%d", len(msg), n)
	}
}

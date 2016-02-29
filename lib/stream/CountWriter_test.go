package stream

import (
	"bytes"
	"testing"
)

func TestCount(t *testing.T) {
	w := NewCountWriter(new(bytes.Buffer))
	w.Write([]byte{'H', 'e', 'l', 'l', 'o'})
	if w.Written() != 5 {
		t.Fatal("Length should be 5")
	}
	// Check second time
	if w.Written() != 5 {
		t.Fatal("Length should be 5")
	}

	w.ResetWritten()
	if w.Written() != 0 {
		t.Fatal("Should be empty")
	}

	w.WriteString("Hello")
	if w.Written() != 5 {
		t.Fatal("Length should be 5")
	}
}

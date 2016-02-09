// Implement multipart yEnc
package yenc

import (
	"bytes"
	"io"
	"fmt"
	"hash/crc32"
)

const PART_SIZE = 768000 // Default size (750kb)

type Writer struct {
	partSize int      // Size per part
	buf *bytes.Buffer // Queue
	filename string   // Original filename

	bytePos int       // Current byte position
	pos int           // Current part position

	byteCount int     // Current size
	posCount int      //
}

// NewWriter returns a new Writer to create multipart yEnc
func NewWriter(buf *bytes.Buffer, filename string, partSize int) *Writer {
	return &Writer{
		buf: buf,
		filename: filename,
		partSize: partSize,
	}
}

// Write p to the buffer
func (w *Writer) Write(p []byte) (int, error) {
	if w.pos != 0 {
		return 0, fmt.Errorf("cannot write once reading started")
	}
	
	n, e := w.buf.Write(p)
	w.byteCount += n
	return n, e
}

// Ensure all data is processed.
func (w *Writer) Close() error {
	if w.bytePos != w.byteCount {
		return fmt.Errorf("buffer remain=%d", w.byteCount-w.bytePos)
	}
	return nil
}

// Parts returns the amount of parts it has ready
func (w *Writer) Parts() int {
	if w.posCount == 0 {
		// Calc parts
		w.posCount = yencParts(w.partSize, w.byteCount)
	}
	return w.posCount
}

// EncodePart writes the next part to w, returning io.EOF when done
func (w *Writer) EncodePart(out io.Writer) (int, error) {
	pos := 0
	{
		if w.posCount == 0 {
			w.Parts()
		}
		if w.pos == 0 && w.posCount == 0 {
			return 0, fmt.Errorf("buffer empty")
		}
		if w.pos == w.posCount {
			return 0, io.EOF
		}
		w.pos = w.pos+1
		pos = w.pos
	}

	//buf := new(bytes.Buffer)
	bufIn := w.buf.Next(w.partSize)
	size := len(bufIn)
	begin := w.bytePos
	w.bytePos = w.bytePos + size
	end := w.bytePos

	// header
	{
		prefix := yencHeader(pos, w.posCount, size, w.filename)
		prefix = prefix + yencPart(begin, end)
		if _, err := out.Write([]byte(prefix)); err != nil {
			return 0, err
		}
	}
	// body
	{
		enc := newEncoder(out, LINE_LENGTH)
		if err := enc.Encode(bufIn); err != nil {
			return 0, err
		}
	}
	// footer
	{
		h := crc32.NewIEEE()
		h.Write(bufIn)
		suffix := yencEnd(w.byteCount, pos, h.Sum32())
		if _, err := out.Write([]byte(suffix)); err != nil {
			return 0, err
		}
	}
	return size, nil
}

func (w *Writer) HasNext() bool {
    if w.pos == w.posCount {
    	return false
    }
    return true
}

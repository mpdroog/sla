// Yenc Encoding abstraction.
package yenc

import (
	"fmt"
	"bytes"
	"github.com/madcowfred/yencode"
	"hash/crc32"
	"io"
)

const ARTICLE_SIZE = 768000 // 750kb

type Part struct {
	Bytes []byte
	Begin int
	End int
}

type Encoder struct {
	FileName string
	fileSize int
	parts map[int]Part
	pos int
}

// Create parts in memory
func (e *Encoder) Build(buf *bytes.Buffer) error {
	e.parts = make(map[int]Part)
	e.pos = 0

	e.fileSize = buf.Len()
	parts := yencParts(e.fileSize)

	for i := 0; i < parts; i++ {
		part := make([]byte, ARTICLE_SIZE)
		n, err := buf.Read(part)
		if err != nil {
			return err
		}
		if n != ARTICLE_SIZE && parts-1 != i {
			return fmt.Errorf("Article size has unexpected size?")
		}
		begin := i*ARTICLE_SIZE
		e.parts[i] = Part{
			Bytes: part[0:n],
			Begin: begin+1,
			End: begin+n,
		}
	}
	return nil
}

func (e *Encoder) HasNext() bool {
	if e.pos < len(e.parts) {
		return true
	}
	return false
}

// Write next part to the given writer and return written
// bytes/error.
func (e *Encoder) Next(w io.Writer) error {
	if e.pos >= len(e.parts) {
		// DevErr: Off by one?
		panic(io.EOF)
	}

	// Pick up next part
	i := e.pos
	e.pos++

	part := e.parts[i]
	n := len(part.Bytes)
	begin := ARTICLE_SIZE * i

	// yEnc opening
	_, err := w.Write([]byte(yencHeader(i, len(e.parts), e.fileSize, e.FileName)))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(yencPart(begin+1, begin+n)))
	if err != nil {
		return err
	}

	// Body
	yencode.Encode(
		part.Bytes,
		w,
	)
	h := crc32.NewIEEE()
	h.Write(part.Bytes)
	// yEnc footer
	w.Write([]byte(yencEnd(n, i, h.Sum32())))

	return nil
}

func (e *Encoder) Parts() int {
	return len(e.parts)
}

func NewEncoder(fileName string) Encoder {
	return Encoder{FileName: fileName}
}

// --------- old
func yencHeader(part, parts, fileSize int, fileName string) string {
	return fmt.Sprintf("=ybegin part=%d total=%d line=128 size=%d name=%s\r\n", part, parts, fileSize, fileName)
}

func yencPart(posBegin, posEnd int) string {
	return fmt.Sprintf("=ypart begin=%d end=%d\r\n", posBegin, posEnd)
}

func yencEnd(size, part int, hash uint32) string {
	return fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X\r\n", size, part, hash)
}

// Determine amount of parts we need.
func yencParts(fileSize int) int {
	parts := fileSize / ARTICLE_SIZE
	mod := fileSize % ARTICLE_SIZE

	// Add one part for smaller than 750kb
	if mod != 0 {
		parts++
	}
	return parts
}

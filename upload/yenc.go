// Yenc utils
package main

import (
	"fmt"
	"bytes"
)

const ARTICLE_SIZE = 768000 // 750kb

type YencPart struct {
	Bytes []byte
	Begin int
	End int
}

func yencParts(fileSize int) int {
	parts := fileSize / ARTICLE_SIZE
	mod := fileSize % ARTICLE_SIZE

	// Add one part for smaller than 750kb
	if mod != 0 {
		parts++
	}
	return parts
}

func yencHeader(part, parts, fileSize int, fileName string) string {
	return fmt.Sprintf("=ybegin part=%d total=%d line=128 size=%d name=%s\r\n", part, parts, fileSize, fileName)
}

func yencPart(posBegin, posEnd int) string {
	return fmt.Sprintf("=ypart begin=%d end=%d\r\n", posBegin, posEnd)
}

func yencEnd(size, part int, hash uint32) string {
	return fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X\r\n", size, part, hash)
}

func Build(buf *bytes.Buffer) (map[int]YencPart, error) {
	out := make(map[int]YencPart)

	fileSize := buf.Len()
	parts := yencParts(fileSize)

	for i := 0; i < parts; i++ {
		part := make([]byte, ARTICLE_SIZE)
		n, e := buf.Read(part)
		if e != nil {
			return nil, e
		}
		if n != ARTICLE_SIZE && parts-1 != i {
			return nil, fmt.Errorf("Article size has unexpected size?")
		}
		begin := i*ARTICLE_SIZE
		out[i] = YencPart{
			Bytes: part[0:n],
			Begin: begin+1,
			End: begin+n,
		}
	}

	return out, nil
}
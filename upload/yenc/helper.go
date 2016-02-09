package yenc

import (
	"fmt"
)

// Determine amount of parts we need.
func yencParts(partSize, fileSize int) int {
	parts := fileSize / partSize
	mod := fileSize % partSize

	// Add one part for smaller than 750kb
	if mod != 0 {
		parts++
	}
	return parts
}

func yencHeader(part, parts, fileSize int, fileName string) string {
	if part <= 0 {
		panic("Part indices must be > 0")
	}
	return fmt.Sprintf("=ybegin part=%d total=%d line=128 size=%d name=%s\r\n", part, parts, fileSize, fileName)
}

func yencPart(posBegin, posEnd int) string {
	return fmt.Sprintf("=ypart begin=%d end=%d\r\n", posBegin, posEnd)
}

func yencEnd(size, part int, hash uint32) string {
	if part <= 0 {
		panic("Part indices must be > 0")
	}
	return fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X\r\n", size, part, hash)
}

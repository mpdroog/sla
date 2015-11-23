// Yenc utils
package main

import "fmt"

func yencHeader(part, parts, fileSize int, fileName string) string {
	return fmt.Sprintf("=ybegin part=%d total=%d line=128 size=%d name=%s\r\n", part, parts, fileSize, fileName)
}

func yencPart(posBegin, posEnd int) string {
	return fmt.Sprintf("=ypart begin=%d end=%d\r\n", posBegin, posEnd)
}

func yencEnd(size, part int, hash uint32) string {
	return fmt.Sprintf("=yend size=%d part=%d pcrc32=%08X\r\n", size, part, hash)
}
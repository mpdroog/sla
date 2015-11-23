package main

import (
	"testing"
)

func TestYencParts(t *testing.T) {
	if yencParts(1) != 1 {
		t.Errorf("yenc parts(1) wrong")
	}
	if yencParts(ARTICLE_SIZE) != 1 {
		t.Errorf("yenc parts(ART_SIZE) wrong")
	}
	if yencParts(1024*1024) != 2 {
		t.Errorf("yenc parts(1024) wrong")
	}
}

func TestHeader(t *testing.T) {
	std := "=ybegin part=1 total=10 line=128 size=1024 name=randomFile123.zip\r\n"
	if yencHeader(1, 10, 1024, "randomFile123.zip") != std {
		t.Errorf("ybegin wrong")
	}
}

func TestPart(t *testing.T) {
	std := "=ypart begin=1020 end=1024\r\n"
	out := yencPart(1020, 1024)
	if out != std {
		t.Errorf("ypart wrong expect: %s received: %s", std, out)
	}
}

func TestEnd(t *testing.T) {
	// 123(10) = 7B(16)
	std := "=yend size=1024 part=1 pcrc32=0000007B\r\n"
	out := yencEnd(1024, 1, 123)
	if out != std {
		t.Errorf("yend wrong expect: %s received: %s", std, out)
	}
}
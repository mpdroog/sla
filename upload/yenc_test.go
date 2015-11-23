package main

import (
	"testing"
	"bytes"
	"archive/zip"
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

func TestBuild(t *testing.T) {
	// Load 700kb+ of testdata
	buf := new(bytes.Buffer)
	{
		w := zip.NewWriter(buf)

		if e := zipAdd(w, "UF-Logo.jpg", "./test/UF-Logo.jpg"); e != nil {
			panic(e)
		}
		if e := zipAdd(w, "Why UF.pdf", "./test/Why UF.pdf"); e != nil {
			panic(e)
		}
		if e := zipAdd(w, "Why UF.tiff", "./test/Why UF.tiff"); e != nil {
			panic(e)
		}

		if e := w.Close(); e != nil {
			panic(e)
		}
	}

	size := len(buf.Bytes())
	if size != 792373 {
		t.Errorf("Buffersize not as hardwired?")
	}
	parts, e := Build(buf)
	if e != nil {
		panic(e)
	}

	if len(parts) != 2 {
		t.Errorf("Parts not as hardwired?")
	}

	for idx, part := range parts {
		var cmp YencPart
		if idx == 0 {
			cmp = YencPart{
				Begin: 1,
				End: 768000,
			}
		} else {
			cmp = YencPart{
				Begin: 768001,
				End: 792373,
			}
		}

		if len(part.Bytes) != cmp.End - cmp.Begin +1 {
			t.Errorf("part.Len mismatch ")
		}
		if part.Begin != cmp.Begin {
			t.Errorf("part.Begin mismatch")
		}
		if part.End != cmp.End {
			t.Errorf("part.End mismatch")
		}
	}
}
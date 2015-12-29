package yenc

import (
    "bytes"
    "io"
    "io/ioutil"
    "testing"
    "fmt"
)

func TestYencodeText(t *testing.T) {
    // open and read the input file
    inbuf, err := ioutil.ReadFile("../test/test1.in")
    if err != nil {
        t.Fatalf("couldn't open test1.in: %s", err)
    }

    // open and read the yencode output file
    testbuf, err := ioutil.ReadFile("../test/test1.ync")
    if err != nil {
        t.Fatalf("couldn't open test1.ync: %s", err)
    }

    // generate a dodgy message
    out := new(bytes.Buffer)

    io.WriteString(out, "=ybegin part=1 total=1 line=128 size=858 name=test1.in\r\n")
    io.WriteString(out, "=ypart begin=1 end=858\r\n")
    enc := newEncoder(out, LINE_LENGTH)
    if err := enc.Encode(inbuf); err != nil {
        t.Fatalf("write failed %s", err)
    }
    io.WriteString(out, "=yend size=858 part=1 pcrc32=3274F3F7\r\n")

    // compare
    if bytes.Compare(testbuf, out.Bytes()) != 0 {
        fmt.Printf("%s NOT %s", testbuf, out.Bytes())
        t.Fatalf("data mismatch")
    }
}

func TestYencodeBinary(t *testing.T) {
    // open and read the input file
    inbuf, err := ioutil.ReadFile("../test/test2.in")
    if err != nil {
        t.Fatalf("couldn't open test2.in: %s", err)
    }

    // open and read the yencode output file
    testbuf, err := ioutil.ReadFile("../test/test2.ync")
    if err != nil {
        t.Fatalf("couldn't open test2.ync: %s", err)
    }

    // generate a dodgy message
    out := new(bytes.Buffer)

    io.WriteString(out, "=ybegin part=1 total=1 line=128 size=76800 name=test2.in\r\n")
    io.WriteString(out, "=ypart begin=1 end=76800\r\n")
    enc := newEncoder(out, LINE_LENGTH)
    if err := enc.Encode(inbuf); err != nil {
        t.Fatalf("write failed %s", err)
    }
    io.WriteString(out, "=yend size=76800 part=1 pcrc32=12AAC2CF\r\n")

    // compare
    if bytes.Compare(testbuf, out.Bytes()) != 0 {
        t.Fatalf("data mismatch")
    }
}

func bench(b *testing.B, n int) {
    inbuf := makeInBuf(n)
    out := new(bytes.Buffer)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        if i > 0 {
            out.Reset()
        }
        enc := newEncoder(out, LINE_LENGTH)
        if err := enc.Encode(inbuf); err != nil {
            panic(err)
        }
    }

    b.SetBytes(int64(len(inbuf)))
}

func BenchmarkEncode10(b *testing.B) {
    bench(b, 10)
}

func BenchmarkEncode100(b *testing.B) {
    bench(b, 100)
}

func BenchmarkEncode1000(b *testing.B) {
    bench(b, 1000)
}

func makeInBuf(length int) []byte {
    chars := length * 256 * 132
    pos := 0

    in := make([]byte, chars)
    for i := 0; i < length; i++ {
        for j := 0; j < 256; j++ {
            for k := 0; k < 132; k++ {
                in[pos] = byte(j)
                pos++
            }
        }
    }

    return in
}
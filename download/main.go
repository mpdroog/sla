package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chrisfarms/yenc"
	"io"
	"net/textproto"
	"os"
	"sla/lib/nntp"
	"sla/lib/nzb"
	"sla/lib/stream"
	"strings"
	"time"
)

type Config struct {
	Address string // server:port
	User    string
	Pass    string
	NzbDir  string
	Output  string
}

type Perf struct {
	Conn  float64
	Auth  float64
	Arts  []float64
	KBsec []float64
	Error []string
}

var C Config

func loadConfig(file string) (Config, error) {
	var c Config
	r, e := os.Open(file)
	if e != nil {
		return c, e
	}
	e = json.NewDecoder(r).Decode(&c)
	return c, e
}

func fail(e error) {
	w, ew := os.Create(C.Output)
	if ew != nil {
		panic(ew)
	}
	defer w.Close()
	enc := json.NewEncoder(w)

	if ew = enc.Encode(Perf{
		Arts: []float64{},
		KBsec: []float64{},
		Error: []string{e.Error()},
	}); ew != nil {
		panic(ew)
	}

	panic(e)
}

func main() {
	var e error
	var verbose, skipyenc bool
	var configPath, date string
	flag.BoolVar(&verbose, "v", false, "Verbosity")
	flag.BoolVar(&skipyenc, "y", false, "Skip yEnc decode")
	flag.StringVar(&configPath, "c", "./config.json", "/Path/to/config.json")
	flag.StringVar(&date, "d", "", "YYYY-mm-dd to download from nzbdir")
	flag.Parse()

	C, e = loadConfig(configPath)
	if e != nil {
		fail(e)
	}
	if !strings.HasSuffix(C.NzbDir, "/") {
		C.NzbDir += "/"
	}
	if date == "" {
		// default to today
		date = time.Now().Format("2006-01-02")
	}
	if verbose {
		fmt.Printf("Config=%+v Date=%+v\n", C, date)
	}

	// Force valid date pattern
	if _, e := time.Parse("2006-01-02", date); e != nil {
		fail(e)
	}

	fd, e := os.Open(C.NzbDir + date + ".nzb")
	if e != nil {
		fail(e)
	}
	defer fd.Close()
	arts, e := nzb.Read(fd)
	if e != nil {
		fail(e)
	}

	if verbose {
		fmt.Println("Connecting to nntp..")
	}
	conn := nntp.New(C.Address, "1", verbose)
	conn.Verbose = verbose
	perfBegin := time.Now()
	var perfInit, perfAuth time.Time
	{
		defer conn.Close()
		if e := conn.Init(); e != nil {
			fail(e)
		}
		perfInit = time.Now()
		if e := conn.Auth(C.User, C.Pass); e != nil {
			fail(e)
		}
		perfAuth = time.Now()
	}

	perfArts := []float64{}
	KBsecs := []float64{}
	lastPerf := time.Now()
	buf := new(bytes.Buffer)
	for _, segment := range arts.File.Segments.Segment {
		buf.Reset()
		conn.Article(segment.Msgid)

		counter := stream.NewCountReader(conn.GetReader())
		rawread := bufio.NewReader(counter)

		_, e = textproto.NewReader(rawread).ReadMIMEHeader()
		if e != nil {
			fail(e)
		}

		if _, e := io.Copy(buf, rawread); e != nil {
			fail(e)
		}
		n := counter.ReadReset()
		if int64(n) <= segment.Bytes {
			panic(fmt.Errorf("ByteCount mismatch, expect>%d recv=%d", segment.Bytes, n))
		}

		if !skipyenc {
			if _, e = yenc.Decode(buf); e != nil {
				fmt.Printf("%+v\n", string(buf.Bytes()))
				fail(e)
			}
		}

		now := time.Now()
		diff := now.Sub(lastPerf)
		kbSec := float64(n/1024) / diff.Seconds()

		if verbose {
			fmt.Println(fmt.Sprintf(
				"Download %s (%d bytes in %s with %f KB/s)",
				segment.Msgid, n, diff.String(), kbSec,
			))
		}

		KBsecs = append(KBsecs, kbSec)
		perfArts = append(perfArts, MilliSeconds(diff))
		lastPerf = now
	}

	w, e := os.Create(C.Output)
	if e != nil {
		fail(e)
	}
	defer w.Close()
	enc := json.NewEncoder(w)

	if e := enc.Encode(Perf{
		Conn:  MilliSeconds(perfInit.Sub(perfBegin)),
		Auth:  MilliSeconds(perfAuth.Sub(perfInit)),
		Arts:  perfArts,
		KBsec: KBsecs,
	}); e != nil {
		fail(e)
	}
}

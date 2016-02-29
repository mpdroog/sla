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
	"strings"
	"time"
)

type Config struct {
	Address string // server:port
	User    string
	Pass    string
	NzbDir  string
}

type Perf struct {
	Conn  float64
	Auth  float64
	Arts  []float64
	KBsec []float64
}

func loadConfig(file string) (Config, error) {
	var c Config
	r, e := os.Open(file)
	if e != nil {
		return c, e
	}
	e = json.NewDecoder(r).Decode(&c)
	return c, e
}

func main() {
	var verbose, skipyenc bool
	var configPath, date string
	flag.BoolVar(&verbose, "v", false, "Verbosity")
	flag.BoolVar(&skipyenc, "y", false, "Skip yEnc decode")
	flag.StringVar(&configPath, "c", "./config.json", "/Path/to/config.json")
	flag.StringVar(&date, "d", "", "YYYY-mm-dd to download from nzbdir")
	flag.Parse()

	c, e := loadConfig(configPath)
	if e != nil {
		panic(e)
	}
	if !strings.HasSuffix(c.NzbDir, "/") {
		c.NzbDir += "/"
	}
	if date == "" {
		// default to today
		date = time.Now().Format("2006-01-02")
	}
	if verbose {
		fmt.Printf("Config=%+v Date=%+v\n", c, date)
	}

	// Force valid date pattern
	if _, e := time.Parse("2006-01-02", date); e != nil {
		panic(e)
	}

	fd, e := os.Open(c.NzbDir + date + ".nzb")
	if e != nil {
		panic(e)
	}
	defer fd.Close()
	arts, e := nzb.Read(fd)
	if e != nil {
		panic(e)
	}

	if verbose {
		fmt.Println("Connecting to nntp..")
	}
	conn := nntp.New(c.Address, "1", verbose)
	conn.Verbose = verbose
	perfBegin := time.Now()
	var perfInit, perfAuth time.Time
	{
		defer conn.Close()
		if e := conn.Init(); e != nil {
			panic(e)
		}
		perfInit = time.Now()
		if e := conn.Auth(c.User, c.Pass); e != nil {
			panic(e)
		}
		perfAuth = time.Now()
	}

	perfArts := []float64{}
	KBsecs := []float64{}
	lastPerf := time.Now()
	buf := new(bytes.Buffer)
	for _, segment := range arts.File.Segments.Segment {
		buf.Reset()
		byteCount := segment.Bytes
		conn.Article(segment.Msgid)
		rawread := bufio.NewReader(conn.GetReader())

		_, e = textproto.NewReader(rawread).ReadMIMEHeader()
		if e != nil {
			panic(e)
		}

		if _, e := io.Copy(buf, rawread); e != nil {
			panic(e)
		}

		if !skipyenc {
			if _, e = yenc.Decode(buf); e != nil {
				panic(e)
			}
		}

		now := time.Now()
		diff := now.Sub(lastPerf)
		kbSec := float64(byteCount/1024) / diff.Seconds()

		if verbose {
			fmt.Println(fmt.Sprintf(
				"Download %s (%d bytes in %s with %f KB/s)",
				segment.Msgid,
				byteCount,
				diff.String(),
				kbSec,
			))
		}

		KBsecs = append(KBsecs, kbSec)
		perfArts = append(perfArts, MilliSeconds(diff))
		lastPerf = now
	}

	stat, e := json.Marshal(Perf{
		Conn:  MilliSeconds(perfInit.Sub(perfBegin)),
		Auth:  MilliSeconds(perfAuth.Sub(perfInit)),
		Arts:  perfArts,
		KBsec: KBsecs,
	})
	if e != nil {
		panic(e)
	}
	if _, e := os.Stdout.Write(stat); e != nil {
		panic(e)
	}
}

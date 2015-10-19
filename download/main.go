package main

import (
	"sla/lib/nntp"
	"fmt"
	"time"
	"strings"
	"flag"
	"encoding/json"
	"os"
	"sla/lib/nzb"
	"net/textproto"
	"io"
	"bufio"
	"bytes"
	"github.com/chrisfarms/yenc"
)

type Config struct {
	Listen string // server:port
	User string
	Pass string
	NzbDir string
}

type Perf struct {
	Conn string
	Auth string
	Arts []string
}

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Verbosity")
	flag.Parse()

	c := Config{
		"news.usenet.farm:119",
		"jethro", "jethro",
		"/usr/local/sla/retention/",
	}
	if !strings.HasSuffix(c.NzbDir, "/") {
		c.NzbDir += "/"
	}

	arts, e := nzb.Open(c.NzbDir + time.Now().Format("2006-01-02") + ".nzb")
	if e != nil {
		panic(e)
	}

	if verbose {
		fmt.Println("Connecting to nntp..")
	}
	conn := nntp.New(c.Listen, "1")
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

	perfArts := []string{}
	lastPerf := time.Now()
	for _, segment := range arts.File.Segments.Segment {
		conn.Article(segment.Msgid)
		rawread := bufio.NewReader(conn.GetReader())

		_, e = textproto.NewReader(rawread).ReadMIMEHeader()
		if e != nil {
			panic(e)
		}

		buf := new(bytes.Buffer)
		if _, e := io.Copy(buf, rawread); e != nil {
			panic(e)
		}

		_, e = yenc.Decode(buf)
		if e != nil {
			panic(e)
		}

		now := time.Now()
		if verbose {
			fmt.Println(fmt.Sprintf(
				"Download %s (%s bytes in %s)",
				segment.Msgid,
				segment.Bytes,
				now.Sub(lastPerf).String(),
			))
		}
		perfArts = append(perfArts, now.Sub(lastPerf).String())
		lastPerf = now
	}

	stat, e := json.Marshal(Perf{
		Conn: perfInit.Sub(perfBegin).String(),
		Auth: perfAuth.Sub(perfInit).String(),
		Arts: perfArts,
	})
	if e != nil {
		panic(e)
	}
	if _, e := os.Stdout.Write(stat); e != nil {
		panic(e)
	}
}

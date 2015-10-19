package main

import (
	"github.com/madcowfred/yencode"
	"sla/lib/nntp"
	"sla/lib/nzb"
	"sla/lib/stream"
	"fmt"
	"archive/zip"
	"os"
	"io"
	"bytes"
	"bufio"
	"time"
	"strings"
	"io/ioutil"
)

type Config struct {
	Listen string // server:port
	User string
	Pass string
	NzbDir string
}

func zipAdd(w *zip.Writer, name string, path string) error {
	in, e := os.Open(path)
	if e != nil {
		return e
	}

	f, e := w.Create("README.txt")
	if e != nil {
		return e
	}
	if _, e := io.Copy(f, bufio.NewReader(in)); e != nil {
		return e
	}
	return nil
}

func headers(subject string, msgid string) string {
	headers := "Message-ID: <" + msgid + ">" + nntp.EOF
	headers += "Date: " + time.Now().String() + nntp.EOF
	headers += "Organization: Usenet.Farm" + nntp.EOF
	headers += "Subject: " + subject + nntp.EOF
	headers += "From: Usenet.Farm" + nntp.EOF
	headers += "Newsgroups: alt.binaries.test" + nntp.EOF
	headers += nntp.EOF // End of header
	return headers;
}

func main() {
	c := Config{
		"news.usenet.farm:119",
		"jethro", "jethro",
		"/usr/local/sla/retention/",
	}
	if !strings.HasSuffix(c.NzbDir, "/") {
		c.NzbDir += "/"
	}

	// Permission check
	{
		stat, e := os.Stat(c.NzbDir)
		if e != nil {
			panic(e)
		}
		if !stat.IsDir() {
			fmt.Println("Not a dir: " + c.NzbDir)
			os.Exit(1)
			return
		}
		if e := ioutil.WriteFile(
			c.NzbDir + "check.txt",
			[]byte("Write permission check."),
			0400,
		); e != nil {
			panic(e)
		}
		if e := os.Remove(c.NzbDir + "check.txt"); e != nil {
			panic(e)
		}
	}

	fmt.Println("Building ZIP..")
	buf := new(bytes.Buffer)
	{
		w := zip.NewWriter(buf)

		if e := zipAdd(w, "README.md", "./dummy/README.txt"); e != nil {
			panic(e)
		}
		if e := zipAdd(w, "100mb.bin", "./dummy/100mb.bin"); e != nil {
			panic(e)
		}

		f, e := w.Create("unique.txt")
		if e != nil {
			panic(e)
		}
		f.Write([]byte(RandStringRunes(16)))

		if e := w.Close(); e != nil {
			panic(e)
		}
	}

	fmt.Println("Connecting to nntp..")
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

	articleSize := 768000 // 750kb
	parts := buf.Len() / articleSize
	mod := buf.Len() % articleSize
	if mod != 0 {
		parts++
	}

	msgids := make(map[string]int64)
	subject := "Completion test " + time.Now().Format("2006-01-02")
	fmt.Println(fmt.Sprintf("Upload file=%s parts(%d)..", subject, parts))
	artPerf := []string{}
	lastPerf := time.Now()
	for i := 0; i < parts; i++ {
	 	if e := conn.Post(); e != nil {
			panic(e)
		}
		part := buf.Next(articleSize)
		msgid := RandStringRunes(16) + "@usenet.farm"

		w := stream.NewCountWriter(conn.GetWriter())
		if _, e := w.WriteString(headers(subject, msgid)); e != nil {
			panic(e)
		}
		yencode.Encode(
			part,
			w,
		)
		msgids[msgid] = w.Written()

		if e := conn.PostClose(); e != nil {
			panic(e)
		}

		now := time.Now()
		d := now.Sub(lastPerf).String()

		fmt.Println(fmt.Sprintf("Posted " + msgid + " in " + d))
		artPerf = append(artPerf, d)
		lastPerf = now
	}

	xml := nzb.Build(subject, msgids)
	if e := ioutil.WriteFile(
		c.NzbDir + time.Now().Format("2006-01-02") + ".xml",
		[]byte(xml), 400,
	); e != nil {
		panic(e)
	}

	fmt.Println(fmt.Sprintf(
		"Perf: conn=%s, auth=%s, arts=[%s]",
		perfInit.Sub(perfBegin).String(),
		perfAuth.Sub(perfInit).String(),
		strings.Join(artPerf, ","),
	))
}

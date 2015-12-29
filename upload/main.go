package main

import (
	"sla/lib/nntp"
	"sla/lib/nzb"
	"sla/lib/stream"
	"sla/upload/yenc"
	"fmt"
	"archive/zip"
	"os"
	"io"
	"bytes"
	"bufio"
	"time"
	"strings"
	"io/ioutil"
	"encoding/json"
	"flag"
	"path/filepath"
)

type Config struct {
	Address string // server:port
	User string
	Pass string
	NzbDir string
	MsgDomain string
	UploadDir string
}

type ArtPerf struct {
	MsgId string
	Time int64 // duration in nanoseconds
	Size int64
	Speed float64 // kb/sec
	BitSpeed float64 // kbit/sec
}

type Perf struct {
	Conn string
	Auth string
	Arts []ArtPerf
}

func zipAdd(w *zip.Writer, name string, path string) error {
	in, e := os.Open(path)
	if e != nil {
		return e
	}

	head := &zip.FileHeader{Name: name}
	head.SetModTime(time.Now())
	f, e := w.CreateHeader(head)
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
	headers += "Date: " + time.Now().Format(time.RFC822) + nntp.EOF
	headers += "Organization: Usenet.Farm" + nntp.EOF
	headers += "Subject: " + subject + nntp.EOF
	headers += "From: Usenet.Farm" + nntp.EOF
	headers += "Newsgroups: alt.binaries.test" + nntp.EOF
	headers += nntp.EOF // End of header
	return headers;
}

func min(a int, b int64) int {
	if a < int(b) {
		return a
	}
	return int(b)
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
	var verbose bool
	var configPath string

	flag.BoolVar(&verbose, "v", false, "Verbosity")
	flag.StringVar(&configPath, "c", "./config.json", "/Path/to/config.json")
	flag.Parse()

	c, e := loadConfig(configPath)
	if e != nil {
		panic(e)
	}
	if !strings.HasSuffix(c.NzbDir, "/") {
		c.NzbDir += "/"
	}

	// Permission check nzbdir
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
	// Permission check uploaddir
	{
		stat, e := os.Stat(c.UploadDir)
		if e != nil {
			panic(e)
		}
		if !stat.IsDir() {
			fmt.Println("Not a dir: " + c.UploadDir)
			os.Exit(1)
			return
		}
	}

	if verbose {
		fmt.Println("Building ZIP from dir=" + c.UploadDir)
	}

	enc := yenc.NewWriter(
		new(bytes.Buffer),
		fmt.Sprintf("sla-%s.zip", time.Now().Format("2006-01-02")),
		yenc.PART_SIZE,
	)
	var partCount = 0
	{
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)

		e := filepath.Walk(c.UploadDir, func(path string, info os.FileInfo, err error) error {
			if path == c.UploadDir {
				// Ignore base
				return nil
			}
			if strings.HasSuffix(info.Name(), ".sh") {
				// Ignore scripts
				if verbose {
					fmt.Println("Skip " + path)
				}
				return nil
			}
			if verbose {
				fmt.Println("Add " + path + " to ZIP.")
			}
			if e := zipAdd(w, info.Name(), path); e != nil {
				return e
			}
			return nil
		})
		if e != nil {
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

		if _, err := enc.Write(buf.Bytes()); err != nil {
			panic(err)
		}
		partCount = enc.Parts()
	}
	if partCount < 50 {
		fmt.Printf("Need at least 50 parts, I got: %d (increase rand file?)\n", partCount)
		os.Exit(1)
	}

	subject := "Completion test " + time.Now().Format("2006-01-02")
	if verbose {
		fmt.Println(fmt.Sprintf("Upload file=%s parts(%d)..", subject, partCount))
	}

	if verbose {
		fmt.Println("Connecting to nntp..")
	}
	conn := nntp.New(c.Address, "1")
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

	var msgids []nzb.Msg
	artPerf := []ArtPerf{}
	lastPerf := time.Now()

	for enc.HasNext() {
	 	if e := conn.Post(); e != nil {
			panic(e)
		}
		msgid := RandStringRunes(16) + c.MsgDomain

		w := stream.NewCountWriter(conn.GetWriter())
		if _, e := w.WriteString(headers(subject, msgid)); e != nil {
			panic(e)
		}
		w.ResetWritten()

		if _, e := enc.EncodePart(w); e != nil {
			panic(e)
		}
		n := w.Written()
		msgids = append(msgids, nzb.Msg{
			Msgid: msgid,
			Size: n,
		})

		if e := conn.PostClose(); e != nil {
			panic(e)
		}

		// Stats
		now := time.Now()
		d := now.Sub(lastPerf)

		if verbose {
			fmt.Println(fmt.Sprintf(
				"Posted %s in %s",
				msgid, d.String(),
			))
		}
		kbSec := float64(n/1024) / d.Seconds()
		artPerf = append(artPerf, ArtPerf{
			MsgId: msgid,
			Time: d.Nanoseconds(),
			Size: n,
			Speed: kbSec,     // kb/sec
			BitSpeed: kbSec*8,// kbit/sec
		})
		lastPerf = now
	}
	if err := enc.Close(); err != nil {
		panic(err)
	}

	xml := nzb.Build(subject, msgids, time.Now().Format(time.RFC822))
	if e := ioutil.WriteFile(
		c.NzbDir + time.Now().Format("2006-01-02") + ".nzb",
		[]byte(xml), 400,
	); e != nil {
		panic(e)
	}

	stat, e := json.Marshal(Perf{
		Conn: perfInit.Sub(perfBegin).String(),
		Auth: perfAuth.Sub(perfInit).String(),
		Arts: artPerf,
	})
	if e != nil {
		panic(e)
	}
	if _, e := os.Stdout.Write(stat); e != nil {
		panic(e)
	}
}

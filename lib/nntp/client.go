package nntp

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const EOF = "\r\n"      // End of File
const EOM = "\r\n.\r\n" // End of multiline data-block
var ERR_RANGE = errors.New("NNTP StatusCode > 2xx")

type Client struct {
	Name    string
	Ready   bool
	Verbose bool
	listen  string
	L       *log.Logger

	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer

	BytesIn  int64
	BytesOut int64
}

func (c *Client) Init() error {
	conn, e := net.Dial("tcp", c.listen)
	if e != nil {
		return e
	}
	c.conn = conn
	c.r = bufio.NewReader(conn)
	c.w = bufio.NewWriter(conn)

	// Welcome
	l, e := c.Read()
	if e != nil {
		return e
	}
	if !strings.HasPrefix(l, "20") {
		// 2xx - Command ok
		// x0x - Connection, setup, and miscellaneous messages
		return errors.New("Invalid welcome: " + l)
	}
	return nil
}

func (c *Client) Read() (string, error) {
	txt, e := c.r.ReadString(EOF[1])
	if e != nil {
		return "", e
	}
	if strings.HasSuffix(txt, EOF) {
		txt = txt[:len(txt)-2] // Strip EOF
	} else {
		return "", errors.New("Line does not end with CrLf")
	}
	// TODO: valid?
	c.BytesIn += int64(len([]byte(txt)))
	c.log("C(%s) << %s", c.Name, txt)
	return txt, nil
}

// Check next line for expected prefixes
func (c *Client) Expect(prefixes []Expect) (string, error) {
	l, e := c.Read()
	if e != nil {
		return "", e
	}

	ok := false
	errRange := false
	for _, prefix := range prefixes {
		if strings.HasPrefix(l, prefix.Prefix) {
			if prefix.IsErr {
				errRange = true
			}
			ok = true
			break
		}
	}
	if !ok {
		return "", fmt.Errorf("Protocol error. Received=%s (Expected=%+v)", l, prefixes)
	}
	if errRange {
		return l, ERR_RANGE
	}
	return l, nil
}

// Send cmd and expect response to begin with prefix
func (c *Client) Send(cmd string, prefixes []Expect) (string, error) {
	c.log("C(%s) >> %s", c.Name, cmd)
	if _, e := c.w.WriteString(cmd + EOF); e != nil {
		return "", e
	}
	if e := c.w.Flush(); e != nil {
		return "", e
	}

	l, e := c.Expect(prefixes)
	if e != nil {
		return l, e
	}
	return l, nil
}

func (c *Client) GetWriter() *bufio.Writer {
	return c.w
}

func (c *Client) GetReader() *DotReader {
	return NewDotReader(c.r, false)
}

func New(listen string, name string, verbose bool) *Client {
	return &Client{
		Name:    name,
		listen:  listen,
		Verbose: verbose,
		// TODO: cleanup..
		L: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}

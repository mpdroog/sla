package nntp

func (c *Client) Auth(user string, pass string) error {
	if _, e := c.Send("authinfo user "+user, []Expect{Expect{"381 ", false}}); e != nil {
		return e
	}
	if _, e := c.Send("authinfo pass "+pass, []Expect{Expect{"281 ", false}}); e != nil {
		return e
	}
	c.Ready = true
	return nil
}

func (c *Client) Post() error {
	if _, e := c.Send("POST", []Expect{Expect{"340 ", false}}); e != nil {
		return e
	}
	return nil
}

func (c *Client) PostClose() error {
	c.w.WriteString(EOM)
	if e := c.w.Flush(); e != nil {
		return e
	}

	_, e := c.Expect([]Expect{Expect{"240 ", false}})
	if e != nil {
		return e
	}
	return nil
}

func (c *Client) Article(msgid string) error {
	if _, e := c.Send("article <"+msgid+">", []Expect{Expect{"201 ", false}}); e != nil {
		return e
	}
	return nil
}

func (c *Client) Close() error {
	c.Ready = false
	// Ignore any err
	c.w.Write([]byte("QUIT\r\n"))
	return c.conn.Close()
}

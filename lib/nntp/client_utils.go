package nntp

// Log worker msg (by default DEBUG-level)
func (c *Client) log(format string, v ...interface{}) {
	if c.Verbose {
		c.L.Printf(format, v...)
	}
}
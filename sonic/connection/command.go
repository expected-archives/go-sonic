package connection

func (c *Connection) RunSimpleCommand(args ...string) (string, error) {
	if err := c.write(args); err != nil {
		return "", err
	}
	return c.readLine()
}

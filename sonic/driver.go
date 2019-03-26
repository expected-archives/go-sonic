package sonic

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

var (
	ClosedError     = errors.New("sonic connection is closed")
	InvalidChanName = errors.New("invalid channel name")
)

// Base contains commons commands to all channels.
type Base interface {
	// Quit stop connection, you can't execute anything after calling this method.
	// Syntax command QUIT
	Quit() error

	// Ping ping the sonic server.
	// Return an error is there is something wrong.
	// If an error occur, the sonic server is maybe down.
	// Syntax command PING
	Ping() error
}

type Driver struct {
	Host     string
	Port     int
	Password string

	channel Channel
	reader  *bufio.Reader
	conn    net.Conn
	closed  bool
}

// Connect open a connection via TCP with the sonic server.
func (c *Driver) Connect() error {
	if !IsChannelValid(c.channel) {
		return InvalidChanName
	}

	c.clean()
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	} else {
		c.conn = conn
		c.reader = bufio.NewReader(c.conn)

		err := c.write(fmt.Sprintf("START %s %s", c.channel, c.Password))
		if err != nil {
			return err
		}

		_, err = c.read()
		_, err = c.read()
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *Driver) Quit() error {
	err := c.write("QUIT")
	if err != nil {
		return err
	}

	// should get ENDED
	_, err = c.read()
	c.clean()
	return err
}

func (c Driver) Ping() error {
	err := c.write("PING")
	if err != nil {
		return err
	}

	// should get PONG
	_, err = c.read()
	if err != nil {
		return err
	}
	return nil
}

func (c *Driver) read() (string, error) {
	if c.closed {
		return "", ClosedError
	}
	buffer := bytes.Buffer{}
	for {
		line, isPrefix, err := c.reader.ReadLine()
		buffer.Write(line)
		if err != nil {
			if err == io.EOF {
				c.clean()
			}
			return "", err
		}
		if !isPrefix {
			break
		}
	}

	str := buffer.String()
	if strings.HasPrefix(str, "ERR ") {
		return "", errors.New(str[4:])
	}
	return str, nil
}

func (c Driver) write(str string) error {
	if c.closed {
		return ClosedError
	}
	_, err := c.conn.Write([]byte(str + "\r\n"))
	return err
}

func (c *Driver) clean() {
	c.closed = true
	_ = c.conn.Close()
	c.conn = nil
	c.reader = nil
}

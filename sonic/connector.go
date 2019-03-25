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

var ClosedError = errors.New("sonic connection is closed")

type Connection struct {
	Host     string
	Port     int
	Password string
	Channel  Channel

	reader *bufio.Reader
	conn   net.Conn
	closed bool
}

func (c *Connection) Connect() error {
	if !IsChannelValid(c.Channel) {
		return errors.New("invalid channel name")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	} else {
		c.conn = conn
		c.reader = bufio.NewReader(c.conn)

		err := c.write(fmt.Sprintf("START %s %s", c.Channel, c.Password))
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

func (c *Connection) read() (string, error) {
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

func (c Connection) write(str string) error {
	if c.closed {
		return ClosedError
	}
	_, err := c.conn.Write([]byte(str + "\r\n"))
	return err
}

func (c *Connection) Quit() error {
	err := c.write("QUIT")
	if err != nil {
		return err
	}
	// should get ENDED
	_, err = c.read()
	c.clean()
	return err
}

func (c *Connection) Ping() error {
	err := c.write("PING")
	if err != nil {
		fmt.Println("err write")
		return err
	}

	// should get PONG
	_, err = c.read()
	if err != nil {
		fmt.Println("err read")
		return err
	}
	return nil
}

func (c *Connection) clean() {
	c.closed = true
	_ = c.conn.Close()
	c.conn = nil
	c.reader = nil
}

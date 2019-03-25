package sonic

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
)

type Connection struct {
	Host     string
	Port     int
	Password string
	Channel  Channel

	reader *bufio.Reader
	conn   net.Conn
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

func (c Connection) read() (string, error) {
	buffer := bytes.Buffer{}
	for {
		line, isPrefix, err := c.reader.ReadLine()
		buffer.Write(line)
		if err != nil || !isPrefix {
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
	_, err := c.conn.Write([]byte(str + "\r\n"))
	return err
}

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
	ClosedError       = errors.New("sonic connection is closed")
	InvalidChanName   = errors.New("invalid channel name")
	InvalidActionName = errors.New("invalid action name")
)

type Driver struct {
	Host     string
	Port     int
	Password string
	Channel  Channel

	reader *bufio.Reader
	conn   net.Conn
	closed bool
}

func NewControl(host string, port int, password string) (*Driver, error) {
	driver := &Driver{
		Host:     host,
		Port:     port,
		Password: password,
		Channel:  Ingest,
	}

	return driver, driver.connect()
}

func (c *Driver) connect() error {
	if !IsChannelValid(c.Channel) {
		return InvalidChanName
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

func (c Driver) Trigger(action Action) error {
	if IsActionValid(action) {
		return InvalidActionName
	}
	err := c.write(fmt.Sprintf("TRIGGER %s", action))
	if err != nil {
		return err
	}

	// should get OK
	_, err = c.read()
	if err != nil {
		return err
	}
	return nil
}

func (c *Driver) clean() {
	c.closed = true
	_ = c.conn.Close()
	c.conn = nil
	c.reader = nil
}

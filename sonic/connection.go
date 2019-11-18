package sonic

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"unicode"
)

type connection struct {
	reader      *bufio.Reader
	conn        net.Conn
	cmdMaxBytes int
	closed      bool
}

func newConnection(d *driver) (*connection, error) {
	c := &connection{}
	c.close()
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", d.Host, d.Port))
	if err != nil {
		return nil, err
	}

	c.closed = false
	c.conn = conn
	c.reader = bufio.NewReader(c.conn)

	err = c.write(fmt.Sprintf("START %s %s", d.channel, d.Password))
	if err != nil {
		return nil, err
	}
	_, err = c.read()
	_, err = c.read()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *connection) read() (string, error) {
	if c.closed {
		return "", ErrClosed
	}
	buffer := bytes.Buffer{}
	for {
		line, isPrefix, err := c.reader.ReadLine()
		buffer.Write(line)
		if err != nil {
			if err == io.EOF {
				c.close()
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
	if strings.HasPrefix(str, "STARTED ") {

		ss := strings.FieldsFunc(str, func(r rune) bool {
			if unicode.IsSpace(r) || r == '(' || r == ')' {
				return true
			}
			return false
		})
		bufferSize, err := strconv.Atoi(ss[len(ss)-1])
		if err != nil {
			return "", errors.New(fmt.Sprintf("Unable to parse STARTED response: %s", str))
		}
		c.cmdMaxBytes = bufferSize
	}
	return str, nil
}

func (c connection) write(str string) error {
	if c.closed {
		return ErrClosed
	}
	_, err := c.conn.Write([]byte(str + "\r\n"))
	return err
}

func (c *connection) close() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.closed = true
	c.reader = nil
}

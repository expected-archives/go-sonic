package connection

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/expectedsh/go-sonic/sonic"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Connection struct {
	host       string
	port       int
	password   string
	channel    sonic.Channel
	conn       net.Conn
	reader     *bufio.Reader
	bufferSize int
}

var BufferSizeRegexp = regexp.MustCompile(`buffer\((\d+)\)`)

func NewConnection(host string, port int, password string) *Connection {
	return &Connection{
		host:     host,
		port:     port,
		password: password,
		channel:  sonic.Control,
	}
}

func (c *Connection) Open() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)

	// this line only contains the sonic version
	_, err = c.RunSimpleCommand("START", string(c.channel), c.password)
	if err != nil {
		return err
	}

	// this line contains the buffer size
	line, err := c.readLine()
	if err != nil {
		return err
	}

	bufferSize, err := strconv.Atoi(BufferSizeRegexp.FindStringSubmatch(line)[1])
	if err != nil {
		return err
	}
	c.bufferSize = bufferSize

	return nil
}

func (c *Connection) write(args []string) error {
	b := []byte(strings.Join(args, " "))
	_, err := c.conn.Write(b)
	return err
}

func (c *Connection) readLine() (string, error) {
	buffer := bytes.Buffer{}
	for {
		line, isPrefix, err := c.reader.ReadLine()
		buffer.Write(line)

		if err != nil {
			return "", err
		}

		if !isPrefix {
			break
		}
	}

	line := buffer.String()
	if strings.HasPrefix(line, "ERR ") {
		return "", errors.New(line[4:])
	}
	return line, nil
}

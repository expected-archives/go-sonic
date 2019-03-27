package sonic

import (
	"errors"
)

var (
	// ErrClosed is throw when an error with the sonic server
	// come from the state of the connection.
	ErrClosed = errors.New("sonic connection is closed")

	// ErrChanName is throw when the channel name is not supported
	// by sonic server.
	ErrChanName = errors.New("invalid channel name")
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

type driver struct {
	Host     string
	Port     int
	Password string

	channel Channel
	*connection
}

// Connect open a connection via TCP with the sonic server.
func (c *driver) Connect() error {
	if !IsChannelValid(c.channel) {
		return ErrChanName
	}

	var err error
	c.connection, err = newConnection(c)
	return err
}

func (c *driver) Quit() error {
	err := c.write("QUIT")
	if err != nil {
		return err
	}

	// should get ENDED
	_, err = c.read()
	c.close()
	return err
}

func (c driver) Ping() error {
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

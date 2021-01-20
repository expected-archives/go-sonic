package sonic

import (
	"errors"
	"fmt"
)

// ErrActionName is throw when the action is invalid.
var ErrActionName = errors.New("invalid action name")

// Controllable  is used for administration purposes.
type Controllable interface {
	// Trigger an action.
	// Command syntax TRIGGER [<action>]?.
	Trigger(action Action) (err error)

	// Quit refer to the Base interface
	Quit() (err error)

	// Ping refer to the Base interface
	Ping() (err error)
}

// controlChannel is used for administration purposes.
type controlChannel struct {
	*driver
}

// NewControl create a new driver instance with a controlChannel instance.
// Only way to get a Controllable implementation.
func NewControl(host string, port int, password string) (Controllable, error) {
	driver := &driver{
		Host:     host,
		Port:     port,
		Password: password,
		channel:  Control,
	}
	err := driver.Connect()
	if err != nil {
		return nil, err
	}
	return controlChannel{
		driver: driver,
	}, nil
}

func (c controlChannel) Trigger(action Action) (err error) {
	if !IsActionValid(action) {
		return ErrActionName
	}
	err = c.write(fmt.Sprintf("TRIGGER %s", action))
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

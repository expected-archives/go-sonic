package sonic

import (
	"errors"
	"fmt"
)

var ErrActionName = errors.New("invalid action name")

// Controllable  is used for administration purposes.
type Controllable interface {
	// Trigger an action.
	// Command syntax TRIGGER [<action>]?.
	Trigger(action Action) (err error)

	// Quit refer to the Base interface
	Quit() (err error)

	// Quit refer to the Base interface
	Ping() (err error)
}

type ControlChannel struct {
	*Driver
}

func NewControl(host string, port int, password string) (Controllable, error) {
	driver := &Driver{
		Host:     host,
		Port:     port,
		Password: password,
		channel:  Control,
	}
	err := driver.Connect()
	if err != nil {
		return nil, err
	}
	return ControlChannel{
		Driver: driver,
	}, nil
}

func (c ControlChannel) Trigger(action Action) (err error) {
	if IsActionValid(action) {
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

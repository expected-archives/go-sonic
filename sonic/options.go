package sonic

import "time"

type controllerOptions struct {
	Host                string
	Port                int
	Password            string
	PoolMinConnections  int
	PoolMaxConnections  int
	PoolPingThreshold   time.Duration
	PoolMaxIdleLifetime time.Duration
	Channel             Channel
}

func (o controllerOptions) With(optionSetters ...OptionSetter) controllerOptions {
	for _, os := range optionSetters {
		os(&o)
	}

	return o
}

func defaultOptions(
	host string,
	port int,
	password string,
	channel Channel,
) controllerOptions {
	return controllerOptions{
		Host:     host,
		Port:     port,
		Password: password,
		Channel:  channel,

		PoolMinConnections:  1,
		PoolMaxConnections:  16,
		PoolMaxIdleLifetime: 5 * time.Minute,
		PoolPingThreshold:   0,
	}
}

// OptionSetter defines an option setter function.
type OptionSetter func(*controllerOptions)

// OptionPoolMaxConnections sets maximum number of idle connections in the pool.
// By default is 16.
func OptionPoolMaxIdleConnections(val int) OptionSetter {
	return func(o *controllerOptions) {
		o.PoolMaxConnections = val
	}
}

// OptionPoolMinIdleConnections sets minimum number of idle connections in the pool.
// By default is 1.
func OptionPoolMinIdleConnections(val int) OptionSetter {
	return func(o *controllerOptions) {
		o.PoolMinConnections = val
	}
}

// OptionPoolPingThreshold sets a minimum ping interval to ensure that
// the connection is healthy before getting it from the pool.
//
// By default is 0s. For disabling set it to 0.
func OptionPoolPingThreshold(val time.Duration) OptionSetter {
	return func(o *controllerOptions) {
		o.PoolPingThreshold = val
	}
}

// OptionPoolMaxIdleLifetime sets a minimum lifetime of idle connection.
//
// By default is 5m. For disabling set it to 0.
func OptionPoolMaxIdleLifetime(val time.Duration) OptionSetter {
	return func(o *controllerOptions) {
		o.PoolMaxIdleLifetime = val
	}
}

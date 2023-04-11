package sonic

// driversHolder defines base interface around driversPool.
type driversHolder struct {
	*driversPool
}

func newDriversHolder(
	opts controllerOptions,
) (*driversHolder, error) {
	df := driverFactory{
		Host:     opts.Host,
		Port:     opts.Port,
		Password: opts.Password,
		Channel:  opts.Channel,
	}

	dp, err := newDriversPool(
		&df,
		opts.PoolMinConnections,
		opts.PoolMaxConnections,
		opts.PoolPingThreshold,
		opts.PoolMaxIdleLifetime,
	)
	if err != nil {
		return nil, err
	}

	return &driversHolder{
		driversPool: dp,
	}, nil
}

// Quit all connections and close the pool. It never returns an error.
func (c *driversHolder) Quit() error {
	c.driversPool.Close()

	return nil
}

// Ping one connection.
func (c *driversHolder) Ping() error {
	d, err := c.Get()
	if err != nil {
		return err
	}
	defer d.close()

	return d.Ping()
}

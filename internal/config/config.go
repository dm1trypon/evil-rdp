package config

func (c *Config) Create() *Config {
	c = &Config{
		logger: logger{
			level: 1,
		},
		stream: stream{
			numThreads: 1,
			displays:   []int{0},
			delay:      33,
			chunkSize:  50000,
		},
		net: net{
			port: 5959,
		},
	}

	return c
}

func (c *Config) GetNetPort() int {
	return c.net.port
}

func (c *Config) GetLoggerLevel() int {
	return c.logger.level
}

func (c *Config) GetStreamDelay() int {
	return c.stream.delay
}

func (c *Config) GetNumThreads() int {
	return c.stream.numThreads
}

func (c *Config) GetStreamChunkSize() int {
	return c.stream.chunkSize
}

func (c *Config) GetStreamDisplays() []int {
	return c.stream.displays
}

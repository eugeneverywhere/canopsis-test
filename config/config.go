package config

import "github.com/lillilli/logger"

// Config - service configuration
type Config struct {
	DB     DBConfig
	Rabbit RabbitConfig

	Log logger.Params
}

// RabbitConfig - configuration for rabbit connection
type RabbitConfig struct {
	Addr         string
	InputChannel string
	User         string
	Password     string
}

// DBConfig - db connection params
type DBConfig struct {
	Host string `env:"DB_HOST"`
	Port int    `env:"DB_PORT"`
	Name string `env:"DB_NAME"`
}

package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Configurtion struct {
	Environment string `envconfig:"ENV"`
	Host        string `envconfig:"HOST" default:"0.0.0.0"`
	Port        int    `envconfig:"APPLICATION_PORT" default:"9091"`
}

func GetConfiguration() (*Configurtion, error) {
	cfg := &Configurtion{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

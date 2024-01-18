package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	BackendBaseUrl string `mapstructure:"backend_base_url"`
}

func LoadConfig(file string) (Config, error) {
	viper.AutomaticEnv()
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

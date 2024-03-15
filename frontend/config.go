package frontend

import (
	"github.com/benjamonnguyen/opendoorchat/keycloak"
	"github.com/spf13/viper"
)

type Config struct {
	Backend struct {
		BaseUrl string
	}
	Address  string
	Keycloak keycloak.Config
}

func LoadConfig(file string) (Config, error) {
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

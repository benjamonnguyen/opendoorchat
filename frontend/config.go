package frontend

import (
	"github.com/spf13/viper"
)

type Config struct {
	Backend struct {
		BaseUrl string
	}
	Keycloak KeycloakCfg
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

type KeycloakCfg struct {
	BaseUrl      string
	Realm        string
	ClientId     string
	ClientSecret string
}

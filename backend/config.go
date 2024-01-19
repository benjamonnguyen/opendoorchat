package backend

import (
	"log"
	"time"

	"github.com/benjamonnguyen/opendoorchat/keycloak"
	"github.com/spf13/viper"
)

type Config struct {
	Host             string
	Port             int
	Domain           string
	LogLevel         string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	RequestTimeout   time.Duration
	Mongo            MongoConfig
	Kafka            KafkaConfig
	Consumers        struct{}
	MailerSendApiKey string
	Keycloak         keycloak.Config
}

func LoadConfig(in string) Config {
	viper.SetConfigFile(in)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("failed reading config file:", err)
	}

	var cfg Config
	if err := viper.UnmarshalExact(&cfg); err != nil {
		log.Fatalln("failed unmarshalling config:", err)
	}
	return cfg
}

type MongoConfig struct {
	URI      string
	Database string
}

type KafkaConfig struct {
	Brokers        string
	User           string
	Password       string
	MaxPollRecords int
	Topics         struct {
		InboundEmails string
		ChatMessages  string
	}
	LogLevel int
}

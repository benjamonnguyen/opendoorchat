package opendoorchat

import (
	"log"
	"time"

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
}

func LoadConfig(path string) Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
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

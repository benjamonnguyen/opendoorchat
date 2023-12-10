package config

import (
	"log"
	"time"

	"github.com/benjamonnguyen/opendoor-chat/commons/db"
	"github.com/benjamonnguyen/opendoor-chat/commons/mq"
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
	Mongo            db.MongoConfig
	Kafka            mq.KafkaConfig
	Consumers        struct{}
	MailerSendApiKey string
}

func LoadConfig(name, cfgType, path string, cfg any) {
	viper.SetConfigName(name)
	viper.SetConfigType(cfgType)
	viper.AddConfigPath(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("failed reading config file:", err)
	}

	if cfg == nil {
		return
	}
	if err := viper.UnmarshalExact(cfg); err != nil {
		log.Fatalln("failed unmarshalling config:", err)
	}
}

package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ElasticURI string `env:"ELASTICSEARCH_URL" env_default:"http://localhost:9200"`
	RedisHost  string `env:"REDIS_URL"`
	MongoURI   string `env:"MONGO_URL" env_default:"mongodb://mongo1,mongo2,mongo3/?replicaSet=rs0"`
	RabbitMQ   string `env:"RABBITMQ_URL"`
}

var ConfigData = &Config{}

func EnvExtractor() error {

	err := cleanenv.ReadEnv(ConfigData)

	if err != nil {
		return err
	}
	return nil

}

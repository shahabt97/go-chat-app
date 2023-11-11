package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	ElasticURI      string `env:"ELASTICSEARCH_URL" env_default:"http://localhost:9200"`
	ElasticPassword string `env:"ELASTICSEARCH_PASSWORD" env_default:""`
	ElasticUsername string `env:"ELASTICSEARCH_USERNAME" env_default:""`
	RedisHost       string `env:"REDIS_URL" env_default:"localhost:6379"`
	RedisPassword   string `env:"REDIS_PASSWORD"  env_default:""`
	MongoURI        string `env:"MONGO_URL" env_default:"mongodb://mongo1,mongo2,mongo3/?replicaSet=rs0"`
	RabbitMQ        string `env:"RABBITMQ_URL"`
}

var ConfigData = &Config{}

func EnvExtractor() error {

	// load env file to ENV
	if err := godotenv.Load(); err != nil {
		return err
	}

	// load env into Config struct
	if err := cleanenv.ReadEnv(ConfigData); err != nil {
		return err
	}
	return nil

}

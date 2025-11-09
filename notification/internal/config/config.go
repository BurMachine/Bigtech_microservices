package config

type Config struct {
	Postgres Postgres `envPrefix:"PG_"`

	Kafka Kafka `envPrefix:"KAFKA_"`
}

type Postgres struct {
	DbHost     string `env:"HOST" envDefault:"localhost"`
	DbPort     string `env:"PORT" envDefault:"5432"`
	DbUser     string `env:"USER" envDefault:"postgres"`
	DbPassword string `env:"PASSWORD" envDefault:"postgres_pass"`
	DbName     string `env:"NAME" envDefault:"notification_db"`
}

type Kafka struct {
	Brokers       []string `env:"BROKERS" envDefault:"localhost:9092"`
	ConsumerGroup string   `env:"KAFKA_CONSUMER_GROUP" envDefault:"notification-service-group"`
	ConsumerName  string   `env:"KAFKA_CONSUMER_NAME" envDefault:"notification-service-consumer"`
	ConsumerTopic string   `env:"KAFKA_CONSUMER_TOPIC" envDefault:"messages"`
}

type Secrets struct {
	AppToken string `yaml:"appToken" env:"APP_TOKEN"`
	AppName  string `yaml:"appName" env:"APP_NAME"`
}

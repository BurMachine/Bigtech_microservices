package config

type Config struct {
	AddrGrpc        string   `env:"GRPC" envDefault:":8082"`
	UserServicePort string   `env:"USER_SERVICE_PORT" envDefault:":8084"`
	Postgres        Postgres `envPrefix:"PG_"`

	Kafka Kafka `envPrefix:"KAFKA_"`
}

type Postgres struct {
	DbHost     string `env:"HOST" envDefault:"localhost"`
	DbPort     string `env:"PORT" envDefault:"5432"`
	DbUser     string `env:"USER" envDefault:"postgres"`
	DbPassword string `env:"PASSWORD" envDefault:"postgres_pass"`
	DbName     string `env:"NAME" envDefault:"social_db"`
}

type Kafka struct {
	Brokers []string `env:"BROKERS" envDefault:"localhost:9092"`
	Topic   string   `env:"TOPIC" envDefault:"social"`
}

type Secrets struct {
	AppToken string `yaml:"appToken" env:"APP_TOKEN"`
	AppName  string `yaml:"appName" env:"APP_NAME"`
}

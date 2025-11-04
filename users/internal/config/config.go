package config

type Config struct {
	AddrGrpc string `env:"GRPC" envDefault:":8082"`

	Postgres Postgres `envPrefix:"PG_"`
}

type Postgres struct {
	DbHost     string `env:"HOST" envDefault:"localhost"`
	DbPort     string `env:"PORT" envDefault:"5432"`
	DbUser     string `env:"USER" envDefault:"postgres"`
	DbPassword string `env:"PASSWORD" envDefault:"postgres_pass"`
	DbName     string `env:"NAME" envDefault:"users_db"`
}

type Secrets struct {
	AppToken string `yaml:"appToken" env:"APP_TOKEN"`
	AppName  string `yaml:"appName" env:"APP_NAME"`
}

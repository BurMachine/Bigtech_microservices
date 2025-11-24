package config

type Config struct {
	AddrGrpc string   `yaml:"addrGrpc" env:"GRPC" envDefault:":8081"`
	Postgres Postgres `yaml:"postgres" envPrefix:"PG_"`
}

type Postgres struct {
	DbHost     string `yaml:"dbHost" env:"HOST" envDefault:"localhost"`
	DbPort     string `yaml:"dbPort" env:"PORT" envDefault:"5432"`
	DbUser     string `yaml:"dbUser" env:"USER" envDefault:"postgres"`
	DbPassword string `yaml:"dbPassword" env:"PASSWORD" envDefault:"postgres_pass"`
	DbName     string `yaml:"dbName" env:"NAME" envDefault:"auth_db"`
}

type Secrets struct {
	AppToken string `yaml:"appToken" env:"APP_TOKEN"`
	AppName  string `yaml:"appName" env:"APP_NAME"`
}

package config

type Config struct {
	AuthPort   string `envconfig:"AUTH_PORT" envDefault:"8081"`
	ChatPort   string `envconfig:"CHAT_PORT" envDefault:"8082"`
	SocialPort string `envconfig:"SOCICAL_PORT" envDefault:"8083"`
	UserPort   string `envconfig:"USER_PORT" envDefault:"8084"`
}

type Secrets struct {
	AppToken string `yaml:"appToken" env:"APP_TOKEN"`
	AppName  string `yaml:"appName" env:"APP_NAME"`
}

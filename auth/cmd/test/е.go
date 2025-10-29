package main

// Config сложная структура для тестирования
type Config struct {
	// Основные настройки приложения
	AppName     string `yaml:"appName" env:"APP_NAME" envDefault:"my-service"`
	AppVersion  string `yaml:"appVersion" env:"APP_VERSION" envDefault:"1.0.0"`
	Environment string `yaml:"environment" env:"ENVIRONMENT" envDefault:"development"`
	Debug       bool   `yaml:"debug" env:"DEBUG" envDefault:"false"`

	// Настройки сервера
	Server Server `yaml:"server" envPrefix:"SERVER_"`

	// Настройки базы данных
	Database Database `yaml:"database" envPrefix:"DB_"`

	// Настройки Redis
	Redis Redis `yaml:"redis" envPrefix:"REDIS_"`

	// Настройки логирования
	Logger Logger `yaml:"logger" envPrefix:"LOG_"`

	// Настройки JWT
	JWT JWT `yaml:"jwt" envPrefix:"JWT_"`

	// Настройки внешних сервисов
	ExternalServices ExternalServices `yaml:"externalServices"`

	// Настройки очередей
	Queue Queue `yaml:"queue" envPrefix:"QUEUE_"`

	// Лимиты и таймауты
	Limits Limits `yaml:"limits" envPrefix:"LIMIT_"`
}

// Server настройки HTTP/gRPC серверов
type Server struct {
	HTTP HTTP `yaml:"http" envPrefix:"HTTP_"`
	GRPC GRPC `yaml:"grpc" envPrefix:"GRPC_"`
}

type HTTP struct {
	Host            string `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Port            int    `yaml:"port" env:"PORT" envDefault:"8080"`
	ReadTimeout     int    `yaml:"readTimeout" env:"READ_TIMEOUT" envDefault:"30"`
	WriteTimeout    int    `yaml:"writeTimeout" env:"WRITE_TIMEOUT" envDefault:"30"`
	ShutdownTimeout int    `yaml:"shutdownTimeout" env:"SHUTDOWN_TIMEOUT" envDefault:"5"`
	EnableCORS      bool   `yaml:"enableCORS" env:"ENABLE_CORS" envDefault:"true"`
}

type GRPC struct {
	Host string `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Port int    `yaml:"port" env:"PORT" envDefault:"9090"`
}

// Database настройки PostgreSQL
type Database struct {
	Primary            DatabaseConnection `yaml:"primary" envPrefix:"PRIMARY_"`
	Replica            DatabaseConnection `yaml:"replica" envPrefix:"REPLICA_"`
	PoolSize           int                `yaml:"poolSize" env:"POOL_SIZE" envDefault:"25"`
	MaxIdleConnections int                `yaml:"maxIdleConnections" env:"MAX_IDLE_CONNECTIONS" envDefault:"10"`
	ConnMaxLifetime    int                `yaml:"connMaxLifetime" env:"CONN_MAX_LIFETIME" envDefault:"3600"`
}

type DatabaseConnection struct {
	Host     string `yaml:"host" env:"HOST" envDefault:"localhost"`
	Port     string `yaml:"port" env:"PORT" envDefault:"5432"`
	User     string `yaml:"user" env:"USER" envDefault:"postgres"`
	Password string `yaml:"password" env:"PASSWORD" envDefault:""`
	Database string `yaml:"database" env:"DATABASE" envDefault:"mydb"`
	SSLMode  string `yaml:"sslMode" env:"SSL_MODE" envDefault:"disable"`
}

// Redis настройки
type Redis struct {
	Host         string `yaml:"host" env:"HOST" envDefault:"localhost"`
	Port         int    `yaml:"port" env:"PORT" envDefault:"6379"`
	Password     string `yaml:"password" env:"PASSWORD" envDefault:""`
	DB           int    `yaml:"db" env:"DB" envDefault:"0"`
	MaxRetries   int    `yaml:"maxRetries" env:"MAX_RETRIES" envDefault:"3"`
	PoolSize     int    `yaml:"poolSize" env:"POOL_SIZE" envDefault:"10"`
	MinIdleConns int    `yaml:"minIdleConns" env:"MIN_IDLE_CONNS" envDefault:"5"`
}

// Logger настройки логирования
type Logger struct {
	Level         string `yaml:"level" env:"LEVEL" envDefault:"info"`
	Format        string `yaml:"format" env:"FORMAT" envDefault:"json"`
	Output        string `yaml:"output" env:"OUTPUT" envDefault:"stdout"`
	AddCaller     bool   `yaml:"addCaller" env:"ADD_CALLER" envDefault:"true"`
	AddStacktrace bool   `yaml:"addStacktrace" env:"ADD_STACKTRACE" envDefault:"false"`
}

// JWT настройки токенов
type JWT struct {
	SecretKey            string `yaml:"secretKey" env:"SECRET_KEY" envDefault:"my-secret-key"`
	AccessTokenDuration  int    `yaml:"accessTokenDuration" env:"ACCESS_TOKEN_DURATION" envDefault:"900"`
	RefreshTokenDuration int    `yaml:"refreshTokenDuration" env:"REFRESH_TOKEN_DURATION" envDefault:"604800"`
	Issuer               string `yaml:"issuer" env:"ISSUER" envDefault:"my-service"`
}

// ExternalServices настройки внешних API
type ExternalServices struct {
	PaymentGateway PaymentGateway `yaml:"paymentGateway" envPrefix:"PAYMENT_"`
	EmailService   EmailService   `yaml:"emailService" envPrefix:"EMAIL_"`
	S3Storage      S3Storage      `yaml:"s3Storage" envPrefix:"S3_"`
}

type PaymentGateway struct {
	BaseURL    string `yaml:"baseUrl" env:"BASE_URL" envDefault:"https://api.payment.com"`
	APIKey     string `yaml:"apiKey" env:"API_KEY" envDefault:""`
	Timeout    int    `yaml:"timeout" env:"TIMEOUT" envDefault:"30"`
	MaxRetries int    `yaml:"maxRetries" env:"MAX_RETRIES" envDefault:"3"`
}

type EmailService struct {
	Provider  string `yaml:"provider" env:"PROVIDER" envDefault:"sendgrid"`
	APIKey    string `yaml:"apiKey" env:"API_KEY" envDefault:""`
	FromEmail string `yaml:"fromEmail" env:"FROM_EMAIL" envDefault:"noreply@example.com"`
}

type S3Storage struct {
	Endpoint        string `yaml:"endpoint" env:"ENDPOINT" envDefault:"s3.amazonaws.com"`
	Region          string `yaml:"region" env:"REGION" envDefault:"us-east-1"`
	AccessKeyID     string `yaml:"accessKeyId" env:"ACCESS_KEY_ID" envDefault:""`
	SecretAccessKey string `yaml:"secretAccessKey" env:"SECRET_ACCESS_KEY" envDefault:""`
	BucketName      string `yaml:"bucketName" env:"BUCKET_NAME" envDefault:"my-bucket"`
}

// Queue настройки очередей (RabbitMQ/Kafka)
type Queue struct {
	Type     string         `yaml:"type" env:"TYPE" envDefault:"rabbitmq"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq" envPrefix:"RABBITMQ_"`
	Kafka    KafkaConfig    `yaml:"kafka" envPrefix:"KAFKA_"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host" env:"HOST" envDefault:"localhost"`
	Port     int    `yaml:"port" env:"PORT" envDefault:"5672"`
	User     string `yaml:"user" env:"USER" envDefault:"guest"`
	Password string `yaml:"password" env:"PASSWORD" envDefault:"guest"`
	VHost    string `yaml:"vhost" env:"VHOST" envDefault:"/"`
}

type KafkaConfig struct {
	Brokers       string `yaml:"brokers" env:"BROKERS" envDefault:"localhost:9092"`
	ConsumerGroup string `yaml:"consumerGroup" env:"CONSUMER_GROUP" envDefault:"my-service"`
	Topic         string `yaml:"topic" env:"TOPIC" envDefault:"events"`
}

// Limits лимиты и таймауты
type Limits struct {
	MaxRequestSize    int `yaml:"maxRequestSize" env:"MAX_REQUEST_SIZE" envDefault:"10485760"`
	RateLimitPerMin   int `yaml:"rateLimitPerMin" env:"RATE_LIMIT_PER_MIN" envDefault:"100"`
	MaxConnections    int `yaml:"maxConnections" env:"MAX_CONNECTIONS" envDefault:"1000"`
	RequestTimeout    int `yaml:"requestTimeout" env:"REQUEST_TIMEOUT" envDefault:"30"`
	MaxConcurrentJobs int `yaml:"maxConcurrentJobs" env:"MAX_CONCURRENT_JOBS" envDefault:"50"`
}

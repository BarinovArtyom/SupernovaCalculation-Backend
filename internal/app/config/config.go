package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	Minio       `toml:"minio"`
	JWT         JWTConfig   `toml:"JWT"`
	Redis       RedisConfig `toml:"redis"`
	Async       AsyncConfig `toml:"async_service"`
}

type Minio struct {
	User     string `toml:"user"`
	Pass     string `toml:"pass"`
	Endpoint string `toml:"endpoint"`
}

type JWTConfig struct {
	Token         string `toml:"Token"`
	SigningMethod jwt.SigningMethod
	ExpiresIn     time.Duration `toml:"Expires-in"`
}

type RedisConfig struct {
	Host        string `toml:"REDIS_HOST"`
	Password    string `toml:"REDIS_PASSWORD"`
	Port        int    `toml:"REDIS_PORT"`
	User        string `toml:"REDIS_USER"`
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type AsyncConfig struct {
	URL   string `toml:"url" env:"ASYNC_SERVICE_URL"`
	Token string `toml:"token" env:"ASYNC_SERVICE_TOKEN"`
}

const (
	envRedisHost = "REDIS_HOST"
	envRedisPort = "REDIS_PORT"
	envRedisUser = "REDIS_USER"
	envRedisPass = "REDIS_PASSWORD"
)

func NewConfig(log *log.Logger) (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	cfg.Redis.Host = os.Getenv(envRedisHost)
	cfg.Redis.Port, err = strconv.Atoi(os.Getenv(envRedisPort))

	if err != nil {
		return nil, fmt.Errorf("redis port must be int value: %w", err)
	}

	cfg.Redis.Password = os.Getenv(envRedisPass)
	cfg.Redis.User = os.Getenv(envRedisUser)
	cfg.JWT.SigningMethod = jwt.SigningMethodHS256
	cfg.JWT.ExpiresIn = 12 * time.Hour

	if cfg.Async.URL == "" {
		cfg.Async.URL = "http://localhost:8000"
	}
	if cfg.Async.Token == "" {
		cfg.Async.Token = "secret123"
	}

	log.Info("config parsed")
	log.Info(cfg.ServiceHost)
	log.Info(cfg.ServicePort)
	log.Info(cfg.Minio)
	log.Info(cfg.Redis)
	log.Info(cfg.JWT)
	log.Info("Async service URL: ", cfg.Async.URL)
	log.Info("Async service token configured: ", len(cfg.Async.Token) > 0)

	return cfg, nil
}

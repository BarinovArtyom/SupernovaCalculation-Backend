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

// Config Структура конфигурации;
// Содержит все конфигурационные данные о сервисе;
// автоподгружается при изменении исходного файла
type Config struct {
	ServiceHost string
	ServicePort int
	Minio       `toml:"minio"`
	JWT         JWTConfig   `toml:"JWT"`
	Redis       RedisConfig `toml:"redis"`
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

const (
	envRedisHost = "REDIS_HOST"
	envRedisPort = "REDIS_PORT"
	envRedisUser = "REDIS_USER"
	envRedisPass = "REDIS_PASSWORD"
)

// NewConfig Создаёт новый объект конфигурации, загружая данные из файла конфигурации
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
	cfg.JWT.SigningMethod = jwt.SigningMethodHS256 //can change inline
	cfg.JWT.ExpiresIn = 12 * time.Hour

	log.Info("config parsed")
	log.Info(cfg.ServiceHost)
	log.Info(cfg.ServicePort)
	log.Info(cfg.Minio)
	log.Info(cfg.Redis)
	log.Info(cfg.JWT)

	return cfg, nil
}

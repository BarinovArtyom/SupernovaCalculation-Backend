package dsn

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// FromEnv собирает DSN строку из переменных окружения
func FromEnv() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}
	host, existHost := os.LookupEnv("DB_HOST")
	port, existPort := os.LookupEnv("DB_PORT")
	user, existUser := os.LookupEnv("DB_USER")
	pass, existPass := os.LookupEnv("DB_PASS")
	dbname, existName := os.LookupEnv("DB_NAME")
	if !existHost || !existName || !existPass || !existPort || !existUser {
		return "", fmt.Errorf("env var is empty")
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname), nil
}

func GetCurrentUserID() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}
	id, exists := os.LookupEnv("CURRENT_SESSION")
	if !exists {
		return "", fmt.Errorf("user env var not read")
	}
	if id == "null" {
		return "null", fmt.Errorf("no active user")
	}
	return id, nil
}

func SetCurrentUserID(id string) error {
	return os.Setenv("CURRENT_SESSION", id)
}

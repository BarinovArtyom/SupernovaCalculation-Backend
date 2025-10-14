package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"lab1/internal/app/minio"

	"github.com/go-redis/redis/v8"
)

type Repository struct {
	db          *gorm.DB
	logger      *logrus.Logger
	minioclient *minio.MinioClient
	RedisClient *redis.Client
}

func New(dsn string, log *logrus.Logger, m *minio.MinioClient, r *redis.Client) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:          db,
		logger:      log,
		minioclient: m,
		RedisClient: r,
	}, nil
}

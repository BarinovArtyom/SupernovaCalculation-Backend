package minio

import (
	"context"
	"lab1/internal/app/config"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	*minio.Client
}

func NewMinioClient(conf *config.Config) *MinioClient {
    useSSL := false

    minioClient, err := minio.New(conf.Minio.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(conf.Minio.User, conf.Minio.Pass, ""),
        Secure: useSSL,
    })
    if err != nil {
        log.Fatalln(err)
    }
    
    // Проверка подключения
    _, err = minioClient.ListBuckets(context.Background())
    if err != nil {
        log.Fatalf("Failed to connect to MinIO: %v", err)
    }
    
    log.Println("Successfully connected to MinIO")
    return &MinioClient{
        minioClient,
    }
}

package main

import (
	"fmt"

	"lab1/internal/app/config"
	"lab1/internal/app/dsn"
	"lab1/internal/app/handler"
	"lab1/internal/app/minio"
	redis_api "lab1/internal/app/redis"
	"lab1/internal/app/repository"
	"lab1/internal/pkg"

	_ "lab1/docs"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title Energy Calculation System
// @version 1.0
// @description Bmstu Open IT Platform
// @contact.name API Support
// @contact.url https://vk.com/aaaaaaeaaaa
// @contact.email Barinovartem383@gmail.com
// @license.name AS IS (NO WARRANTY)
// @host 127.0.0.1
// @schemes http
// @BasePath /

func main() {
	logger := logrus.New()
	router := gin.Default()
	router.Use(handler.CORSMiddleware())
	conf, err := config.NewConfig(logger)
	minioClient := minio.NewMinioClient(conf)

	if err != nil {
		logger.Fatalf("Error with configuration reading: #{err}")
	}
	redisClient, err := redis_api.New(conf.Redis)
	if err != nil {
		logger.Fatalf("Error creating redis: #{err}")
	}
	postgresString, errPost := dsn.FromEnv()

	if errPost != nil {
		logger.Fatalf("Error with reading postgres line: #{err}")
	}
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString, logger, minioClient, redisClient)
	if errRep != nil {
		logger.Fatalf("Error from repo: #{err}")
	}

	hand := handler.NewHandler(logger, rep, minioClient, conf)
	application := pkg.NewApp(conf, router, logger, hand)
	application.RunApp()
}

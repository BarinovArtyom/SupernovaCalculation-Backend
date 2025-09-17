package api

import (
	"log"

	"lab1/internal/app/handler"
	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	r.GET("/home", handler.GetScopes)

	r.GET("/details/:id", handler.GetScopeDetails)

	r.GET("/current_request/:id", handler.GetCurrentRequest)

	r.Run()

	log.Println("Server down")
}

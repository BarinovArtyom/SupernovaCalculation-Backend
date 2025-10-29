package pkg

import (
	"lab1/internal/app/config"
	"lab1/internal/app/handler"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Logger  *logrus.Logger
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, l *logrus.Logger, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Logger:  l,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	a.Logger.Info("Server start up")
	a.Handler.RegisterHandler(a.Router)
	serverAdress := a.Config.ServiceHost + ":" + strconv.Itoa(a.Config.ServicePort)
	if err := a.Router.Run(serverAdress); err != nil {
		a.Logger.Fatalln(err)
	}
	a.Logger.Info("Server down")
}

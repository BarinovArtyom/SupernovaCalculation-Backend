package handler

import (
	"lab1/internal/app/minio"
	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Logger     *logrus.Logger
	Repository *repository.Repository
}

func NewHandler(l *logrus.Logger, r *repository.Repository, m *minio.MinioClient) *Handler {
	return &Handler{
		Logger:     l,
		Repository: r,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// Домен услуги
	router.GET("/scopes", h.GetScopes)                 // GET список с фильтрацией
	router.GET("/scope/:id", h.GetScopeByID)           // GET одна запись
	router.POST("/scope/addtostar", h.AddToStar)       // POST добавления в заявку-черновик
	router.POST("/scope/add", h.AddScope)              // POST добавление (без изображения)
	router.POST("/scope/addpicture/:id", h.AddPicture) // POST добавление изображения
	router.PUT("/scope/edit/:id", h.EditScope)         // PUT изменение
	router.DELETE("/scope/delete/:id", h.DeleteScope)  // DELETE удаление

	// Домен заявки
	router.GET("/stars", h.GetStars)                        // GET список
	router.GET("/star/:id", h.GetCurrentEnergyCalculations) // GET одна запись (поля заявки + ее услуги)
	router.GET("/star/active", h.GetStarIcon)               // GET иконки корзины
	router.PUT("/star/edit/:id", h.EditStar)                // PUT изменения полей заявки по теме
	router.PUT("/star/form/:id", h.FormStar)                // PUT сформировать создателем
	router.PUT("/star/finish/:id", h.FinishStar)            // PUT завершить/отклонить модератором
	router.POST("/star/ldelete/:id", h.DeleteStar)          //
	router.DELETE("/star/delete/:id", h.DeleteStar)         // DELETE удаление

	// Домен м-м
	router.PUT("/calc/edit/:star_id/:scope_id", h.EditCalcInStar)          // PUT изменение
	router.DELETE("/calc/delete/:star_id/:scope_id", h.DeleteCalcFromStar) // DELETE удаление из заявки (без PK м-м)

	// Домен пользователь
	router.POST("/user/register", h.RegisterUser) // POST регистрация
	router.POST("/user/login", h.LoginUser)       // POST аутентификация
	router.POST("/user/logout", h.LogoutUser)     // POST деавторизация
	router.GET("/user/profile", h.GetUserProfile) // GET полей пользователя после аутентификации (для личного кабинета)
	router.PUT("/user/edit", h.UpdateUser)        // PUT пользователя (личный кабинет)
}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

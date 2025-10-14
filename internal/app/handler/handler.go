package handler

import (
	"lab1/internal/app/config"
	"lab1/internal/app/minio"
	"lab1/internal/app/repository"
	"lab1/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Logger     *logrus.Logger
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(l *logrus.Logger, r *repository.Repository, m *minio.MinioClient, c *config.Config) *Handler {
	return &Handler{
		Logger:     l,
		Repository: r,
		Config:     c,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Домен услуги
	router.GET("/scopes", h.GetScopes)                                                       // GET список с фильтрацией
	router.GET("/scope/:id", h.GetScopeByID)                                                 // GET одна запись
	router.POST("/scope/addtostar", h.WithAuthCheck(role.Moderator, role.User), h.AddToStar) // POST добавления в заявку-черновик
	router.POST("/scope/add", h.WithAuthCheck(role.Moderator), h.AddScope)                   // POST добавление (без изображения)
	router.POST("/scope/addpicture/:id", h.WithAuthCheck(role.Moderator), h.AddPicture)      // POST добавление изображения
	router.PUT("/scope/edit/:id", h.WithAuthCheck(role.Moderator), h.EditScope)              // PUT изменение
	router.DELETE("/scope/delete/:id", h.WithAuthCheck(role.Moderator), h.DeleteScope)       // DELETE удаление

	// Домен заявки
	router.GET("/stars", h.WithAuthCheck(role.Moderator, role.User), h.GetStars)                        // GET список
	router.GET("/star/:id", h.WithAuthCheck(role.Moderator, role.User), h.GetCurrentEnergyCalculations) // GET одна запись (поля заявки + ее услуги)
	router.GET("/star/active", h.WithAuthCheck(role.Moderator, role.User), h.GetStarIcon)               // GET иконки корзины
	router.PUT("/star/edit/:id", h.WithAuthCheck(role.Moderator, role.User), h.EditStar)                // PUT изменения полей заявки по теме
	router.PUT("/star/form/:id", h.WithAuthCheck(role.Moderator, role.User), h.FormStar)                // PUT сформировать создателем
	router.PUT("/star/finish/:id", h.WithAuthCheck(role.Moderator), h.FinishStar)                       // PUT завершить/отклонить модератором
	router.POST("/star/ldelete/:id", h.DeleteStar)                                                      //
	router.DELETE("/star/delete/:id", h.WithAuthCheck(role.Moderator, role.User), h.DeleteStar)         // DELETE удаление

	// Домен м-м
	router.PUT("/calc/edit/:star_id/:scope_id", h.WithAuthCheck(role.Moderator, role.User), h.EditCalcInStar)          // PUT изменение
	router.DELETE("/calc/delete/:star_id/:scope_id", h.WithAuthCheck(role.Moderator, role.User), h.DeleteCalcFromStar) // DELETE удаление из заявки (без PK м-м)

	// Домен пользователь
	router.POST("/user/register", h.RegisterUser)                                                     // POST регистрация
	router.POST("/user/login", h.LoginUser)                                                           // POST аутентификация
	router.POST("/user/logout", h.WithAuthCheck(role.Moderator, role.User, role.Guest), h.LogoutUser) // POST деавторизация
	router.GET("/user/profile", h.WithAuthCheck(role.Moderator, role.User), h.GetUserProfile)         // GET полей пользователя после аутентификации (для личного кабинета)
	router.PUT("/user/edit", h.WithAuthCheck(role.Moderator, role.User), h.UpdateUser)                // PUT пользователя (личный кабинет)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")
}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

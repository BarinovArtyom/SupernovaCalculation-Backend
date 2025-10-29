package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"lab1/internal/app/ds"
	"lab1/internal/app/dsn"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetScopes(ctx *gin.Context) {
	var scopes []ds.Scope
	var err error

	searchQuery := ctx.Query("search") // получаем значение из нашего поля
	if searchQuery == "" {             // если поле поиска пусто, то просто получаем из репозитория все записи
		scopes, err = h.Repository.GetScopes()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		scopes, err = h.Repository.GetScopesByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	user_idStr, _ := dsn.GetCurrentUserID()
	user_id, _ := strconv.Atoi(user_idStr)
	star_id, err := h.Repository.GetActiveStar(user_id)
	if err != nil {
		logrus.Error("Ошибка получения id звезды", err)
		star_id = 0
	}

	calccount, err := h.Repository.CalcsInStar(star_id)
	if err != nil {
		logrus.Error("Ошибка получения расчетов", err)
		calccount = 0
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Scopes":    scopes,
		"Star_ID":   star_id,
		"CalcCount": calccount,
		"query":     searchQuery, // передаем введенный запрос обратно на страницу
		// в ином случае оно будет очищаться при нажатии на кнопку
	})
}

func (h *Handler) GetScopeByID(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	specs, err := h.Repository.GetScopeByID(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Specs": specs,
	})
}

func (h *Handler) AddToStar(c *gin.Context) {
	scopeID := c.PostForm("scope_id")
	if scopeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope_id is required"})
		return
	}

	scopeIDInt, err := strconv.Atoi(scopeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope_id"})
		return
	}

	err = h.Repository.AddToStar(scopeIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusSeeOther, "/scopes")
}

func (h *Handler) AddScope(ctx *gin.Context) {
	var scope ds.Scope

	if err := ctx.BindJSON(&scope); err != nil {
		ctx.JSON(http.StatusBadRequest, "неверные данные")
		return
	}

	id, err := h.Repository.CreateScope(&scope)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	scope.ID = id
	ctx.JSON(http.StatusCreated, scope)
}

func (h *Handler) EditScope(ctx *gin.Context) {
	var scope ds.Scope
	id := ctx.Param("id")

	if err := ctx.BindJSON(&scope); err != nil {
		ctx.JSON(http.StatusBadRequest, "incorrect JSON format")
		return
	}

	intid, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "incorrect id format")
		return
	}

	scope.ID = intid
	err = h.Repository.EditScope(&scope)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error: ": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, scope)
}

func (h *Handler) DeleteScope(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Получаем scope чтобы узнать имя файла изображения
	scope, err := h.Repository.GetScopeByID(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ДОБАВЬТЕ ЛОГИРОВАНИЕ
	logrus.Infof("Deleting scope ID: %d, ImgLink: '%s'", id, scope.ImgLink)

	imageName := fmt.Sprintf("%s.jpg", idStr)

	// Удаляем изображение если оно есть
	if scope.ImgLink != "" {
		logrus.Infof("Attempting to delete image: %s", scope.ImgLink)
		err = h.Repository.DeletePicture(id, imageName)
		if err != nil {
			logrus.Errorf("Failed to delete image: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logrus.Infof("Successfully deleted image: %s", scope.ImgLink)
	} else {
		logrus.Info("No image to delete (ImgLink is empty)")
	}

	// Удаляем сам scope
	err = h.Repository.DeleteScope(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Scope (id-%d) deleted", id)})
}

func (h *Handler) AddPicture(ctx *gin.Context) {
	scope_id := ctx.Param("id")
	file, fileHeader, err := ctx.Request.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "failed to upload image", "error": err.Error()})
		return
	}
	defer file.Close()

	imageName := fmt.Sprintf("%s.jpg", scope_id)

	err = h.Repository.UploadPicture(scope_id, imageName, file, fileHeader.Size)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Image successfully uploaded"})
}

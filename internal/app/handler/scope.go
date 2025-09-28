package handler

import (
	"net/http"
	"strconv"

	"lab1/internal/app/ds"

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

	star_id, err := h.Repository.GetActiveStar()
	if err != nil {
		logrus.Error("Ошибка получения id звезды", err)
		star_id = 0
	}

	calccount, err := h.Repository.CalcsInStar(star_id)
	if err != nil {
		logrus.Error("Ошибка получения расчетов", err)
		calccount = 0
	}

	ctx.HTML(http.StatusOK, "scopes.tmpl", gin.H{
		"Title":     "Scopes",
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

	ctx.HTML(http.StatusOK, "specs.tmpl", gin.H{
		"Title": specs.Name,
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

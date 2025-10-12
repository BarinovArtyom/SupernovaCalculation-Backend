package handler

import (
	"lab1/internal/app/ds"
	"lab1/internal/app/dsn"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req ds.Users

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if req.Login == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, "Login password and status are required")
		return
	}

	if err := h.Repository.CreateUser(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":         req.ID,
		"login":      req.Login,
		"mod_status": req.ModStatus,
	})
}

func (h *Handler) GetUserProfile(ctx *gin.Context) {
	userID, err := dsn.GetCurrentUserID()
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"login":      user.Login,
		"mod_status": user.ModStatus,
	})
}

func (h *Handler) UpdateUser(ctx *gin.Context) {
	userID, err := dsn.GetCurrentUserID()
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Создаем объект пользователя для обновления
	updateUser := ds.Users{
		Login:    req.Login,
		Password: req.Password,
	}

	err = h.Repository.UpdateUser(updateUser, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":    userID,
		"login": req.Login,
	})
}

func (h *Handler) LoginUser(ctx *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if req.Login == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, "Login and password are required")
		return
	}

	id, err := h.Repository.Login(req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.Repository.GetUserByID(strconv.Itoa(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"login":      user.Login,
		"mod_status": user.ModStatus,
	})
}

func (h *Handler) LogoutUser(ctx *gin.Context) {
	if err := h.Repository.Logout(); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

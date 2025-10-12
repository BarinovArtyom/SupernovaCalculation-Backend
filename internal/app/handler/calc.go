package handler

import (
	"fmt"
	"lab1/internal/app/dsn"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteCalcFromStar(ctx *gin.Context) {
	scopeID := ctx.Param("star_id")
	starID := ctx.Param("scope_id")

	userID, err := dsn.GetCurrentUserID()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "no user authenticated")
		return
	}

	intUID, err := strconv.Atoi(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "invalid user id")
		return
	}

	intStarID, err := strconv.Atoi(starID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "invalid star id")
		return
	}

	star, err := h.Repository.GetStarByID(intStarID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "star does not exist")
		return
	}

	if star.UserID != intUID {
		ctx.JSON(http.StatusBadRequest, "attempt to delete unowned star")
		return
	}

	if err := h.Repository.DeleteCalcFromStar(starID, scopeID); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// Используем ваш существующий метод для подсчета расчетов
	ln, err := h.Repository.CalcsInStar(intStarID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Если расчетов не осталось - удаляем всю заявку (помечаем как deleted)
	if ln == 0 {
		if err := h.Repository.DeleteStar(intStarID); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("calc deleted from star (%s) and star removed", starID)})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("calc deleted from star (%s)", starID)})
	}
}

func (h *Handler) EditCalcInStar(ctx *gin.Context) {
	scopeIDStr := ctx.Param("scope_id")
	starIDStr := ctx.Param("star_id")

	scopeID, err := strconv.Atoi(scopeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Invalid scope_id")
		return
	}

	starID, err := strconv.Atoi(starIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Invalid star_id")
		return
	}

	var input struct {
		InpMass float64 `json:"inp_mass" binding:"required"`
		InpTexp float64 `json:"inp_texp" binding:"required"`
		InpDist float64 `json:"inp_dist" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = h.Repository.EditCalcInStar(starID, scopeID, input.InpMass, input.InpTexp, input.InpDist)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Error changing calculation")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Calculation updated successfully"})
}

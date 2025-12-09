package handler

import (
	"fmt"
	async "lab1/internal/app/async"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteCalcFromStar godoc
// @Summary Delete calculation from star
// @Description Delete calculation from user's star. If no calculations remain, the star is deleted.
// @Tags Calc
// @Param star_id path string true "Star ID"
// @Param scope_id path string true "Scope ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calc/delete/{star_id}/{scope_id} [delete]
func (h *Handler) DeleteCalcFromStar(ctx *gin.Context) {
	scopeID := ctx.Param("scope_id")
	starID := ctx.Param("star_id")
	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("error retrieving token payload: %s", err))
		return
	}
	userID := payload.UserID

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

	if star.UserID != userID {
		ctx.JSON(http.StatusBadRequest, "attempt to delete unowned star")
		return
	}

	if err := h.Repository.DeleteCalcFromStar(starID, scopeID); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ln, err := h.Repository.CalcsInStar(intStarID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

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

// EditCalcInStar godoc
// @Summary Edit calculation in star
// @Description Update calculation parameters in user's star
// @Tags Calc
// @Param star_id path string true "Star ID"
// @Param scope_id path string true "Scope ID"
// @Param input body ds.Calc true "Calculation parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calc/edit/{star_id}/{scope_id} [put]
func (h *Handler) EditCalcInStar(ctx *gin.Context) {
	scopeIDStr := ctx.Param("scope_id")
	starIDStr := ctx.Param("star_id")
	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("error retrieving token payload: %s", err))
		return
	}
	userID := payload.UserID

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

	star, err := h.Repository.GetStarByID(starID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "star does not exist")
		return
	}

	if star.UserID != userID {
		ctx.JSON(http.StatusBadRequest, "attempt to edit unowned star")
		return
	}

	err = h.Repository.EditCalcInStar(starID, scopeID, input.InpMass, input.InpTexp, input.InpDist)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, "Error changing calculation")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Calculation updated successfully"})
}

// ReceiveCalculationResult godoc
// @Summary Receive calculation result from async service
// @Description Callback endpoint for async service to send calculation results
// @Tags Calculation
// @Param input body async.CalculationResult true "Calculation result"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calc/ [put]
func (h *Handler) ReceiveCalculationResult(ctx *gin.Context) {
	// Проверка авторизационного токена
	authHeader := ctx.GetHeader("Authorization")
	expectedToken := "Token " + h.Config.Async.Token

	if authHeader != expectedToken {
		h.Logger.Warnf("Unauthorized callback attempt: %s", authHeader)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
		return
	}

	var result async.CalculationResult
	if err := ctx.BindJSON(&result); err != nil {
		h.Logger.Errorf("Failed to bind result: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	h.Logger.Infof("Received calculation result: %+v", result)

	// Обновляем расчет в базе данных
	if err := h.Repository.UpdateCalculationResult(result); err != nil {
		h.Logger.Errorf("Failed to update calculation: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update calculation"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Calculation updated"})
}

package handler

import (
	"errors"
	"fmt"
	"lab1/internal/app/ds"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetCurrentEnergyCalculations godoc
// @Summary Get star with calculations
// @Description Get detailed star information with all scope calculations
// @Tags Star
// @Produce json
// @Param id path string true "Star ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /star/{id} [get]
func (h *Handler) GetCurrentEnergyCalculations(ctx *gin.Context) {
	id := ctx.Param("id")
	star_id, err := strconv.Atoi(id)
	if err != nil || star_id < 0 {
		logrus.Error(err)
	}

	current_star, err := h.Repository.GetStarByID(star_id)
	if err != nil {
		logrus.Error(err)
	}

	matching_scopes, err := h.Repository.GetScopesByStar(star_id)
	if err != nil {
		logrus.Error(err)
	}

	matching_calcs, err := h.Repository.GetCalcsByStar(star_id)
	if err != nil {
		logrus.Error(err)
	}

	calcMap := make(map[int]ds.Calc)
	for _, calc := range matching_calcs {
		calcMap[calc.ScopeID] = calc
	}

	type ScopeWithCalc struct {
		ds.Scope
		Calc ds.Calc
	}

	var scopesWithCalcs []ScopeWithCalc
	for _, scope := range matching_scopes {
		if calc, exists := calcMap[scope.ID]; exists {
			scopesWithCalcs = append(scopesWithCalcs, ScopeWithCalc{
				Scope: scope,
				Calc:  calc,
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"current_star": current_star,
		"scopes":       scopesWithCalcs,
	})
}

// DeleteStar godoc
// @Summary Delete star
// @Description Delete star and redirect to scopes page
// @Tags Star
// @Param id path string true "Star ID"
// @Success 303 "Redirect to /scopes"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /star/delete/{id} [delete]
func (h *Handler) DeleteStar(ctx *gin.Context) {
	starIDStr := ctx.Param("id") // Получаем параметр из пути
	starID, err := strconv.Atoi(starIDStr)
	if err != nil || starID < 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid star id: %s", starIDStr))
		return
	}

	err = h.Repository.DeleteStar(starID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/scopes")
}

// GetStarIcon godoc
// @Summary Get active star info
// @Description Get current user's active star ID and calculation count
// @Tags Star
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /star/active [get]
func (h *Handler) GetStarIcon(ctx *gin.Context) {
	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("error retrieving token payload: %s", err))
		return
	}
	userID := payload.UserID
	starID, err := h.Repository.GetActiveStar(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, gin.H{
				"star_id": 0,
				"count":   0,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	count, err := h.Repository.CalcsInStar(starID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"star_id": starID,
		"count":   count,
	})
}

// GetStars godoc
// @Summary Get stars list
// @Description Get stars with role-based access control.
// @Description - Moderators (modstatus=2): see all stars except "active" and "deleted" statuses
// @Description - Users (modstatus=1): see only their own stars except "active" and "deleted" statuses
// @Description - Guests (modstatus=0): access denied (401)
// @Tags Star
// @Param status query string false "Status filter (applies to non-active, non-deleted stars)"
// @Param start_date query string false "Start date filter (2006-Jan-02)"
// @Param end_date query string false "End date filter (2006-Jan-02)"
// @Success 200 {array} ds.Star
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /stars [get]
func (h *Handler) GetStars(ctx *gin.Context) {
	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("error retrieving token payload: %s", err))
		return
	}

	status := ctx.Query("status")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, err = time.Parse("2006-Jan-02", startDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "Invalid start date format")
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-Jan-02", endDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "Invalid end date format")
			return
		}
	}

	starsPtr, err := h.Repository.GetStars(payload.Role, payload.UserID, status, startDateStr != "", endDateStr != "", startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// Разыменовываем указатель на срез
	stars := *starsPtr

	// Добавляем информацию о завершенных расчетах
	for i := range stars {
		starID := fmt.Sprintf("%d", stars[i].ID)

		// Получаем количество завершенных расчетов (у которых есть результат)
		completedCount, err := h.Repository.GetCompletedCalculationsCount(starID)
		if err != nil {
			h.Logger.Warnf("Failed to get completed calculations for star %d: %v", stars[i].ID, err)
			stars[i].CompletedCalculations = 0
		} else {
			stars[i].CompletedCalculations = completedCount
		}
	}

	ctx.JSON(http.StatusOK, stars)
}

// EditStar godoc
// @Summary Edit star
// @Description Update star name and constellation
// @Tags Star
// @Param id path string true "Star ID"
// @Param input body ds.Star true "Star data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /star/edit/{id} [put]
func (h *Handler) EditStar(ctx *gin.Context) {
	id := ctx.Param("id")

	var input struct {
		Name          string `json:"name" binding:"required"`
		Constellation string `json:"constellation" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Repository.EditStar(id, input.Name, input.Constellation); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Star updated successfully"})
}

// FormStar godoc
// @Summary Form star
// @Description Change star status to formed and update calculations
// @Tags Star
// @Accept json
// @Produce json
// @Param id path string true "Star ID"
// @Param input body map[string]string true "Calculation values"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /star/form/{id} [put]
func (h *Handler) FormStar(ctx *gin.Context) {
	id := ctx.Param("id")

	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("error retrieving token payload: %s", err)})
		return
	}
	creatorID := payload.UserID

	// Получаем JSON данные вместо FormData
	var requestData map[string]string
	if err := ctx.BindJSON(&requestData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Фильтруем только calc_values_
	calcValues := make(map[string]string)
	for key, value := range requestData {
		if strings.HasPrefix(key, "calc_values_") {
			scopeID := strings.TrimPrefix(key, "calc_values_")
			calcValues[scopeID] = value
		}
	}

	if err := h.Repository.FormStar(id, creatorID, calcValues); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Status changed and calculations updated"})
}

// FinishStar godoc
// @Summary Finish star
// @Description Complete or deny star (moderator only)
// @Tags Star
// @Param id path string true "Star ID"
// @Param input body map[string]string true "Status: completed or denied"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /star/finish/{id} [put]
func (h *Handler) FinishStar(ctx *gin.Context) {
	idStr := ctx.Param("id")
	_, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid star ID"})
		return
	}

	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error retrieving token payload"})
		return
	}
	modID := payload.UserID

	isMod, err := h.Repository.ModStatusCheck(modID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if isMod < 2 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "attempt to finish star as regular user"})
		return
	}

	star, calcs, err := h.Repository.GetStarWithCalculations(idStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get star data"})
		return
	}

	if star.Status != "formed" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Звезда не сформирована"})
		return
	}

	var request struct {
		Status string `json:"status"`
	}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	status := request.Status

	if status == "completed" {
		if err := h.Repository.SetStarStatus(idStr, "completed", modID); err != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		h.Repository.SendCalculationsToAsyncService(calcs, h.AsyncClient, idStr)

		ctx.JSON(http.StatusOK, gin.H{
			"status":             "completed",
			"message":            "Star completed successfully, calculations started",
			"calculations_count": len(calcs),
		})

	} else if status == "declined" {
		if err := h.Repository.SetStarStatus(idStr, status, modID); err != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": status})
	} else {
		ctx.JSON(http.StatusConflict, gin.H{"error": "attempt to finish star with wrong status"})
	}
}

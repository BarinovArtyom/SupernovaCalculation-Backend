package handler

import (
	"errors"
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/dsn"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

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

	if current_star.Status != "active" {
		ctx.JSON(http.StatusBadRequest, "Звезда неактивна")
		return
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

func (h *Handler) GetStarIcon(ctx *gin.Context) {
	user_idStr, _ := dsn.GetCurrentUserID()
	user_id, _ := strconv.Atoi(user_idStr)
	starID, err := h.Repository.GetActiveStar(user_id)
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

	// Возвращаем результат
	ctx.JSON(http.StatusOK, gin.H{
		"star_id": starID,
		"count":   count,
	})
}

func (h *Handler) GetStars(ctx *gin.Context) {
	status := ctx.Query("status")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	var startDate, endDate time.Time
	var err error

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

	stars, err := h.Repository.GetStars(status, startDateStr != "", endDateStr != "", startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, stars)
}

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

func (h *Handler) FormStar(ctx *gin.Context) {
	id := ctx.Param("id")

	creatorID, err := dsn.GetCurrentUserID()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctx.Request.ParseForm(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	calcValues := make(map[string]string)
	for key, values := range ctx.Request.PostForm {
		if strings.HasPrefix(key, "calc_values_") && len(values) > 0 {
			scopeID := strings.TrimPrefix(key, "calc_values_")
			calcValues[scopeID] = values[0]
		}
	}

	if err := h.Repository.FormStar(id, creatorID, calcValues); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Status changed and calculations updated"})
}

func (h *Handler) FinishStar(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	isMod, err := h.Repository.ModStatusCheck()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if !isMod {
		ctx.JSON(http.StatusConflict, fmt.Errorf("attempt to finish star as regular user").Error())
		return
	}

	star, err := h.Repository.GetStarByID(id)
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
		calcs, err := h.Repository.CalculateStar(idStr)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := h.Repository.SetStarStatus(idStr, status); err != nil {
			ctx.JSON(http.StatusConflict, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": status, "calculations": calcs})
	} else if status == "denied" {
		if err := h.Repository.SetStarStatus(idStr, status); err != nil {
			ctx.JSON(http.StatusConflict, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": status})
	} else {
		ctx.JSON(http.StatusConflict, errors.New("attempt to finish star with wrong status"))
	}
}

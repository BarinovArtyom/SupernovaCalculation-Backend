package handler

import (
	"fmt"
	"lab1/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetCurrentCalculations(ctx *gin.Context) {
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
		ctx.HTML(http.StatusOK, "error.tmpl", gin.H{
			"error_message": "Заявка не является активной",
		})
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

	ctx.HTML(http.StatusOK, "calculations.tmpl", gin.H{
		"current_star": current_star,
		"scopes":       scopesWithCalcs,
	})
}

func (h *Handler) DeleteStar(ctx *gin.Context) {
	starIDStr := ctx.PostForm("star_id") // исправлено на star_id

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

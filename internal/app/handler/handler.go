package handler

import (
	"net/http"
	"strconv"
	"strings"

	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetScopes(ctx *gin.Context) {
	query := ctx.Query("search")
	var MatchedScopes []repository.Scope

	allScopes := h.Repository.GetScopes()

	if query == "" {
		MatchedScopes = allScopes
	} else {
		for _, scope := range allScopes {
			if strings.Contains(strings.ToLower(scope.Name), strings.ToLower(query)) {
				MatchedScopes = append(MatchedScopes, scope)
			}
		}
	}

	star_id, err := h.Repository.StarIDByStatus()
	if err != nil {
		logrus.Error("Error getting star ID: ", err)
		star_id = 0
	}

	CalcCount := h.Repository.CountCalcsByStarID(star_id, h.Repository.GetCalcs())

	ctx.HTML(http.StatusOK, "scopes.tmpl", gin.H{
		"Title":     "Scopes",
		"Scopes":    MatchedScopes,
		"Star_ID":   star_id,
		"CalcCount": CalcCount,
	})
}

func (h *Handler) GetScopeDetails(ctx *gin.Context) {
	id := ctx.Param("id")
	serv_id, err := strconv.Atoi(id)
	if err != nil {
		ctx.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	allScopes := h.Repository.GetScopes()

	specs, err := h.Repository.ScopeByID(serv_id, allScopes)
	if err != nil {
		ctx.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	ctx.HTML(http.StatusOK, "specs.tmpl", gin.H{
		"Title": specs.Name,
		"Specs": specs,
	})
}

func (h *Handler) GetCurrentRequest(ctx *gin.Context) {
	id := ctx.Param("id")
	star_id, err := strconv.Atoi(id)
	if err != nil || star_id < 0 {
		ctx.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	current_star := h.Repository.StarByID(star_id, h.Repository.GetStars())

	allCalcs := h.Repository.GetCalcs()
	matching_scopes, err := h.Repository.ScopeByStar(star_id, allCalcs)
	if err != nil {
		ctx.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	matching_calcs := h.Repository.CalcByStar(star_id, h.Repository.GetCalcs())

	calcMap := make(map[int]repository.Calc)
	for _, calc := range matching_calcs {
		calcMap[calc.Scope_ID] = calc
	}

	ctx.HTML(http.StatusOK, "current_request.tmpl", gin.H{
		"current_star":  current_star,
		"current_scope": matching_scopes,
		"calcMap":       calcMap,
	})
}

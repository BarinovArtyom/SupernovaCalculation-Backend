package handler

import (
	"encoding/json"
	"fmt"
	"lab1/internal/app/config"
	"lab1/internal/app/ds"
	redis_api "lab1/internal/app/redis"
	"lab1/internal/app/role"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get current user profile information
// @Tags User
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/profile [get]
func (h *Handler) GetUserProfile(ctx *gin.Context) {
	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("error retrieving token payload: %s", err))
		return
	}
	userID := payload.UserID

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

// UpdateUser godoc
// @Summary Update user
// @Description Update current user profile (login and password)
// @Tags User
// @Param input body ds.Users true "User data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/edit [put]
func (h *Handler) UpdateUser(ctx *gin.Context) {
	payload, err := h.GetTokenPayload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("error retrieving token payload: %s", err))
		return
	}
	userID := payload.UserID

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "Invalid JSON format")
		return
	}

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

type loginReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	Guest    bool   `json:"guest" binding:"required"`
}

type loginResp struct {
	Login       string    `json:"login" binding:"required"`
	Role        role.Role `json:"role" binding:"required"`
	ExpiresIn   string    `json:"expires_in" binding:"required"`
	AccessToken string    `json:"access_token" binding:"required"`
	TokenType   string    `json:"token_type" binding:"required"`
}

func GenerateJWT(cfg *config.Config, userID int, role role.Role) (string, error) {
	token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(cfg.JWT.ExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "scope-admin",
		},
		UserID: userID,
		Role:   role,
	})
	tokenstr, err := token.SignedString([]byte(cfg.JWT.Token))
	if err != nil {
		return "", err
	}
	return tokenstr, nil
}

// LoginUser godoc
// @Summary      Login the specified user
// @Description  very very friendly response
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user body loginReq true "user data"
// @Success      200  {object} loginResp
// @Failure      403
// @Router       /user/login [post]
func (h *Handler) LoginUser(gCtx *gin.Context) {
	req := &loginReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if req.Guest {
		token, err := GenerateJWT(h.Config, 0, role.Role(role.Guest))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to generate a token"))
			return
		}
		if token == "" {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		gCtx.JSON(http.StatusOK, loginResp{
			Login:       "Гость",
			Role:        0,
			ExpiresIn:   time.Duration(h.Config.JWT.ExpiresIn).String(),
			AccessToken: token,
			TokenType:   "Bearer",
		})
		return

	}
	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if req.Login == user.Login && user.Password == h.Repository.GenerateHashString(req.Password) {
		token, err := GenerateJWT(h.Config, user.ID, role.Role(user.ModStatus))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to generate a token"))
			return
		}
		if token == "" {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		gCtx.JSON(http.StatusOK, loginResp{
			Login:       user.Login,
			Role:        user.ModStatus,
			ExpiresIn:   time.Duration(h.Config.JWT.ExpiresIn).String(),
			AccessToken: token,
			TokenType:   "Bearer",
		})
		return
	}

	gCtx.AbortWithStatus(http.StatusForbidden)
}

type registerReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerResp struct {
	Ok bool `json:"ok" binding:"required"`
}

// RegisterUser godoc
// @Summary User registration
// @Description Register new user account
// @Tags User
// @Param input body registerReq true "Registration data"
// @Success 200 {object} registerResp
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/register [post]
func (h *Handler) RegisterUser(gCtx *gin.Context) {
	req := &registerReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Password == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("pass is empty"))
		return
	}

	if req.Login == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("login is empty"))
		return
	}

	err = h.Repository.RegisterUser(&ds.Users{
		ModStatus: role.User,
		Login:     req.Login,
		Password:  h.Repository.GenerateHashString(req.Password), // пароли делаем в хешированном виде и далее будем сравнивать хеши, чтобы их не угнали с базой вместе
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}

// LogoutUser godoc
// @Summary User logout
// @Description Logout user and blacklist JWT token
// @Tags User
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/logout [post]
func (h *Handler) LogoutUser(gCtx *gin.Context) {
	// получаем заголовок
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
		gCtx.AbortWithStatus(http.StatusBadRequest) // отдаем что нет доступа

		return // завершаем обработку
	}

	// отрезаем префикс
	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)

		return
	}

	// сохраняем в блеклист редиса
	err = redis_api.WriteJWTToBlacklist(h.Repository.RedisClient, gCtx.Request.Context(), jwtStr, h.Config.JWT.ExpiresIn)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)

		return
	}

	gCtx.Status(http.StatusOK)
}

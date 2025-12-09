package handler

import (
	"errors"
	"fmt"
	"lab1/internal/app/ds"
	redis_api "lab1/internal/app/redis"
	"lab1/internal/app/role"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
)

const jwtPrefix = "Bearer "

// WithAuthCheck godoc
// @Summary      Check JWT
// @Tags         Middleware
// @Produce      json
// @Param        Authorization header string true "Auth Bearer token header"
// @Success      200
// @Failure      403
func (h *Handler) WithAuthCheck(assignedRoles ...role.Role) func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
			gCtx.AbortWithStatus(http.StatusForbidden) // отдаем что нет доступа

			return // завершаем обработку
		}

		// отрезаем префикс
		jwtStr = jwtStr[len(jwtPrefix):]

		// проверяем jwt в блеклист редиса
		err := redis_api.CheckJWTInBlacklist(h.Repository.RedisClient, gCtx.Request.Context(), jwtStr)
		if err == nil { // значит что токен в блеклисте
			gCtx.AbortWithStatus(http.StatusForbidden)

			return
		}
		if !errors.Is(err, redis.Nil) { // значит что это не ошибка отсуствия - внутренняя ошибка
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("redis internal: %s", err))
			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err != nil {
			gCtx.AbortWithStatus(http.StatusInternalServerError)
			log.Println(err)

			return
		}

		myClaims := token.Claims.(*ds.JWTClaims)

		for _, oneOfAssignedRole := range assignedRoles {
			if myClaims.Role == oneOfAssignedRole {
				return
			}
		}
		gCtx.AbortWithStatus(http.StatusForbidden)
		h.Logger.Printf("role %d is not assigned in %s", myClaims.Role, assignedRoles)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

type Payload struct {
	Role   role.Role
	UserID int
}

func (h *Handler) GetTokenPayload(gCtx *gin.Context) (*Payload, error) {
	jwtStr := gCtx.GetHeader("Authorization")
	if len(jwtStr) < len(jwtPrefix) {
		return nil, fmt.Errorf("no valid auth header provided")
	}
	jwtStr = jwtStr[len(jwtPrefix):]

	token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.Token), nil
	})
	if err != nil {
		return nil, err
	}

	myClaims := token.Claims.(*ds.JWTClaims)

	return &Payload{Role: myClaims.Role, UserID: myClaims.UserID}, nil
}

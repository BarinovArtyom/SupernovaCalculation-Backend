package ds

import (
	"lab1/internal/app/role"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	jwt.StandardClaims           // все что точно необходимо по RFC
	UserID             int       `json:"user_uuid"` // наши данные - uuid этого пользователя в базе данных
	Role               role.Role `json:"roles"`     // список доступов в нашей системе
}

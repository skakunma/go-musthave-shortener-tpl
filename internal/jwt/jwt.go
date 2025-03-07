package jwtauth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const (
	TokenEXP  = time.Hour * 3
	SecretKEY = "supersecretkey"
)

func BuildJWTString(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenEXP)),
		},
		UserID: userId,
	})

	tokenString, err := token.SignedString([]byte(SecretKEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

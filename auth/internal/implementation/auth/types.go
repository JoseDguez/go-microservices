package auth

import "github.com/golang-jwt/jwt/v5"

type User struct {
	userID   string
	password string
}

type MyClaims struct {
	jwt.RegisteredClaims
}

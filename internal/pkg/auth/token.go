package auth

import (
	"restaurant-ordering-system/internal/pkg/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role Role `json:"role"`
	jwt.RegisteredClaims
}

type Role string

const (
	AdminRole    Role = "admin"
	CustomerRole Role = "customer"
)

type CustomerJWTGenerator func(customerID model.CustomerID) (string, error)

func NewCustomerJWTGenerator(key []byte, ttl time.Duration) CustomerJWTGenerator {
	return func(customerID model.CustomerID) (string, error) {
		return GenerateCustomerJWT(customerID, key, ttl)
	}
}

func GenerateCustomerJWT(customerID model.CustomerID, key []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: CustomerRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   customerID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

type AdminJWTGenerator func() (string, error)

func NewAdminJWTGenerator(key []byte, ttl time.Duration) AdminJWTGenerator {
	return func() (string, error) {
		return GenerateAdminJWT(key, ttl)
	}
}

func GenerateAdminJWT(key []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: AdminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

type JWTParser func(tokenString string) (*Claims, error)

func NewJWTParser(key []byte) JWTParser {
	return func(tokenString string) (*Claims, error) {
		return ParseJWT(tokenString, key)
	}
}

func ParseJWT(tokenString string, key []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return key, nil
	})
	if token != nil {
		if claims, ok := token.Claims.(*Claims); ok {
			return claims, err
		}
	}
	return nil, err
}

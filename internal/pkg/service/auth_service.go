package service

import (
	"context"

	"restaurant-ordering-system/internal/pkg/auth"
	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db          *pgxpool.Pool
	queries     *repository.Queries
	generateJWT auth.CustomerJWTGenerator
}

func NewAuthService(db *pgxpool.Pool, generateJWT auth.CustomerJWTGenerator) *AuthService {
	return &AuthService{
		db:          db,
		queries:     repository.New(db),
		generateJWT: generateJWT,
	}
}

func (s *AuthService) GenerateToken(ctx context.Context, loginID model.LoginID, password string) (string, error) {
	c, err := s.queries.GetCustomerByLogin(ctx, string(loginID))
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(c.PasswordHash), []byte(password)); err != nil {
		return "", err
	}

	token, err := s.generateJWT(model.CustomerID(c.ID))
	if err != nil {
		return "", err
	}

	return token, nil
}

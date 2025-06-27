package service

// CustomerService provides methods for managing customers
import (
	"context"

	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func NewCustomer(repoCustomer repository.Customer) model.Customer {
	return model.Customer{
		ID:          model.CustomerID(repoCustomer.ID),
		Name:        repoCustomer.Name,
		Email:       repoCustomer.Email,
		PhoneNumber: repoCustomer.PhoneNumber.String,
		CreatedAt:   repoCustomer.CreatedAt.Time,
		UpdatedAt:   repoCustomer.UpdatedAt.Time,
	}
}

type CustomerService struct {
	db      *pgxpool.Pool
	queries *repository.Queries
}

func NewCustomerService(db *pgxpool.Pool) *CustomerService {
	return &CustomerService{
		db:      db,
		queries: repository.New(db),
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, params model.CreateCustomerParams) (model.Customer, error) {
	passwordHash, err := bcrypt.GenerateFromPassword(params.Password, bcrypt.DefaultCost)
	if err != nil {
		return model.Customer{}, err
	}
	c, err := s.queries.CreateCustomer(ctx, repository.CreateCustomerParams{
		LoginID:      params.LoginID.String(),
		Email:        params.Email,
		PasswordHash: string(passwordHash),
		Name:         params.Name,
		PhoneNumber:  pgtype.Text{String: params.PhoneNumber, Valid: params.PhoneNumber != ""},
	})
	if err != nil {
		return model.Customer{}, err
	}
	return NewCustomer(c), nil
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id model.CustomerID) (model.Customer, error) {
	c, err := s.queries.GetCustomerByID(ctx, uuid.UUID(id))
	if err != nil {
		return model.Customer{}, err
	}
	return NewCustomer(c), nil
}

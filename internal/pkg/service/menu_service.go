package service

// MenuService provides methods for managing menu items
import (
	"context"
	"time"

	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/repository"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewMenuItem(repoItem repository.MenuItem) *model.MenuItem {
	var deletedAt *time.Time
	if repoItem.DeletedAt.Valid {
		deletedAt = &repoItem.DeletedAt.Time
	}
	return &model.MenuItem{
		ID:              model.MenuItemID(repoItem.ID),
		Name:            repoItem.Name,
		Description:     repoItem.Description.String,
		PhotoPathinfo:   repoItem.PhotoPathinfo.String,
		Price:           repoItem.Price,
		PortionSize:     repoItem.PortionSize,
		Available:       repoItem.Available,
		ModifiersConfig: repoItem.ModifiersConfig,
		CreatedAt:       repoItem.CreatedAt.Time,
		DeletedAt:       deletedAt,
	}
}

type MenuService struct {
	db      *pgxpool.Pool
	queries *repository.Queries
}

func NewMenuService(db *pgxpool.Pool) *MenuService {
	return &MenuService{
		db:      db,
		queries: repository.New(db),
	}
}

func (s *MenuService) CreateMenuItem(ctx context.Context, params model.CreateMenuItemParams) (*model.MenuItem, error) {
	item, err := s.queries.CreateMenuItem(ctx, repository.CreateMenuItemParams{
		Name:            params.Name,
		Description:     pgtype.Text{String: params.Description, Valid: params.Description != ""},
		PhotoPathinfo:   pgtype.Text{String: params.PhotoPath, Valid: params.PhotoPath != ""},
		Price:           params.Price,
		PortionSize:     params.PortionSize,
		Available:       params.Available,
		ModifiersConfig: params.ModifiersConfig,
	})
	if err != nil {
		return nil, err
	}
	return NewMenuItem(item), nil
}

func (s *MenuService) GetMenuItem(ctx context.Context, id model.MenuItemID) (*model.MenuItem, error) {
	item, err := s.queries.GetMenuItem(ctx, int16(id))
	if err != nil {
		return nil, err
	}
	return NewMenuItem(item), nil
}

func (s *MenuService) ListMenuItems(ctx context.Context) ([]*model.MenuItem, error) {
	repoItems, err := s.queries.ListMenuItems(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*model.MenuItem, len(repoItems))
	for i, item := range repoItems {
		items[i] = NewMenuItem(item)
	}
	return items, nil
}

func (s *MenuService) UpdateMenuItem(ctx context.Context, id model.MenuItemID, params model.UpdateMenuItemParams) (*model.MenuItem, error) {
	item, err := s.queries.UpdateMenuItem(ctx, repository.UpdateMenuItemParams{
		ID:              int16(id),
		Name:            params.Name,
		Description:     pgtype.Text{String: params.Description, Valid: params.Description != ""},
		PhotoPathinfo:   pgtype.Text{String: params.PhotoPath, Valid: params.PhotoPath != ""},
		Price:           params.Price,
		PortionSize:     params.PortionSize,
		Available:       params.Available,
		ModifiersConfig: params.ModifiersConfig,
	})
	if err != nil {
		return nil, err
	}
	return NewMenuItem(item), nil
}

func (s *MenuService) DeleteMenuItem(ctx context.Context, id model.MenuItemID) error {
	return s.queries.SoftDeleteMenuItem(ctx, int16(id))
}

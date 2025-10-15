package service

// TabService provides methods for managing tabs
import (
	"context"
	"errors"
	"time"

	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/repository"
	"restaurant-ordering-system/internal/pkg/repository/cache"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func NewTab(repoTab repository.TabWithOrders) *model.Tab {
	var closedAt *time.Time
	if repoTab.ClosedAt.Valid {
		closedAt = &repoTab.ClosedAt.Time
	}

	orders := make([]*model.Order, len(repoTab.Orders))
	for i, order := range repoTab.Orders {
		orders[i] = NewOrder(order)
	}

	customGuestNames := make(map[model.GuestID]string, len(repoTab.GuestNames))
	for scopedID, name := range repoTab.GuestNames {
		guestID := model.GuestID{
			TabID:  model.TabID(repoTab.ID),
			Scoped: model.ScopedGuestID(scopedID),
		}
		customGuestNames[guestID] = name
	}

	return &model.Tab{
		ID:               model.TabID(repoTab.ID),
		TotalPrice:       repoTab.TotalPrice,
		Orders:           orders,
		CustomGuestNames: customGuestNames,
		CreatedAt:        repoTab.CreatedAt.Time,
		ClosedAt:         closedAt,
	}
}

type TabService struct {
	db           *pgxpool.Pool
	rdb          *redis.Client
	queries      *repository.Queries
	rqueries     *cache.RedisQueries
	cacheService *CacheService
}

func NewTabService(db *pgxpool.Pool, rdb *redis.Client, cacheService *CacheService) *TabService {
	return &TabService{
		db:           db,
		rdb:          rdb,
		queries:      repository.New(db),
		rqueries:     cache.New(rdb),
		cacheService: cacheService,
	}
}

func (s *TabService) CreateTab(ctx context.Context) (model.TabID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return model.TabID{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	row, err := qtx.CreateTab(ctx)
	if err != nil {
		return model.TabID{}, err
	}
	tabID := row.ID
	if err := qtx.CreateGuestIDSequence(ctx, tabID); err != nil {
		return model.TabID{}, err
	}
	if err := qtx.CreateOrderIDSequence(ctx, tabID); err != nil {
		return model.TabID{}, err
	}
	orderID, err := qtx.CreateOrder(ctx, tabID)
	if err != nil {
		return model.TabID{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.TabID{}, err
	}

	s.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		return cache.New(p).CreateTab(ctx, model.TabID(row.ID), row.CreatedAt.Time, model.ScopedOrderID(orderID))
	})

	return model.TabID(tabID), nil
}

func (s *TabService) VisitTab(ctx context.Context, tabID model.TabID, customerID model.CustomerID) error {
	return s.checkTabNotClosed(ctx, tabID, func(qtx *repository.Queries) error {
		return qtx.VisitTab(ctx, repository.VisitTabParams{
			TabID:      uuid.UUID(tabID),
			CustomerID: uuid.UUID(customerID),
		})
	})
}

func (s *TabService) CreateGuest(ctx context.Context, tabID model.TabID) (model.GuestID, error) {
	var scopedID model.ScopedGuestID
	if err := s.checkTabNotClosed(ctx, tabID, func(qtx *repository.Queries) error {
		scopedIDInt, err := qtx.CreateGuest(ctx, uuid.UUID(tabID))
		if err != nil {
			return err
		}

		scopedID = model.ScopedGuestID(scopedIDInt)

		return nil
	}); err != nil {
		return model.GuestID{}, err
	}

	return model.GuestID{
		TabID:  tabID,
		Scoped: scopedID,
	}, nil
}

func (s *TabService) UpdateGuestName(ctx context.Context, guestID model.GuestID, name string) error {
	if len(name) == 0 {
		return errors.New("name is empty")
	}

	if err := s.checkTabNotClosed(ctx, guestID.TabID, func(qtx *repository.Queries) error {
		return qtx.UpdateGuestName(ctx, repository.UpdateGuestNameParams{
			ID:       uuid.UUID(guestID.TabID),
			ScopedID: int16(guestID.Scoped),
			Name:     name,
		})
	}); err != nil {
		return err
	}

	if _, err := s.rdb.Pipelined(ctx, func(p redis.Pipeliner) error {
		cache.New(p).UpdateGuestName(ctx, guestID.TabID, guestID.Scoped, name)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *TabService) GetOpenTab(ctx context.Context, tabID model.TabID) (*model.Tab, error) {
	tab, err := s.rqueries.GetOpenTabWithOrders(ctx, tabID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if tab, err := s.cacheService.GetAndCacheTab(ctx, tabID); tab != nil {
				if tab.ClosedAt != nil {
					return nil, errors.New("tab is already closed")
				}
				return tab, nil
			} else {
				return nil, err
			}
		}
		return nil, err
	}

	return tab, nil
}

func (s *TabService) CloseTab(ctx context.Context, tabID model.TabID) (time.Time, error) {
	closedTabID := uuid.UUID(tabID)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	tab, err := qtx.GetTabForNoKeyUpdate(ctx, closedTabID)
	if err != nil {
		return time.Time{}, err
	}
	if tab.ClosedAt.Valid {
		return time.Time{}, errors.New("tab is already closed")
	}

	if err := qtx.DeleteNotSentOrders(ctx, closedTabID); err != nil {
		return time.Time{}, err
	}
	if err := qtx.DeleteOrderIDSequence(ctx, closedTabID); err != nil {
		return time.Time{}, err
	}
	if err := qtx.DeleteGuestIDSequence(ctx, closedTabID); err != nil {
		return time.Time{}, err
	}
	closedAtPg, err := qtx.CloseTab(ctx, closedTabID)
	if err != nil {
		return time.Time{}, err
	}
	tab.ClosedAt = closedAtPg
	closedAt := closedAtPg.Time

	if err := s.rdb.Watch(ctx, func(tx *redis.Tx) error {
		_, orderItemIDs, err := cache.WatchAndGetNotSentOrderIDAndItemIDs(ctx, tx, tabID)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil // Already invalidated
			}
			return err
		}
		if _, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			cache.New(p).InvalidateTab(ctx, tabID, orderItemIDs)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return time.Time{}, err
	}

	go s.cacheService.GetAndCacheTab(ctx, tabID)

	return closedAt, nil
}

func (s *TabService) checkTabNotClosed(ctx context.Context, tabID model.TabID, do func(qtx *repository.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	tab, err := qtx.GetTabForShare(ctx, uuid.UUID(tabID))
	if err != nil {
		return err
	}
	if tab.ClosedAt.Valid {
		return errors.New("tab is already closed")
	}

	if err := do(qtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *TabService) GetVisitedTabs(ctx context.Context, customerID model.CustomerID) ([]*model.Tab, error) {
	repoTabs, err := s.queries.GetVisitedTabsWithOrders(ctx, uuid.UUID(customerID))
	if err != nil {
		return nil, err
	}
	tabs := make([]*model.Tab, len(repoTabs))
	for i, repoTab := range repoTabs {
		tabs[i] = NewTab(repoTab)
	}
	return tabs, nil
}

package service

import (
	"context"
	"encoding/json"

	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/repository"
	"restaurant-ordering-system/internal/pkg/repository/cache"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"k8s.io/utils/keymutex"
)

func NewCacheService(db *pgxpool.Pool, rdb *redis.Client) *CacheService {
	return &CacheService{
		group:   new(singleflight.Group),
		mutex:   keymutex.NewHashed(int(db.Config().MaxConns)),
		queries: repository.New(db),
		rdb:     rdb,
	}
}

type CacheService struct {
	group   *singleflight.Group
	mutex   keymutex.KeyMutex
	queries *repository.Queries
	rdb     *redis.Client
}

func (s *CacheService) GetAndCacheTab(ctx context.Context, id model.TabID) (*model.Tab, error) {
	key := id.String()
	v, err, _ := s.group.Do(key, func() (any, error) {
		s.mutex.LockKey(key)
		defer s.mutex.UnlockKey(key)

		s.group.Forget(key)

		row, err := s.queries.GetTabWithOrdersForShare(ctx, uuid.UUID(id))
		if err != nil {
			return nil, err
		}

		repoTab := repository.TabWithOrders{
			ID:         row.ID,
			TotalPrice: row.TotalPrice,
			CreatedAt:  row.CreatedAt,
			ClosedAt:   row.ClosedAt,
			GuestNames: row.GuestNames,
		}
		if err := json.Unmarshal(row.Orders, &repoTab.Orders); err != nil {
			return nil, err
		}

		tab := NewTab(repoTab)

		if _, err := s.rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
			return cache.New(p).CacheTab(ctx, tab)
		}); err != nil {
			return tab, err
		}

		return tab, nil
	})
	tab, _ := v.(*model.Tab)
	return tab, err
}

package cache

import (
	"restaurant-ordering-system/internal/pkg/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestTab(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	rq := New(rdb)

	tabID := model.TabID(uuid.New())
	rq.CacheTab(t.Context(), &model.Tab{
		ID:        tabID,
		CreatedAt: time.Now(),
		Orders: []*model.Order{
			{
				ID: model.OrderID{
					TabID:  tabID,
					Scoped: 1,
				},
			},
		},
	})
}

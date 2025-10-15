package service

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

func NewOrder(repoOrder repository.OrderWithItems) *model.Order {
	var sentAt *time.Time
	if repoOrder.SentAt.Valid {
		sentAt = &repoOrder.SentAt.Time
	}

	items := make([]*model.OrderItem, len(repoOrder.Items))
	for i, item := range repoOrder.Items {
		items[i] = NewOrderItem(item)
	}

	return &model.Order{
		ID: model.OrderID{
			TabID:  model.TabID(repoOrder.TabID),
			Scoped: model.ScopedOrderID(repoOrder.ScopedID),
		},
		SentAt: sentAt,
		Items:  items,
	}
}

func NewOrderItem(repoItem repository.OrderItemWithMenu) *model.OrderItem {
	guestOwnerIDs := make([]model.GuestID, len(repoItem.GuestOwners))
	for i, id := range repoItem.GuestOwners {
		guestOwnerIDs[i] = model.GuestID{
			TabID:  model.TabID(repoItem.TabID),
			Scoped: model.ScopedGuestID(id),
		}
	}
	customerOwnerIDs := make([]model.CustomerID, len(repoItem.CustomerOwners))
	for i, id := range repoItem.CustomerOwners {
		customerOwnerIDs[i] = model.CustomerID(id)
	}
	return &model.OrderItem{
		ID: model.OrderItemID{
			OrderID: model.OrderID{
				TabID:  model.TabID(repoItem.TabID),
				Scoped: model.ScopedOrderID(repoItem.OrderID),
			},
			Scoped: model.ScopedOrderItemID(repoItem.ScopedID),
		},
		Quantity:         repoItem.Quantity,
		Modifiers:        repoItem.Modifiers,
		GuestOwnerIDs:    guestOwnerIDs,
		CustomerOwnerIDs: customerOwnerIDs,
		MenuItemID:       model.MenuItemID(repoItem.MenuItemID),
		Name:             repoItem.Name,
		Description:      repoItem.Description.String,
		PhotoPathinfo:    repoItem.PhotoPathinfo.String,
		Price:            repoItem.Price,
		PortionSize:      repoItem.PortionSize,
		ModifiersConfig:  repoItem.ModifiersConfig,
	}
}

type OrderService struct {
	db           *pgxpool.Pool
	rdb          *redis.Client
	queries      *repository.Queries
	rqueries     *cache.RedisQueries
	cacheService *CacheService
}

func NewOrderService(db *pgxpool.Pool, rdb *redis.Client, cacheService *CacheService) *OrderService {
	return &OrderService{
		db:           db,
		rdb:          rdb,
		queries:      repository.New(db),
		rqueries:     cache.New(rdb),
		cacheService: cacheService,
	}
}

func (s *OrderService) CreateOrderItem(ctx context.Context, params model.CreateOrderItemParams) (model.OrderItemID, error) {
	if params.Quantity < 1 {
		return model.OrderItemID{}, errors.New("quantity is < 1")
	}

	menuItem, err := s.queries.GetNotDeletedMenuItem(ctx, int16(params.MenuItemID))
	if err != nil {
		return model.OrderItemID{}, err
	}
	if !menuItem.Available {
		return model.OrderItemID{}, errors.New("menu item is not available")
	}

	visitingGuestIDs := make([]model.GuestID, 0, len(params.GuestOwnerIDs))
	for _, guestID := range params.GuestOwnerIDs {
		if guestID.TabID == params.OrderID.TabID {
			visitingGuestIDs = append(visitingGuestIDs, guestID)
		}
	}

	customerIDs := make([]uuid.UUID, len(params.CustomerOwnerIDs))
	for i, id := range params.CustomerOwnerIDs {
		customerIDs[i] = uuid.UUID(id)
	}
	visitingCustomerIDs, err := s.queries.IsVisitingCustomerIDs(ctx, repository.IsVisitingCustomerIDsParams{
		TabID:       uuid.UUID(params.OrderID.TabID),
		CustomerIds: customerIDs,
	})
	if err != nil {
		return model.OrderItemID{}, err
	}

	var orderItemID model.OrderItemID
	if err := s.checkOrderNotSent(ctx, params.OrderID, func(tx *redis.Tx) error {
		scopedID, err := cache.New(tx).GetNextOrderItemID(ctx, params.OrderID)
		if err != nil {
			return err
		}
		orderItemID = model.OrderItemID{
			OrderID: params.OrderID,
			Scoped:  scopedID,
		}
		customerOwnerIDs := make([]model.CustomerID, len(visitingCustomerIDs))
		for i, id := range visitingCustomerIDs {
			customerOwnerIDs[i] = model.CustomerID(id)
		}

		if _, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
			cache.New(p).CreateOrderItem(ctx, &model.OrderItem{
				ID:               orderItemID,
				Quantity:         params.Quantity,
				Modifiers:        params.Modifiers,
				GuestOwnerIDs:    visitingGuestIDs,
				CustomerOwnerIDs: customerOwnerIDs,
				MenuItemID:       params.MenuItemID,
				Name:             menuItem.Name,
				Description:      menuItem.Description.String,
				PhotoPathinfo:    menuItem.PhotoPathinfo.String,
				Price:            menuItem.Price,
				PortionSize:      menuItem.PortionSize,
				ModifiersConfig:  menuItem.ModifiersConfig,
			})
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return model.OrderItemID{}, err
	}

	return orderItemID, nil
}

func (s *OrderService) DeleteOrderItem(ctx context.Context, orderItemID model.OrderItemID) error {
	return s.checkOrderItemNotSent(ctx, orderItemID, func(q *cache.RedisQueries) {
		q.DeleteOrderItem(ctx, orderItemID)
	})
}

func (s *OrderService) UpdateOrderItemModifiers(ctx context.Context, orderItemID model.OrderItemID, modifiers []byte) error {
	return s.checkOrderItemNotSent(ctx, orderItemID, func(q *cache.RedisQueries) {
		q.UpdateOrderItemModifiers(ctx, orderItemID, modifiers)
	})
}

func (s *OrderService) UpdateOrderItemQuantity(ctx context.Context, orderItemID model.OrderItemID, quantity int16) error {
	return s.checkOrderItemNotSent(ctx, orderItemID, func(q *cache.RedisQueries) {
		q.UpdateOrderItemQuantity(ctx, orderItemID, quantity)
	})
}

func (s *OrderService) AddOrderItemGuestOwner(ctx context.Context, orderItemID model.OrderItemID, guestID model.GuestID) error {
	return s.checkOrderItemNotSent(ctx, orderItemID, func(q *cache.RedisQueries) {
		q.AddOrderItemGuestOwner(ctx, orderItemID, guestID)
	})
}

func (s *OrderService) RemoveOrderItemGuestOwner(ctx context.Context, orderItemID model.OrderItemID, guestID model.GuestID) error {
	return s.checkOrderItemNotSent(ctx, orderItemID, func(q *cache.RedisQueries) {
		q.RemoveOrderItemGuestOwner(ctx, orderItemID, guestID)
	})
}

func (s *OrderService) AddOrderItemCustomerOwner(ctx context.Context, orderItemID model.OrderItemID, customerID model.CustomerID) error {
	return s.checkOrderItemNotSent(ctx, orderItemID, func(q *cache.RedisQueries) {
		q.AddOrderItemCustomerOwner(ctx, orderItemID, customerID)
	})
}

func (s *OrderService) RemoveOrderItemCustomerOwner(ctx context.Context, orderItemID model.OrderItemID, customerID model.CustomerID) error {
	return s.checkOrderItemNotSent(ctx, orderItemID, func(q *cache.RedisQueries) {
		q.RemoveOrderItemCustomerOwner(ctx, orderItemID, customerID)
	})
}

func (s *OrderService) checkOrderItemNotSent(ctx context.Context, id model.OrderItemID, fn func(q *cache.RedisQueries)) error {
	return s.checkOrderNotSent(ctx, id.OrderID, func(tx *redis.Tx) error {
		_, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
			fn(cache.New(p))
			return nil
		})
		return err
	})
}

func (s *OrderService) checkOrderNotSent(ctx context.Context, id model.OrderID, fn func(tx *redis.Tx) error) error {
	return s.rdb.Watch(ctx, func(tx *redis.Tx) error {
		ok, err := cache.WatchAndCheckOrderNotSent(ctx, tx, id)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				return err
			}

			if err := tx.Unwatch(ctx).Err(); err != nil {
				return err
			}

			if _, err := s.cacheService.GetAndCacheTab(ctx, id.TabID); err != nil {
				return err
			}

			ok, err = cache.WatchAndCheckOrderNotSent(ctx, tx, id)
			if err != nil {
				return err
			}
		}
		if !ok {
			return errors.New("order is already sent")
		}

		return fn(tx)
	})
}

func (s *OrderService) SendOrder(ctx context.Context, toBeSentOrderID model.OrderID) error {
	tabID := uuid.UUID(toBeSentOrderID.TabID)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	tab, err := qtx.GetTabForShare(ctx, tabID)
	if err != nil {
		return err
	}
	if tab.ClosedAt.Valid {
		return errors.New("tab is already closed")
	}

	if err := s.rdb.Watch(ctx, func(tx *redis.Tx) error {
		var miss bool
		notSentOrderID, orderItemIDs, err := cache.WatchAndGetNotSentOrderIDAndItemIDs(ctx, tx, toBeSentOrderID.TabID)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				return err
			}
			miss = true
			order, err := qtx.GetOrderWithItems(ctx, repository.GetOrderWithItemsParams{
				TabID:    uuid.UUID(toBeSentOrderID.TabID),
				ScopedID: int16(toBeSentOrderID.Scoped),
			})
			if err != nil {
				return err
			}
			if order.SentAt.Valid {
				return errors.New("order is already sent")
			}
			if len(order.Items) == 0 {
				return errors.New("order is empty")
			}
		} else {
			if notSentOrderID != toBeSentOrderID {
				return errors.New("order is already sent")
			}
			if len(orderItemIDs) == 0 {
				return errors.New("order is empty")
			}
			items, err := cache.WatchAndGetOrderItems(ctx, tx, orderItemIDs)
			if err != nil {
				return err
			}
			for i, item := range items {
				orderItemIDs[i] = item.ID
			}
			if err := s.replaceOrderItems(ctx, toBeSentOrderID, items); err != nil {
				return err
			}
		}

		if err := qtx.SendOrder(ctx, repository.SendOrderParams{
			TabID:    tabID,
			ScopedID: int16(toBeSentOrderID.Scoped),
		}); err != nil {
			return err
		}
		if err := qtx.UpdateTabTotalPrice(ctx, tabID); err != nil {
			return err
		}
		if _, err := qtx.CreateOrder(ctx, tabID); err != nil {
			return err
		}

		if !miss {
			if _, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
				cache.New(p).InvalidateTab(ctx, toBeSentOrderID.TabID, orderItemIDs)
				return nil
			}); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	go s.cacheService.GetAndCacheTab(ctx, toBeSentOrderID.TabID)

	return nil
}

func (s *OrderService) replaceOrderItems(ctx context.Context, orderID model.OrderID, items []*model.OrderItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteOrderItems(ctx, repository.DeleteOrderItemsParams{
		TabID:   uuid.UUID(orderID.TabID),
		OrderID: int16(orderID.Scoped),
	}); err != nil {
		return err
	}

	params := make([]repository.CreateOrderItemsParams, len(items))
	for i, item := range items {
		guestOwners := make([]int16, len(item.GuestOwnerIDs))
		for i, id := range item.GuestOwnerIDs {
			guestOwners[i] = int16(id.Scoped)
		}
		customerOwners := make([]uuid.UUID, len(item.CustomerOwnerIDs))
		for i, id := range item.CustomerOwnerIDs {
			customerOwners[i] = uuid.UUID(id)
		}
		params[i] = repository.CreateOrderItemsParams{
			TabID:          uuid.UUID(item.ID.OrderID.TabID),
			OrderID:        int16(item.ID.OrderID.Scoped),
			ScopedID:       int16(item.ID.Scoped),
			MenuItemID:     int16(item.MenuItemID),
			Quantity:       item.Quantity,
			Modifiers:      item.Modifiers,
			GuestOwners:    guestOwners,
			CustomerOwners: customerOwners,
		}
	}

	if _, err := qtx.CreateOrderItems(ctx, params); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

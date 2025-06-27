package service

import (
	"context"
	"errors"
	"time"

	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db      *pgxpool.Pool
	queries *repository.Queries
}

func NewOrderService(db *pgxpool.Pool) *OrderService {
	return &OrderService{
		db:      db,
		queries: repository.New(db),
	}
}

func (s *OrderService) CreateOrderItem(ctx context.Context, params model.CreateOrderItemParams) (model.OrderItemID, error) {
	var scopedID model.ScopedOrderItemID
	if err := s.checkOrderNotSent(ctx, params.OrderID, func(qtx *repository.Queries) error {
		tabID := uuid.UUID(params.OrderID.TabID)

		visitingGuestIDs := make([]int16, 0)
		for _, guestID := range params.GuestOwnerIDs {
			if guestID.TabID == params.OrderID.TabID {
				visitingGuestIDs = append(visitingGuestIDs, int16(guestID.Scoped))
			}
		}

		customerIDs := make([]uuid.UUID, len(params.CustomerOwnerIDs))
		for i, id := range params.CustomerOwnerIDs {
			customerIDs[i] = uuid.UUID(id)
		}

		visitingCustomerIDs, err := qtx.IsVisitingCustomerIDs(ctx, repository.IsVisitingCustomerIDsParams{
			TabID:       tabID,
			CustomerIds: customerIDs,
		})
		if err != nil {
			return err
		}

		scopedIDInt, err := qtx.CreateOrderItem(ctx, repository.CreateOrderItemParams{
			TabID:          tabID,
			OrderID:        int16(params.OrderID.Scoped),
			MenuItemID:     int16(params.MenuItemID),
			Quantity:       params.Quantity,
			Modifiers:      params.Modifiers,
			GuestOwners:    visitingGuestIDs,
			CustomerOwners: visitingCustomerIDs,
		})
		if err != nil {
			return err
		}

		scopedID = model.ScopedOrderItemID(scopedIDInt)

		return nil
	}); err != nil {
		return model.OrderItemID{}, err
	}

	return model.OrderItemID{
		OrderID: params.OrderID,
		Scoped:  scopedID,
	}, nil
}

func (s *OrderService) DeleteOrderItem(ctx context.Context, orderItemID model.OrderItemID) error {
	return s.checkOrderNotSent(ctx, orderItemID.OrderID, func(qtx *repository.Queries) error {
		return qtx.DeleteOrderItem(ctx, repository.DeleteOrderItemParams{
			TabID:    uuid.UUID(orderItemID.OrderID.TabID),
			OrderID:  int16(orderItemID.OrderID.Scoped),
			ScopedID: int16(orderItemID.Scoped),
		})
	})
}

func (s *OrderService) UpdateOrderItemModifiers(ctx context.Context, orderItemID model.OrderItemID, modifiers []byte) error {
	return s.checkOrderNotSent(ctx, orderItemID.OrderID, func(qtx *repository.Queries) error {
		return qtx.UpdateOrderItemModifiers(ctx, repository.UpdateOrderItemModifiersParams{
			TabID:     uuid.UUID(orderItemID.OrderID.TabID),
			OrderID:   int16(orderItemID.OrderID.Scoped),
			ScopedID:  int16(orderItemID.Scoped),
			Modifiers: modifiers,
		})
	})
}

func (s *OrderService) UpdateOrderItemQuantity(ctx context.Context, orderItemID model.OrderItemID, quantity int16) error {
	return s.checkOrderNotSent(ctx, orderItemID.OrderID, func(qtx *repository.Queries) error {
		return qtx.UpdateOrderItemQuantity(ctx, repository.UpdateOrderItemQuantityParams{
			TabID:    uuid.UUID(orderItemID.OrderID.TabID),
			OrderID:  int16(orderItemID.OrderID.Scoped),
			ScopedID: int16(orderItemID.Scoped),
			Quantity: quantity,
		})
	})
}

func (s *OrderService) AddOrderItemGuestOwner(ctx context.Context, orderItemID model.OrderItemID, guestID model.GuestID) error {
	return s.checkOrderNotSent(ctx, orderItemID.OrderID, func(qtx *repository.Queries) error {
		return qtx.AddOrderItemGuestOwner(ctx, repository.AddOrderItemGuestOwnerParams{
			TabID:    uuid.UUID(orderItemID.OrderID.TabID),
			OrderID:  int16(orderItemID.OrderID.Scoped),
			ScopedID: int16(orderItemID.Scoped),
			GuestID:  int16(guestID.Scoped),
		})
	})
}

func (s *OrderService) RemoveOrderItemGuestOwner(ctx context.Context, orderItemID model.OrderItemID, guestID model.GuestID) error {
	return s.checkOrderNotSent(ctx, orderItemID.OrderID, func(qtx *repository.Queries) error {
		return qtx.RemoveOrderItemGuestOwner(ctx, repository.RemoveOrderItemGuestOwnerParams{
			TabID:    uuid.UUID(orderItemID.OrderID.TabID),
			OrderID:  int16(orderItemID.OrderID.Scoped),
			ScopedID: int16(orderItemID.Scoped),
			GuestID:  int16(guestID.Scoped),
		})
	})
}

func (s *OrderService) AddOrderItemCustomerOwner(ctx context.Context, orderItemID model.OrderItemID, customerID model.CustomerID) error {
	return s.checkOrderNotSent(ctx, orderItemID.OrderID, func(qtx *repository.Queries) error {
		return qtx.AddOrderItemCustomerOwner(ctx, repository.AddOrderItemCustomerOwnerParams{
			TabID:      uuid.UUID(orderItemID.OrderID.TabID),
			OrderID:    int16(orderItemID.OrderID.Scoped),
			ScopedID:   int16(orderItemID.Scoped),
			CustomerID: uuid.UUID(customerID),
		})
	})
}

func (s *OrderService) RemoveOrderItemCustomerOwner(ctx context.Context, orderItemID model.OrderItemID, customerID model.CustomerID) error {
	return s.checkOrderNotSent(ctx, orderItemID.OrderID, func(qtx *repository.Queries) error {
		return qtx.RemoveOrderItemCustomerOwner(ctx, repository.RemoveOrderItemCustomerOwnerParams{
			TabID:      uuid.UUID(orderItemID.OrderID.TabID),
			OrderID:    int16(orderItemID.OrderID.Scoped),
			ScopedID:   int16(orderItemID.Scoped),
			CustomerID: uuid.UUID(customerID),
		})
	})
}

func (s *OrderService) SendOrder(ctx context.Context, orderID model.OrderID) error {
	tabID := uuid.UUID(orderID.TabID)
	sentOrderID := int16(orderID.Scoped)

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
	order, err := qtx.GetOrderForNoKeyUpdate(ctx, repository.GetOrderForNoKeyUpdateParams{
		TabID:    tabID,
		ScopedID: sentOrderID,
	})
	if err != nil {
		return err
	}
	if order.SentAt.Valid {
		return errors.New("order is already sent")
	}

	if err := qtx.DeleteOrderItemIDSequence(ctx, repository.DeleteOrderItemIDSequenceParams{
		TabID:   tabID,
		OrderID: sentOrderID,
	}); err != nil {
		return err
	}
	if err := qtx.SendOrder(ctx, repository.SendOrderParams{
		TabID:    tabID,
		ScopedID: sentOrderID,
	}); err != nil {
		return err
	}
	if err := qtx.UpdateTabTotalPrice(ctx, tabID); err != nil {
		return err
	}
	nextOrderID, err := qtx.CreateOrder(ctx, tabID)
	if err != nil {
		return err
	}
	if err := qtx.CreateOrderItemIDSequence(ctx, repository.CreateOrderItemIDSequenceParams{
		TabID:   tabID,
		OrderID: nextOrderID,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *OrderService) checkOrderNotSent(ctx context.Context, orderID model.OrderID, do func(qtx *repository.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	order, err := qtx.GetOrderForShare(ctx, repository.GetOrderForShareParams{
		TabID:    uuid.UUID(orderID.TabID),
		ScopedID: int16(orderID.Scoped),
	})
	if err != nil {
		return err
	}
	if order.SentAt.Valid {
		return errors.New("order is already sent")
	}

	if err := do(qtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

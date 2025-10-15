package cache

import (
	"context"
	"errors"
	"math"
	"slices"

	"restaurant-ordering-system/internal/pkg/model"

	"github.com/redis/go-redis/v9"
)

func (q *RedisQueries) GetNextOrderItemID(ctx context.Context, id model.OrderID) (model.ScopedOrderItemID, error) {
	nextItemID, err := q.rdb.Incr(ctx, orderItemIDSequenceKey(id.TabID)).Result()
	if err != nil {
		return 0, err
	}
	if nextItemID > math.MaxInt16 {
		return 0, errors.New("out of range")
	}
	return model.ScopedOrderItemID(nextItemID), nil
}

func (q *RedisQueries) CreateOrderItem(ctx context.Context, item *model.OrderItem) {
	q.cacheOrderItem(ctx, item)
}

func (q *RedisQueries) UpdateOrderItemModifiers(ctx context.Context, id model.OrderItemID, modifiers []byte) {
	q.rdb.HSet(ctx, orderItemKey(id), "modifiers", string(modifiers))
}

func (q *RedisQueries) UpdateOrderItemQuantity(ctx context.Context, id model.OrderItemID, quantity int16) {
	q.rdb.HSet(ctx, orderItemKey(id), "quantity", quantity)
}

func (q *RedisQueries) AddOrderItemGuestOwner(ctx context.Context, orderItemID model.OrderItemID, guestID model.GuestID) {
	q.rdb.SAdd(ctx, orderItemGuestOwnersListKey(orderItemID), guestID.Scoped)
}

func (q *RedisQueries) RemoveOrderItemGuestOwner(ctx context.Context, orderItemID model.OrderItemID, guestID model.GuestID) {
	q.rdb.SRem(ctx, orderItemGuestOwnersListKey(orderItemID), guestID.Scoped)
}

func (q *RedisQueries) AddOrderItemCustomerOwner(ctx context.Context, orderItemID model.OrderItemID, customerID model.CustomerID) {
	q.rdb.SAdd(ctx, orderItemCustomerOwnersListKey(orderItemID), customerID)
}

func (q *RedisQueries) RemoveOrderItemCustomerOwner(ctx context.Context, orderItemID model.OrderItemID, customerID model.CustomerID) {
	q.rdb.SRem(ctx, orderItemCustomerOwnersListKey(orderItemID), customerID)
}

func (q *RedisQueries) DeleteOrderItem(ctx context.Context, id model.OrderItemID) {
	q.rdb.ZRem(ctx, orderItemsListKey(id.OrderID.TabID), id.Scoped)
	q.rdb.Del(ctx, orderItemKey(id))
	q.rdb.Del(ctx, orderItemGuestOwnersListKey(id))
	q.rdb.Del(ctx, orderItemCustomerOwnersListKey(id))
}

func WatchAndCheckOrderNotSent(ctx context.Context, tx *redis.Tx, orderID model.OrderID) (bool, error) {
	if err := tx.Watch(ctx, tabNotSentOrderIDKey(orderID.TabID)).Err(); err != nil {
		return false, err
	}
	notSentOrderIDStr, err := tx.Get(ctx, tabNotSentOrderIDKey(orderID.TabID)).Result()
	if err != nil {
		return false, err
	}
	notSentOrderID, err := parseInt16(notSentOrderIDStr)
	if err != nil {
		return false, err
	}
	return model.ScopedOrderID(notSentOrderID) == orderID.Scoped, nil
}

func WatchAndGetNotSentOrderIDAndItemIDs(ctx context.Context, tx *redis.Tx, tabID model.TabID) (model.OrderID, []model.OrderItemID, error) {
	tx.Watch(ctx, tabNotSentOrderIDKey(tabID), orderItemsListKey(tabID))

	cmds, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.Get(ctx, tabNotSentOrderIDKey(tabID))
		p.ZRange(ctx, orderItemsListKey(tabID), 0, -1)
		return nil
	})
	if err != nil {
		return model.OrderID{}, nil, err
	}

	orderIDStr, err := cmds[0].(*redis.StringCmd).Result()
	if err != nil {
		return model.OrderID{}, nil, err
	}
	orderIDInt, err := parseInt16(orderIDStr)
	if err != nil {
		return model.OrderID{}, nil, err
	}
	orderID := model.OrderID{
		TabID:  tabID,
		Scoped: model.ScopedOrderID(orderIDInt),
	}

	orderItemIDsStr, err := cmds[1].(*redis.StringSliceCmd).Result()
	if err != nil {
		return model.OrderID{}, nil, err
	}
	orderItemIDs := make([]model.OrderItemID, len(orderItemIDsStr))
	for i, orderItemIDStr := range orderItemIDsStr {
		orderItemIDInt, err := parseInt16(orderItemIDStr)
		if err != nil {
			return model.OrderID{}, nil, err
		}
		orderItemIDs[i] = model.OrderItemID{
			OrderID: orderID,
			Scoped:  model.ScopedOrderItemID(orderItemIDInt),
		}
	}
	return orderID, orderItemIDs, nil
}

func WatchAndGetOrderItems(ctx context.Context, tx *redis.Tx, orderItemIDs []model.OrderItemID) ([]*model.OrderItem, error) {
	if len(orderItemIDs) == 0 {
		return nil, nil
	}

	p := tx.Pipeline()
	orderItemKeys := make([]string, 3*len(orderItemIDs))
	for i, orderItemID := range orderItemIDs {
		j := 3 * i

		orderItemKeys[j+0] = orderItemKey(orderItemID)
		p.HGetAll(ctx, orderItemKey(orderItemID))

		orderItemKeys[j+1] = orderItemGuestOwnersListKey(orderItemID)
		p.SMembers(ctx, orderItemGuestOwnersListKey(orderItemID))

		orderItemKeys[j+2] = orderItemCustomerOwnersListKey(orderItemID)
		p.SMembers(ctx, orderItemCustomerOwnersListKey(orderItemID))
	}
	if err := tx.Watch(ctx, orderItemKeys...).Err(); err != nil {
		return nil, err
	}

	cmds, err := p.Exec(ctx)
	if err != nil {
		return nil, err
	}

	orderItems := make([]*model.OrderItem, len(orderItemIDs))
	var i int
	for cmds := range slices.Chunk(cmds, 3) {
		if orderItems[i], err = OrderItemFromCmds(orderItemIDs[i].OrderID, cmds[0], cmds[1], cmds[2]); err != nil {
			return nil, err
		}
	}

	return orderItems, nil
}

func OrderItemFromCmds(orderID model.OrderID, orderItemCmd, guestOwnersCmd, customerOwnersCmd redis.Cmder) (*model.OrderItem, error) {
	oi := new(model.OrderItem)
	oi.ID.OrderID = orderID
	m, err := orderItemCmd.(*redis.MapStringStringCmd).Result()
	if err != nil {
		return nil, err
	}
	scopedIDInt, err := parseInt16(m["scoped_id"])
	if err != nil {
		return nil, err
	}
	oi.ID.Scoped = model.ScopedOrderItemID(scopedIDInt)
	menuItemIDInt, err := parseInt16(m["menu_item_id"])
	if err != nil {
		return nil, err
	}
	oi.MenuItemID = model.MenuItemID(menuItemIDInt)
	if oi.Quantity, err = parseInt16(m["quantity"]); err != nil {
		return nil, err
	}
	if len(m["modifiers"]) > 0 {
		oi.Modifiers = []byte(m["modifiers"])
	}
	guestOwnerIDsStr, err := guestOwnersCmd.(*redis.StringSliceCmd).Result()
	if err != nil {
		return nil, err
	}
	oi.GuestOwnerIDs = make([]model.GuestID, len(guestOwnerIDsStr))
	for i, guestOwnerIDStr := range guestOwnerIDsStr {
		guestOwnerIDInt, err := parseInt16(guestOwnerIDStr)
		if err != nil {
			return nil, err
		}
		oi.GuestOwnerIDs[i] = model.GuestID{
			TabID:  oi.ID.OrderID.TabID,
			Scoped: model.ScopedGuestID(guestOwnerIDInt),
		}
	}
	customerOwnerIDsStr, err := customerOwnersCmd.(*redis.StringSliceCmd).Result()
	if err != nil {
		return nil, err
	}
	oi.CustomerOwnerIDs = make([]model.CustomerID, len(customerOwnerIDsStr))
	for i, customerOwnerIDStr := range customerOwnerIDsStr {
		if oi.CustomerOwnerIDs[i], err = model.ParseCustomerID(customerOwnerIDStr); err != nil {
			return nil, err
		}
	}
	return oi, nil
}

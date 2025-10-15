package cache

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"restaurant-ordering-system/internal/pkg/model"

	"github.com/redis/go-redis/v9"
)

func (q *RedisQueries) CreateTab(ctx context.Context, tabID model.TabID, createdAt time.Time, orderID model.ScopedOrderID) error {
	return q.CacheTab(ctx, &model.Tab{
		ID:        tabID,
		CreatedAt: createdAt,
		Orders: []*model.Order{
			{
				ID: model.OrderID{
					TabID:  tabID,
					Scoped: orderID,
				},
			},
		},
	})
}

func (q *RedisQueries) CacheTab(ctx context.Context, tab *model.Tab) error {
	tabValues := []any{
		"id", tab.ID,
		"total_price", tab.TotalPrice,
		"created_at", tab.CreatedAt.Format(time.RFC3339Nano),
	}
	if tab.ClosedAt != nil {
		tabValues = append(tabValues, "closed_at", tab.ClosedAt.Format(time.RFC3339Nano))
	}
	q.rdb.HSet(ctx, tabKey(tab.ID), tabValues...)
	if tab.ClosedAt != nil {
		q.rdb.Expire(ctx, tabKey(tab.ID), tabCacheTTL)
	}

	if len(tab.CustomGuestNames) > 0 {
		for guestID, name := range tab.CustomGuestNames {
			q.UpdateGuestName(ctx, tab.ID, guestID.Scoped, name)
		}
		if tab.ClosedAt != nil {
			q.rdb.Expire(ctx, tabGuestNamesKey(tab.ID), tabCacheTTL)
		}
	}

	var sentOrder []*model.Order
	if lastOrder := tab.Orders[len(tab.Orders)-1]; lastOrder.SentAt != nil {
		sentOrder = tab.Orders
	} else {
		sentOrder = tab.Orders[:len(tab.Orders)-1]
		q.cacheNotSentOrder(ctx, lastOrder)
	}

	if err := q.cacheSentOrders(ctx, sentOrder, tab.ClosedAt != nil); err != nil {
		return err
	}

	return nil
}

func (q *RedisQueries) cacheNotSentOrder(ctx context.Context, order *model.Order) {
	q.rdb.Del(ctx, orderItemsListKey(order.ID.TabID))

	var maxID model.ScopedOrderItemID
	for _, item := range order.Items {
		maxID = max(maxID, model.ScopedOrderItemID(item.ID.Scoped))
		q.cacheOrderItem(ctx, item)
	}
	q.rdb.Set(ctx, orderItemIDSequenceKey(order.ID.TabID), maxID, 0)
	q.rdb.Set(ctx, tabNotSentOrderIDKey(order.ID.TabID), order.ID.Scoped, 0)
}

func (q *RedisQueries) cacheOrderItem(ctx context.Context, item *model.OrderItem) {
	q.rdb.ZAdd(ctx, orderItemsListKey(item.ID.OrderID.TabID), redis.Z{
		Score:  float64(item.ID.Scoped),
		Member: item.ID.Scoped,
	})
	q.rdb.HSet(ctx, orderItemKey(item.ID),
		"scoped_id", item.ID.Scoped,
		"menu_item_id", item.MenuItemID,
		"name", item.Name,
		"description", item.Description,
		"photo_pathinfo", item.PhotoPathinfo,
		"price", item.Price,
		"portion_size", item.PortionSize,
		"modifiers_config", string(item.ModifiersConfig),
	)
	q.UpdateOrderItemModifiers(ctx, item.ID, item.Modifiers)
	q.UpdateOrderItemQuantity(ctx, item.ID, item.Quantity)
	q.rdb.Del(ctx, orderItemGuestOwnersListKey(item.ID))
	for _, guestID := range item.GuestOwnerIDs {
		q.AddOrderItemGuestOwner(ctx, item.ID, guestID)
	}
	q.rdb.Del(ctx, orderItemCustomerOwnersListKey(item.ID))
	for _, customerID := range item.CustomerOwnerIDs {
		q.AddOrderItemCustomerOwner(ctx, item.ID, customerID)
	}
}

func (q *RedisQueries) cacheSentOrders(ctx context.Context, orders []*model.Order, tabClosed bool) error {
	if len(orders) > 0 {
		tabID := orders[0].ID.TabID

		sentOrdersJSON := make([]any, len(orders))
		for i, order := range orders {
			orderJSON, err := json.Marshal(order)
			if err != nil {
				return err
			}
			sentOrdersJSON[i] = string(orderJSON)
		}
		q.rdb.Del(ctx, ordersListKey(tabID))
		q.rdb.RPush(ctx, ordersListKey(tabID), sentOrdersJSON...)
		if tabClosed {
			q.rdb.Expire(ctx, ordersListKey(tabID), tabCacheTTL)
		}
	}

	return nil
}

func (q *RedisQueries) InvalidateTab(ctx context.Context, tabID model.TabID, notSentOrderItemIDs []model.OrderItemID) {
	q.rdb.Del(ctx, tabKey(tabID))
	q.rdb.Del(ctx, tabGuestNamesKey(tabID))
	q.rdb.Del(ctx, ordersListKey(tabID))
	q.rdb.Del(ctx, tabNotSentOrderIDKey(tabID))
	q.rdb.Del(ctx, orderItemIDSequenceKey(tabID))
	q.rdb.Del(ctx, orderItemsListKey(tabID))
	if len(notSentOrderItemIDs) > 0 {
		orderItemKeys := make([]string, len(notSentOrderItemIDs))
		orderItemGuestOwnersListKeys := make([]string, len(notSentOrderItemIDs))
		orderItemCustomerOwnersListKeys := make([]string, len(notSentOrderItemIDs))
		for i, id := range notSentOrderItemIDs {
			orderItemKeys[i] = orderItemKey(id)
			orderItemGuestOwnersListKeys[i] = orderItemGuestOwnersListKey(id)
			orderItemCustomerOwnersListKeys[i] = orderItemCustomerOwnersListKey(id)
		}
		q.rdb.Del(ctx, orderItemKeys...)
		q.rdb.Del(ctx, orderItemGuestOwnersListKeys...)
		q.rdb.Del(ctx, orderItemCustomerOwnersListKeys...)
	}
}

func (q *RedisQueries) GetOpenTabWithOrders(ctx context.Context, id model.TabID) (*model.Tab, error) {
	tab, err := q.getTabWithOrders(ctx, id)
	if err != nil {
		return nil, err
	}
	if tab.ClosedAt != nil {
		return nil, errors.New("tab is already closed")
	}
	return tab, nil
}

func (q *RedisQueries) getTabWithOrders(ctx context.Context, id model.TabID) (*model.Tab, error) {
	tabs, err := q.getTabsWithOrders(ctx, []model.TabID{id})
	if err != nil {
		return nil, err
	}
	return tabs[0], nil
}

func (q *RedisQueries) getTabsWithOrders(ctx context.Context, ids []model.TabID) ([]*model.Tab, error) {
	cmds, err := q.rdb.Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, id := range ids {
			p.Exists(ctx, tabKey(id))
			p.HGetAll(ctx, tabKey(id))
			p.HGetAll(ctx, tabGuestNamesKey(id))
			p.LRange(ctx, ordersListKey(id), 0, -1)
			p.Get(ctx, tabNotSentOrderIDKey(id))
			p.ZRange(ctx, orderItemsListKey(id), 0, -1)
		}
		return nil
	})
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	tabs := make([]*model.Tab, len(ids))
	orderItemIDsListStr := make([][]string, len(ids))
	for i := range len(ids) {
		j := i * 6

		tabExists, err := cmds[j].(*redis.IntCmd).Result()
		if err != nil {
			return nil, err
		}
		if tabExists == 0 {
			return nil, redis.Nil
		}

		tabRedis, err := cmds[j+1].(*redis.MapStringStringCmd).Result()
		if err != nil {
			return nil, err
		}
		guestNamesRedis, err := cmds[j+2].(*redis.MapStringStringCmd).Result()
		if err != nil {
			return nil, err
		}
		ordersRedis, err := cmds[j+3].(*redis.StringSliceCmd).Result()
		if err != nil {
			return nil, err
		}
		notSentOrderIDRedis, err := cmds[j+4].(*redis.StringCmd).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		orderItemsRedis, err := cmds[j+5].(*redis.StringSliceCmd).Result()
		if err != nil {
			return nil, err
		}

		tab := new(model.Tab)
		if tab.ID, err = model.ParseTabID(tabRedis["id"]); err != nil {
			return nil, err
		}
		if tab.TotalPrice, err = parseInt32(tabRedis["total_price"]); err != nil {
			return nil, err
		}
		if tab.CreatedAt, err = time.Parse(time.RFC3339Nano, tabRedis["created_at"]); err != nil {
			return nil, err
		}
		if closedAtStr := tabRedis["closed_at"]; closedAtStr != "" {
			closedAt, err := time.Parse(time.RFC3339Nano, closedAtStr)
			if err != nil {
				return nil, err
			}
			tab.ClosedAt = &closedAt
		}
		tab.CustomGuestNames = make(map[model.GuestID]string)
		for guestIDStr, guestName := range guestNamesRedis {
			scopedID, err := parseInt16(guestIDStr)
			if err != nil {
				return nil, err
			}
			guestID := model.GuestID{
				TabID:  tab.ID,
				Scoped: model.ScopedGuestID(scopedID),
			}
			tab.CustomGuestNames[guestID] = guestName
		}
		tab.Orders = make([]*model.Order, len(ordersRedis))
		for i, orderRedis := range ordersRedis {
			if err := json.Unmarshal([]byte(orderRedis), &tab.Orders[i]); err != nil {
				return nil, err
			}
		}
		if notSentOrderIDRedis != "" {
			notSentOrderIDInt, err := parseInt16(notSentOrderIDRedis)
			if err != nil {
				return nil, err
			}
			tab.Orders = append(tab.Orders, &model.Order{
				ID: model.OrderID{
					TabID:  tab.ID,
					Scoped: model.ScopedOrderID(notSentOrderIDInt),
				},
			})
		}

		tabs[i] = tab
		orderItemIDsListStr[i] = orderItemsRedis
	}

	cmds, err = q.rdb.Pipelined(ctx, func(p redis.Pipeliner) error {
		for i, orderItemIDsStr := range orderItemIDsListStr {
			for _, orderItemIDStr := range orderItemIDsStr {
				orders := tabs[i].Orders
				lastOrder := orders[len(orders)-1]
				scopedID, err := parseInt16(orderItemIDStr)
				if err != nil {
					return err
				}
				orderItemID := model.OrderItemID{
					OrderID: lastOrder.ID,
					Scoped:  model.ScopedOrderItemID(scopedID),
				}
				p.HGetAll(ctx, orderItemKey(orderItemID))
				p.SMembers(ctx, orderItemGuestOwnersListKey(orderItemID))
				p.SMembers(ctx, orderItemCustomerOwnersListKey(orderItemID))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var j int
	for i, orderItemIDsStr := range orderItemIDsListStr {
		if len(orderItemIDsStr) == 0 {
			continue
		}
		orders := tabs[i].Orders
		lastOrder := orders[len(orders)-1]
		lastOrder.Items = make([]*model.OrderItem, len(orderItemIDsStr))
		if lastOrder.Items[i], err = OrderItemFromCmds(lastOrder.ID, cmds[j+0], cmds[j+1], cmds[j+2]); err != nil {
			return nil, err
		}
		j += 3
	}

	return tabs, nil
}

func (q *RedisQueries) UpdateGuestName(ctx context.Context, tabID model.TabID, scopedGuestID model.ScopedGuestID, name string) {
	q.rdb.HSet(ctx, tabGuestNamesKey(tabID), strconv.Itoa(int(scopedGuestID)), name)
}

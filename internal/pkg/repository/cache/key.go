package cache

import (
	"errors"
	"fmt"
	"strconv"

	"restaurant-ordering-system/internal/pkg/model"
)

func tabKey(id model.TabID) string {
	return fmt.Sprintf("tab:%s", id)
}

func tabGuestNamesKey(id model.TabID) string {
	return fmt.Sprintf("tab:%s:guest_names", id)
}

func ordersListKey(id model.TabID) string {
	return fmt.Sprintf("tab:%s:orders", id)
}

func tabNotSentOrderIDKey(id model.TabID) string {
	return fmt.Sprintf("tab:%s:not_sent_order:id", id)
}

func orderItemIDSequenceKey(id model.TabID) string {
	return fmt.Sprintf("tab:%s:not_sent_order:order_item_id_sequence", id)
}

func orderItemsListKey(id model.TabID) string {
	return fmt.Sprintf("tab:%s:not_sent_order:order_items", id)
}

func orderItemKey(id model.OrderItemID) string {
	return fmt.Sprintf("tab:%s:order:%d:order_item:%d", id.OrderID.TabID, id.OrderID.Scoped, id.Scoped)
}

func orderItemGuestOwnersListKey(id model.OrderItemID) string {
	return fmt.Sprintf("tab:%s:order:%d:order_item:%d:guest_owners", id.OrderID.TabID, id.OrderID.Scoped, id.Scoped)
}

func orderItemCustomerOwnersListKey(id model.OrderItemID) string {
	return fmt.Sprintf("tab:%s:order:%d:order_item:%d:customer_owners", id.OrderID.TabID, id.OrderID.Scoped, id.Scoped)
}

func parseInt16(s string) (int16, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	i, err := strconv.ParseInt(s, 10, 16)
	return int16(i), err
}

func parseInt32(s string) (int32, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	i, err := strconv.ParseInt(s, 10, 32)
	return int32(i), err
}

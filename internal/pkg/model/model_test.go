package model

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTabSerialization(t *testing.T) {
	tabID := TabID(uuid.MustParse("11111111-2222-3333-4444-555555555555"))
	orderID := OrderID{TabID: tabID, Scoped: 1}
	orderItemID := OrderItemID{OrderID: orderID, Scoped: 1}
	customerID := CustomerID(uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
	guestID := GuestID{TabID: tabID, Scoped: 1}
	tabCreatedAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	orderSentAt := time.Date(2024, 6, 1, 12, 5, 0, 0, time.UTC)
	tabClosedAt := time.Date(2024, 6, 3, 14, 0, 0, 0, time.UTC)
	quantity := int16(2)
	price := int32(1200)
	totalPrice := int32(price * int32(quantity))

	tab := Tab{
		ID:         tabID,
		TotalPrice: totalPrice,
		Orders: []*Order{
			{
				ID: orderID,
				Items: []*OrderItem{
					{
						ID:               orderItemID,
						Quantity:         quantity,
						Modifiers:        []byte(`{"spicy":false}`),
						GuestOwnerIDs:    []GuestID{guestID},
						CustomerOwnerIDs: []CustomerID{customerID},
						MenuItemID:       10,
						Name:             "Pizza",
						Description:      "Cheese pizza",
						PhotoPathinfo:    "/img/pizza.jpg",
						Price:            price,
						PortionSize:      1,
						ModifiersConfig:  []byte(`{"extra_cheese":true}`),
					},
				},
				SentAt: &orderSentAt,
			},
		},
		CustomGuestNames: map[GuestID]string{
			guestID: "Alice",
		},
		CreatedAt: tabCreatedAt,
		ClosedAt:  &tabClosedAt,
	}

	b, err := json.MarshalIndent(tab, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal tab: %v", err)
	}
	fmt.Println(string(b))
}

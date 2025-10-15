// Package model defines the core domain models for the restaurant ordering system
package model

import (
	"strconv"

	"github.com/google/uuid"
)

type TabID uuid.UUID

func (id TabID) String() string {
	return uuid.UUID(id).String()
}

func (id TabID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id TabID) MarshalBinary() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseTabID(s string) (TabID, error) {
	u, err := uuid.Parse(s)
	return TabID(u), err
}

type GuestID struct {
	TabID  TabID
	Scoped ScopedGuestID
}

func (id GuestID) String() string {
	return id.TabID.String() + "." + strconv.FormatInt(int64(id.Scoped), 10)
}

func (id GuestID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

type ScopedGuestID int16

func (id ScopedGuestID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id ScopedGuestID) MarshalBinary() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseGuestID(s string) (GuestID, error) {
	parts := splitID(s)
	if len(parts) != 2 {
		return GuestID{}, strconv.ErrSyntax
	}
	tabID, err := ParseTabID(parts[0])
	if err != nil {
		return GuestID{}, err
	}
	scoped, err := strconv.ParseInt(parts[1], 10, 16)
	if err != nil {
		return GuestID{}, err
	}
	if scoped <= 0 {
		return GuestID{}, strconv.ErrSyntax
	}
	return GuestID{TabID: tabID, Scoped: ScopedGuestID(scoped)}, nil
}

type OrderID struct {
	TabID  TabID
	Scoped ScopedOrderID
}

func (id OrderID) String() string {
	return id.TabID.String() + "." + strconv.FormatInt(int64(id.Scoped), 10)
}

func (id OrderID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

type ScopedOrderID int16

func (id ScopedOrderID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id ScopedOrderID) MarshalBinary() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseOrderID(s string) (OrderID, error) {
	parts := splitID(s)
	if len(parts) != 2 {
		return OrderID{}, strconv.ErrSyntax
	}
	tabID, err := ParseTabID(parts[0])
	if err != nil {
		return OrderID{}, err
	}
	scoped, err := strconv.ParseInt(parts[1], 10, 16)
	if err != nil {
		return OrderID{}, err
	}
	if scoped <= 0 {
		return OrderID{}, strconv.ErrSyntax
	}
	return OrderID{TabID: tabID, Scoped: ScopedOrderID(scoped)}, nil
}

type OrderItemID struct {
	OrderID OrderID
	Scoped  ScopedOrderItemID
}

func (id OrderItemID) String() string {
	return id.OrderID.String() + "." + id.Scoped.String()
}

func (id OrderItemID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

type ScopedOrderItemID int16

func (id ScopedOrderItemID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id ScopedOrderItemID) MarshalBinary() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseOrderItemID(s string) (OrderItemID, error) {
	parts := splitID(s)
	if len(parts) != 3 {
		return OrderItemID{}, strconv.ErrSyntax
	}
	orderID, err := ParseOrderID(parts[0] + "." + parts[1])
	if err != nil {
		return OrderItemID{}, err
	}
	scoped, err := strconv.ParseInt(parts[2], 10, 16)
	if err != nil {
		return OrderItemID{}, err
	}
	if scoped <= 0 {
		return OrderItemID{}, strconv.ErrSyntax
	}
	return OrderItemID{OrderID: orderID, Scoped: ScopedOrderItemID(scoped)}, nil
}

type CustomerID uuid.UUID

func (id CustomerID) String() string {
	return uuid.UUID(id).String()
}

func (id CustomerID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *CustomerID) UnmarshalText(b []byte) error {
	parsed, err := uuid.ParseBytes(b)
	if err != nil {
		return err
	}
	*id = CustomerID(parsed)
	return nil
}

func (id CustomerID) MarshalBinary() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseCustomerID(s string) (CustomerID, error) {
	u, err := uuid.Parse(s)
	return CustomerID(u), err
}

type MenuItemID int16

func (id MenuItemID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id MenuItemID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id MenuItemID) MarshalBinary() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseMenuItemID(s string) (MenuItemID, error) {
	val, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, strconv.ErrSyntax
	}
	return MenuItemID(val), nil
}

type MenuTagID int16

func (id MenuTagID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id MenuTagID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseMenuTagID(s string) (MenuTagID, error) {
	val, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, strconv.ErrSyntax
	}
	return MenuTagID(val), nil
}

type MenuTagDimensionID int16

func (id MenuTagDimensionID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id MenuTagDimensionID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func ParseMenuTagDimensionID(s string) (MenuTagDimensionID, error) {
	val, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, strconv.ErrSyntax
	}
	return MenuTagDimensionID(val), nil
}

type LoginID string

func (id LoginID) String() string {
	return string(id)
}

func (id LoginID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *LoginID) UnmarshalText(b []byte) error {
	*id = LoginID(string(b))
	return nil
}

func ParseLoginID(s string) (LoginID, error) {
	return LoginID(s), nil
}

// splitID splits a dot-delimited ID string into its parts.
func splitID(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

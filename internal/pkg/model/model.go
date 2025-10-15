// Package model defines the core domain models for the restaurant ordering system
package model

import (
	"encoding/json"
	"time"
)

// Tab represents a dining session that tracks customer orders
type Tab struct {
	ID               TabID              `json:"id"`
	TotalPrice       int32              `json:"total_price"`
	Orders           []*Order           `json:"orders"`
	CustomGuestNames map[GuestID]string `json:"custom_guest_names"`
	CreatedAt        time.Time          `json:"created_at"`
	ClosedAt         *time.Time         `json:"closed_at,omitempty"`
}

func (t Tab) MarshalJSON() ([]byte, error) {
	type Alias Tab
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    t.ID.String(),
		Alias: (*Alias)(&t),
	})
}

func (t *Tab) UnmarshalJSON(data []byte) error {
	type Alias Tab
	var v struct {
		ID string `json:"id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseTabID(v.ID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	*t = Tab(*v.Alias)

	return nil
}

// Order represents a group of items ordered together
type Order struct {
	ID     OrderID      `json:"id"`
	Items  []*OrderItem `json:"items"`
	SentAt *time.Time   `json:"sent_at,omitempty"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    o.ID.String(),
		Alias: (*Alias)(&o),
	})
}

func (o *Order) UnmarshalJSON(data []byte) error {
	type Alias Order
	var v struct {
		ID string `json:"id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseOrderID(v.ID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	*o = Order(*v.Alias)

	return nil
}

// OrderItem represents a single item in an order
type OrderItem struct {
	ID               OrderItemID  `json:"id"`
	Quantity         int16        `json:"quantity"`
	Modifiers        []byte       `json:"modifiers"`
	GuestOwnerIDs    []GuestID    `json:"guest_owner_ids"`
	CustomerOwnerIDs []CustomerID `json:"customer_owner_ids"`
	MenuItemID       MenuItemID   `json:"menu_item_id"`
	Name             string       `json:"name"`
	Description      string       `json:"description"`
	PhotoPathinfo    string       `json:"photo_pathinfo"`
	Price            int32        `json:"price"`
	PortionSize      int16        `json:"portion_size"`
	ModifiersConfig  []byte       `json:"modifiers_config"`
}

func (oi OrderItem) MarshalJSON() ([]byte, error) {
	type Alias OrderItem
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    oi.ID.String(),
		Alias: (*Alias)(&oi),
	})
}

func (oi *OrderItem) UnmarshalJSON(data []byte) error {
	type Alias OrderItem
	var v struct {
		ID         string `json:"id"`
		MenuItemID string `json:"menu_item_id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseOrderItemID(v.ID)
	if err != nil {
		return err
	}

	menuItemID, err := ParseMenuItemID(v.MenuItemID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	v.Alias.MenuItemID = menuItemID
	*oi = OrderItem(*v.Alias)

	return nil
}

// Guest represents an unregistered person dining in the restaurant
type Guest struct {
	ID         GuestID `json:"id"`
	CustomName string  `json:"custom_name"`
}

func (g Guest) MarshalJSON() ([]byte, error) {
	type Alias Guest
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    g.ID.String(),
		Alias: (*Alias)(&g),
	})
}

func (g *Guest) UnmarshalJSON(data []byte) error {
	type Alias Guest
	var v struct {
		ID string `json:"id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseGuestID(v.ID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	*g = Guest(*v.Alias)

	return nil
}

// Customer represents a registered person dining in the restaurant
type Customer struct {
	ID          CustomerID `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	PhoneNumber string     `json:"phone_number"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (c Customer) MarshalJSON() ([]byte, error) {
	type Alias Customer
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    c.ID.String(),
		Alias: (*Alias)(&c),
	})
}

func (c *Customer) UnmarshalJSON(data []byte) error {
	type Alias Customer
	var v struct {
		ID string `json:"id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseCustomerID(v.ID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	*c = Customer(*v.Alias)

	return nil
}

// MenuItem represents a food or drink item available for ordering
type MenuItem struct {
	ID              MenuItemID `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	PhotoPathinfo   string     `json:"photo_pathinfo"`
	Price           int32      `json:"price"`
	PortionSize     int16      `json:"portion_size"`
	Available       bool       `json:"available"`
	ModifiersConfig []byte     `json:"modifiers_config"`
	MenuTags        []MenuTag  `json:"menu_tags"`
	CreatedAt       time.Time  `json:"created_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

func (mt MenuItem) MarshalJSON() ([]byte, error) {
	type Alias MenuItem
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    mt.ID.String(),
		Alias: (*Alias)(&mt),
	})
}

func (mt *MenuItem) UnmarshalJSON(data []byte) error {
	type Alias MenuItem
	var v struct {
		ID string `json:"id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseMenuItemID(v.ID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	*mt = MenuItem(*v.Alias)

	return nil
}

// MenuTag represents a label for categorizing menu items
type MenuTag struct {
	ID            MenuTagID        `json:"id"`
	Value         string           `json:"value"`
	Description   string           `json:"description"`
	Dimension     MenuTagDimension `json:"dimension"`
	Prerequisites []MenuTag        `json:"prerequisites"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func (mtd MenuTag) MarshalJSON() ([]byte, error) {
	type Alias MenuTag
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    mtd.ID.String(),
		Alias: (*Alias)(&mtd),
	})
}

func (mtd *MenuTag) UnmarshalJSON(data []byte) error {
	type Alias MenuTag
	var v struct {
		ID string `json:"id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseMenuTagID(v.ID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	*mtd = MenuTag(*v.Alias)

	return nil
}

// MenuTagDimension represents a category for menu tags
type MenuTagDimension struct {
	ID          MenuTagDimensionID `json:"id"`
	Value       string             `json:"value"`
	Description string             `json:"description"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func (mtd MenuTagDimension) MarshalJSON() ([]byte, error) {
	type Alias MenuTagDimension
	return json.Marshal(&struct {
		ID string `json:"id"`
		*Alias
	}{
		ID:    mtd.ID.String(),
		Alias: (*Alias)(&mtd),
	})
}

func (mtd *MenuTagDimension) UnmarshalJSON(data []byte) error {
	type Alias MenuTagDimension
	var v struct {
		ID string `json:"id"`
		*Alias
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	id, err := ParseMenuTagDimensionID(v.ID)
	if err != nil {
		return err
	}

	v.Alias.ID = id
	*mtd = MenuTagDimension(*v.Alias)

	return nil
}

type CreateCustomerParams struct {
	LoginID     LoginID `json:"login_id"`
	Email       string  `json:"email"`
	Password    []byte  `json:"password"`
	Name        string  `json:"name"`
	PhoneNumber string  `json:"phone_number"`
}

type CreateMenuItemParams struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	PhotoPath       string `json:"photo_path"`
	Price           int32  `json:"price"`
	PortionSize     int16  `json:"portion_size"`
	Available       bool   `json:"available"`
	ModifiersConfig []byte `json:"modifiers_config"`
}

type UpdateMenuItemParams CreateMenuItemParams

type CreateOrderItemParams struct {
	OrderID          OrderID      `json:"order_id"`
	MenuItemID       MenuItemID   `json:"menu_item_id"`
	Quantity         int16        `json:"quantity"`
	Modifiers        []byte       `json:"modifiers"`
	GuestOwnerIDs    []GuestID    `json:"guest_owner_ids"`
	CustomerOwnerIDs []CustomerID `json:"customer_owner_ids"`
}

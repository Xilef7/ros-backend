package model

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
)

func TestTabID_String_ParseTabID(t *testing.T) {
	u := uuid.New()
	tabID := TabID(u)
	s := tabID.String()
	if s != u.String() {
		t.Errorf("TabID.String() = %q, want %q", s, u.String())
	}
	got, err := ParseTabID(s)
	if err != nil {
		t.Fatalf("ParseTabID(%q) error: %v", s, err)
	}
	if got != tabID {
		t.Errorf("ParseTabID(%q) = %v, want %v", s, got, tabID)
	}
	_, err = ParseTabID("not-a-uuid")
	if err == nil {
		t.Error("ParseTabID should fail for invalid input")
	}
}

func TestGuestID_String_ParseGuestID(t *testing.T) {
	tabID := TabID(uuid.New())
	scoped := ScopedGuestID(5)
	guestID := GuestID{TabID: tabID, Scoped: scoped}
	s := guestID.String()
	got, err := ParseGuestID(s)
	if err != nil {
		t.Fatalf("ParseGuestID(%q) error: %v", s, err)
	}
	if got != guestID {
		t.Errorf("ParseGuestID(%q) = %v, want %v", s, got, guestID)
	}
	_, err = ParseGuestID("invalid")
	if err == nil {
		t.Error("ParseGuestID should fail for invalid input")
	}
}

func TestOrderID_String_ParseOrderID(t *testing.T) {
	tabID := TabID(uuid.New())
	scoped := ScopedOrderID(7)
	orderID := OrderID{TabID: tabID, Scoped: scoped}
	s := orderID.String()
	got, err := ParseOrderID(s)
	if err != nil {
		t.Fatalf("ParseOrderID(%q) error: %v", s, err)
	}
	if got != orderID {
		t.Errorf("ParseOrderID(%q) = %v, want %v", s, got, orderID)
	}
	_, err = ParseOrderID("invalid")
	if err == nil {
		t.Error("ParseOrderID should fail for invalid input")
	}
}

func TestOrderItemID_String_ParseOrderItemID(t *testing.T) {
	tabID := TabID(uuid.New())
	orderID := OrderID{TabID: tabID, Scoped: ScopedOrderID(3)}
	scoped := ScopedOrderItemID(9)
	orderItemID := OrderItemID{OrderID: orderID, Scoped: scoped}
	s := orderItemID.String()
	got, err := ParseOrderItemID(s)
	if err != nil {
		t.Fatalf("ParseOrderItemID(%q) error: %v", s, err)
	}
	if got != orderItemID {
		t.Errorf("ParseOrderItemID(%q) = %v, want %v", s, got, orderItemID)
	}
	_, err = ParseOrderItemID("invalid")
	if err == nil {
		t.Error("ParseOrderItemID should fail for invalid input")
	}
}

func TestCustomerID_String_ParseCustomerID(t *testing.T) {
	u := uuid.New()
	customerID := CustomerID(u)
	s := customerID.String()
	if s != u.String() {
		t.Errorf("CustomerID.String() = %q, want %q", s, u.String())
	}
	got, err := ParseCustomerID(s)
	if err != nil {
		t.Fatalf("ParseCustomerID(%q) error: %v", s, err)
	}
	if got != customerID {
		t.Errorf("ParseCustomerID(%q) = %v, want %v", s, got, customerID)
	}
	_, err = ParseCustomerID("not-a-uuid")
	if err == nil {
		t.Error("ParseCustomerID should fail for invalid input")
	}
}

func TestMenuItemID_String_ParseMenuItemID(t *testing.T) {
	id := MenuItemID(42)
	s := id.String()
	if s != "42" {
		t.Errorf("MenuItemID.String() = %q, want %q", s, "42")
	}
	got, err := ParseMenuItemID(s)
	if err != nil {
		t.Fatalf("ParseMenuItemID(%q) error: %v", s, err)
	}
	if got != id {
		t.Errorf("ParseMenuItemID(%q) = %v, want %v", s, got, id)
	}
	_, err = ParseMenuItemID("notanumber")
	if err == nil {
		t.Error("ParseMenuItemID should fail for invalid input")
	}
}

func TestMenuTagID_String_ParseMenuTagID(t *testing.T) {
	id := MenuTagID(11)
	s := id.String()
	if s != "11" {
		t.Errorf("MenuTagID.String() = %q, want %q", s, "11")
	}
	got, err := ParseMenuTagID(s)
	if err != nil {
		t.Fatalf("ParseMenuTagID(%q) error: %v", s, err)
	}
	if got != id {
		t.Errorf("ParseMenuTagID(%q) = %v, want %v", s, got, id)
	}
	_, err = ParseMenuTagID("notanumber")
	if err == nil {
		t.Error("ParseMenuTagID should fail for invalid input")
	}
}

func TestMenuTagDimensionID_String_ParseMenuTagDimensionID(t *testing.T) {
	id := MenuTagDimensionID(8)
	s := id.String()
	if s != "8" {
		t.Errorf("MenuTagDimensionID.String() = %q, want %q", s, "8")
	}
	got, err := ParseMenuTagDimensionID(s)
	if err != nil {
		t.Fatalf("ParseMenuTagDimensionID(%q) error: %v", s, err)
	}
	if got != id {
		t.Errorf("ParseMenuTagDimensionID(%q) = %v, want %v", s, got, id)
	}
	_, err = ParseMenuTagDimensionID("notanumber")
	if err == nil {
		t.Error("ParseMenuTagDimensionID should fail for invalid input")
	}
}

func TestLoginID_String_ParseLoginID(t *testing.T) {
	id := LoginID("user123")
	s := id.String()
	if s != "user123" {
		t.Errorf("LoginID.String() = %q, want %q", s, "user123")
	}
	got, err := ParseLoginID(s)
	if err != nil {
		t.Fatalf("ParseLoginID(%q) error: %v", s, err)
	}
	if got != id {
		t.Errorf("ParseLoginID(%q) = %v, want %v", s, got, id)
	}
}

func TestSplitID(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"a.b", []string{"a", "b"}},
		{"a.b.c", []string{"a", "b", "c"}},
		{"abc", []string{"abc"}},
		{"", []string{""}},
	}
	for _, tt := range tests {
		got := splitID(tt.in)
		if len(got) != len(tt.want) {
			t.Errorf("splitID(%q) = %v, want %v", tt.in, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitID(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
			}
		}
	}
}

func TestParseGuestID_InvalidTabID(t *testing.T) {
	_, err := ParseGuestID("notauuid.1")
	if err == nil {
		t.Error("ParseGuestID should fail for invalid TabID")
	}
}

func TestParseOrderID_InvalidTabID(t *testing.T) {
	_, err := ParseOrderID("notauuid.1")
	if err == nil {
		t.Error("ParseOrderID should fail for invalid TabID")
	}
}

func TestParseOrderItemID_InvalidOrderID(t *testing.T) {
	_, err := ParseOrderItemID("notauuid.1.1")
	if err == nil {
		t.Error("ParseOrderItemID should fail for invalid OrderID")
	}
}

func TestParseOrderItemID_InvalidScoped(t *testing.T) {
	tabID := TabID(uuid.New())
	orderID := OrderID{TabID: tabID, Scoped: ScopedOrderID(1)}
	s := orderID.String() + ".notanumber"
	_, err := ParseOrderItemID(s)
	if err == nil {
		t.Error("ParseOrderItemID should fail for invalid ScopedOrderItemID")
	}
}

func TestParseGuestID_InvalidScoped(t *testing.T) {
	tabID := TabID(uuid.New())
	s := tabID.String() + ".notanumber"
	_, err := ParseGuestID(s)
	if err == nil {
		t.Error("ParseGuestID should fail for invalid ScopedGuestID")
	}
}

func TestParseOrderID_InvalidScoped(t *testing.T) {
	tabID := TabID(uuid.New())
	s := tabID.String() + ".notanumber"
	_, err := ParseOrderID(s)
	if err == nil {
		t.Error("ParseOrderID should fail for invalid ScopedOrderID")
	}
}

func TestParseMenuItemID_OutOfRange(t *testing.T) {
	_, err := ParseMenuItemID(strconv.FormatInt(1<<20, 10))
	if err == nil {
		t.Error("ParseMenuItemID should fail for out of int16 range")
	}
}

func TestParseGuestID_NegativeScoped(t *testing.T) {
	tabID := TabID(uuid.New())
	s := tabID.String() + ".-1"
	_, err := ParseGuestID(s)
	if err == nil {
		t.Error("ParseGuestID should fail for negative ScopedGuestID")
	}
}

func TestParseOrderID_NegativeScoped(t *testing.T) {
	tabID := TabID(uuid.New())
	s := tabID.String() + ".-1"
	_, err := ParseOrderID(s)
	if err == nil {
		t.Error("ParseOrderID should fail for negative ScopedOrderID")
	}
}

func TestParseOrderItemID_NegativeScoped(t *testing.T) {
	tabID := TabID(uuid.New())
	orderID := OrderID{TabID: tabID, Scoped: ScopedOrderID(1)}
	s := orderID.String() + ".-1"
	_, err := ParseOrderItemID(s)
	if err == nil {
		t.Error("ParseOrderItemID should fail for negative ScopedOrderItemID")
	}
}

func TestParseMenuItemID_Negative(t *testing.T) {
	_, err := ParseMenuItemID("-1")
	if err == nil {
		t.Error("ParseMenuItemID should fail for negative value")
	}
}

func TestParseMenuTagID_Negative(t *testing.T) {
	_, err := ParseMenuTagID("-1")
	if err == nil {
		t.Error("ParseMenuTagID should fail for negative value")
	}
}

func TestParseMenuTagDimensionID_Negative(t *testing.T) {
	_, err := ParseMenuTagDimensionID("-1")
	if err == nil {
		t.Error("ParseMenuTagDimensionID should fail for negative value")
	}
}

package grpcapp

import (
	"context"

	"restaurant-ordering-system/api/proto"
	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/service"

	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderServiceServer struct {
	proto.UnimplementedOrderServiceServer
	OrderService *service.OrderService
}

func NewOrderServiceServer(orderService *service.OrderService) *OrderServiceServer {
	return &OrderServiceServer{OrderService: orderService}
}

func (s *OrderServiceServer) CreateOrderItem(ctx context.Context, req *proto.CreateOrderItemRequest) (*proto.OrderItemID, error) {
	orderID, err := model.ParseOrderID(req.GetOrderId())
	if err != nil {
		return nil, err
	}
	menuItemID, err := model.ParseMenuItemID(req.GetMenuItemId())
	if err != nil {
		return nil, err
	}
	var guestOwnerIDs []model.GuestID
	for _, gid := range req.GetGuestOwnerIds() {
		parsed, err := model.ParseGuestID(gid)
		if err != nil {
			return nil, err
		}
		guestOwnerIDs = append(guestOwnerIDs, parsed)
	}
	var customerOwnerIDs []model.CustomerID
	for _, cid := range req.GetCustomerOwnerIds() {
		parsed, err := model.ParseCustomerID(cid)
		if err != nil {
			return nil, err
		}
		customerOwnerIDs = append(customerOwnerIDs, parsed)
	}
	params := model.CreateOrderItemParams{
		OrderID:          orderID,
		MenuItemID:       menuItemID,
		Quantity:         int16(req.GetQuantity()),
		Modifiers:        req.GetModifiers(),
		GuestOwnerIDs:    guestOwnerIDs,
		CustomerOwnerIDs: customerOwnerIDs,
	}
	id, err := s.OrderService.CreateOrderItem(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &proto.OrderItemID{}
	resp.SetId(id.String())
	return resp, nil
}

func (s *OrderServiceServer) DeleteOrderItem(ctx context.Context, req *proto.DeleteOrderItemRequest) (*emptypb.Empty, error) {
	id, err := model.ParseOrderItemID(req.GetId())
	if err != nil {
		return nil, err
	}
	if err := s.OrderService.DeleteOrderItem(ctx, id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderServiceServer) UpdateOrderItemModifiers(ctx context.Context, req *proto.UpdateOrderItemModifiersRequest) (*emptypb.Empty, error) {
	id, err := model.ParseOrderItemID(req.GetOrderItemId())
	if err != nil {
		return nil, err
	}
	modifiers := req.GetModifiers()
	if err := s.OrderService.UpdateOrderItemModifiers(ctx, id, modifiers); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderServiceServer) UpdateOrderItemQuantity(ctx context.Context, req *proto.UpdateOrderItemQuantityRequest) (*emptypb.Empty, error) {
	id, err := model.ParseOrderItemID(req.GetOrderItemId())
	if err != nil {
		return nil, err
	}
	if err := s.OrderService.UpdateOrderItemQuantity(ctx, id, int16(req.GetQuantity())); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderServiceServer) AddOrderItemGuestOwner(ctx context.Context, req *proto.AddOrderItemGuestOwnerRequest) (*emptypb.Empty, error) {
	orderItemID, err := model.ParseOrderItemID(req.GetOrderItemId())
	if err != nil {
		return nil, err
	}
	guestID, err := model.ParseGuestID(req.GetGuestId())
	if err != nil {
		return nil, err
	}
	if err := s.OrderService.AddOrderItemGuestOwner(ctx, orderItemID, guestID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderServiceServer) RemoveOrderItemGuestOwner(ctx context.Context, req *proto.RemoveOrderItemGuestOwnerRequest) (*emptypb.Empty, error) {
	orderItemID, err := model.ParseOrderItemID(req.GetOrderItemId())
	if err != nil {
		return nil, err
	}
	guestID, err := model.ParseGuestID(req.GetGuestId())
	if err != nil {
		return nil, err
	}
	if err := s.OrderService.RemoveOrderItemGuestOwner(ctx, orderItemID, guestID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderServiceServer) AddOrderItemCustomerOwner(ctx context.Context, req *proto.AddOrderItemCustomerOwnerRequest) (*emptypb.Empty, error) {
	orderItemID, err := model.ParseOrderItemID(req.GetOrderItemId())
	if err != nil {
		return nil, err
	}
	customerID, err := model.ParseCustomerID(req.GetCustomerId())
	if err != nil {
		return nil, err
	}
	if err := s.OrderService.AddOrderItemCustomerOwner(ctx, orderItemID, customerID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderServiceServer) RemoveOrderItemCustomerOwner(ctx context.Context, req *proto.RemoveOrderItemCustomerOwnerRequest) (*emptypb.Empty, error) {
	orderItemID, err := model.ParseOrderItemID(req.GetOrderItemId())
	if err != nil {
		return nil, err
	}
	customerID, err := model.ParseCustomerID(req.GetCustomerId())
	if err != nil {
		return nil, err
	}
	if err := s.OrderService.RemoveOrderItemCustomerOwner(ctx, orderItemID, customerID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *OrderServiceServer) SendOrder(ctx context.Context, req *proto.SendOrderRequest) (*emptypb.Empty, error) {
	orderID, err := model.ParseOrderID(req.GetOrderId())
	if err != nil {
		return nil, err
	}
	if err := s.OrderService.SendOrder(ctx, orderID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

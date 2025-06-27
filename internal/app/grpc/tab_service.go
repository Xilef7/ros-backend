package grpcapp

import (
	"context"
	"errors"

	"restaurant-ordering-system/api/proto"
	"restaurant-ordering-system/internal/pkg/auth"
	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/service"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TabServiceServer struct {
	proto.UnimplementedTabServiceServer
	TabService *service.TabService
}

func NewTabServiceServer(tabService *service.TabService) *TabServiceServer {
	return &TabServiceServer{TabService: tabService}
}

func (s *TabServiceServer) CreateTab(ctx context.Context, req *emptypb.Empty) (*proto.TabID, error) {
	tabID, err := s.TabService.CreateTab(ctx)
	if err != nil {
		return nil, err
	}
	resp := &proto.TabID{}
	resp.SetId(tabID.String())
	return resp, nil
}

func (s *TabServiceServer) VisitTab(ctx context.Context, req *proto.VisitTabRequest) (*emptypb.Empty, error) {
	claims, ok := auth.FromContext(ctx)
	if !ok {
		return nil, errors.New("not authenticated")
	}
	subjectID, err := model.ParseCustomerID(claims.Subject)
	if err != nil {
		return nil, err
	}
	customerID, err := model.ParseCustomerID(req.GetCustomerId())
	if err != nil {
		return nil, err
	}
	if subjectID != customerID {
		return nil, errors.New("not authorized")
	}
	tabID, err := model.ParseTabID(req.GetTabId())
	if err != nil {
		return nil, err
	}
	if err := s.TabService.VisitTab(ctx, tabID, customerID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *TabServiceServer) CreateGuest(ctx context.Context, req *proto.CreateGuestRequest) (*proto.GuestID, error) {
	tabID, err := model.ParseTabID(req.GetTabId())
	if err != nil {
		return nil, err
	}
	guestID, err := s.TabService.CreateGuest(ctx, tabID)
	if err != nil {
		return nil, err
	}
	resp := &proto.GuestID{}
	resp.SetId(guestID.String())
	return resp, nil
}

func (s *TabServiceServer) UpdateGuestName(ctx context.Context, req *proto.UpdateGuestNameRequest) (*emptypb.Empty, error) {
	guestID, err := model.ParseGuestID(req.GetGuestId())
	if err != nil {
		return nil, err
	}
	if err := s.TabService.UpdateGuestName(ctx, guestID, req.GetGuestId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *TabServiceServer) GetOpenTab(ctx context.Context, req *proto.GetOpenTabRequest) (*proto.Tab, error) {
	tabID, err := model.ParseTabID(req.GetTabId())
	if err != nil {
		return nil, err
	}
	tab, err := s.TabService.GetOpenTab(ctx, tabID)
	if err != nil {
		return nil, err
	}
	return modelTabToProtoTab(tab), nil
}

func (s *TabServiceServer) CloseTab(ctx context.Context, req *proto.CloseTabRequest) (*proto.CloseTabResponse, error) {
	tabID, err := model.ParseTabID(req.GetTabId())
	if err != nil {
		return nil, err
	}
	closedAt, err := s.TabService.CloseTab(ctx, tabID)
	if err != nil {
		return nil, err
	}
	resp := &proto.CloseTabResponse{}
	resp.SetClosedAt(timestamppb.New(closedAt))
	return resp, nil
}

func (s *TabServiceServer) GetVisitedTabs(ctx context.Context, req *proto.GetVisitedTabsRequest) (*proto.GetVisitedTabsResponse, error) {
	claims, ok := auth.FromContext(ctx)
	if !ok {
		return nil, errors.New("not authenticated")
	}
	subjectID, err := model.ParseCustomerID(claims.Subject)
	if err != nil {
		return nil, err
	}
	customerID, err := model.ParseCustomerID(req.GetCustomerId())
	if err != nil {
		return nil, err
	}
	if subjectID != customerID {
		return nil, errors.New("not authorized")
	}
	tabs, err := s.TabService.GetVisitedTabs(ctx, customerID)
	if err != nil {
		return nil, err
	}
	resp := &proto.GetVisitedTabsResponse{}
	var protoTabs []*proto.Tab
	for _, tab := range tabs {
		protoTabs = append(protoTabs, modelTabToProtoTab(tab))
	}
	resp.SetTabs(protoTabs)
	return resp, nil
}

func modelTabToProtoTab(tab *model.Tab) *proto.Tab {
	ptab := &proto.Tab{}
	ptab.SetId(tab.ID.String())
	ptab.SetTotalPrice(tab.TotalPrice)
	ptab.SetCreatedAt(timestamppb.New(tab.CreatedAt))
	if tab.ClosedAt != nil {
		ptab.SetClosedAt(timestamppb.New(*tab.ClosedAt))
	}
	var protoOrders []*proto.Order
	for _, order := range tab.Orders {
		protoOrders = append(protoOrders, modelOrderToProtoOrder(order))
	}
	ptab.SetOrders(protoOrders)
	guestNames := make(map[string]string)
	for gid, name := range tab.CustomGuestNames {
		guestNames[gid.String()] = name
	}
	ptab.SetCustomGuestNames(guestNames)
	return ptab
}

func modelOrderToProtoOrder(order *model.Order) *proto.Order {
	po := &proto.Order{}
	po.SetId(order.ID.String())
	if order.SentAt != nil {
		po.SetSentAt(timestamppb.New(*order.SentAt))
	}
	var protoItems []*proto.OrderItem
	for _, item := range order.Items {
		protoItems = append(protoItems, modelOrderItemToProtoOrderItem(item))
	}
	po.SetItems(protoItems)
	return po
}

func modelOrderItemToProtoOrderItem(item *model.OrderItem) *proto.OrderItem {
	poi := &proto.OrderItem{}
	poi.SetId(item.ID.String())
	poi.SetQuantity(int32(item.Quantity))
	poi.SetModifiers(item.Modifiers)
	var guestOwnerIDs []string
	for _, gid := range item.GuestOwnerIDs {
		guestOwnerIDs = append(guestOwnerIDs, gid.String())
	}
	poi.SetGuestOwnerIds(guestOwnerIDs)
	var customerOwnerIDs []string
	for _, cid := range item.CustomerOwnerIDs {
		customerOwnerIDs = append(customerOwnerIDs, cid.String())
	}
	poi.SetCustomerOwnerIds(customerOwnerIDs)
	poi.SetMenuItemId(item.MenuItemID.String())
	poi.SetName(item.Name)
	poi.SetDescription(item.Description)
	poi.SetPhotoPathinfo(item.PhotoPathinfo)
	poi.SetPrice(item.Price)
	poi.SetPortionSize(int32(item.PortionSize))
	poi.SetModifiersConfig(item.ModifiersConfig)
	return poi
}

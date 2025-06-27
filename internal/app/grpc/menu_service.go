package grpcapp

import (
	"context"

	"restaurant-ordering-system/api/proto"
	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/service"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MenuServiceServer struct {
	proto.UnimplementedMenuServiceServer
	MenuService *service.MenuService
}

func NewMenuServiceServer(menuService *service.MenuService) *MenuServiceServer {
	return &MenuServiceServer{MenuService: menuService}
}

func (s *MenuServiceServer) CreateMenuItem(ctx context.Context, req *proto.CreateMenuItemRequest) (*proto.MenuItem, error) {
	item := req.GetMenuItem()
	params := model.CreateMenuItemParams{
		Name:            item.GetName(),
		Description:     item.GetDescription(),
		PhotoPath:       item.GetPhotoPathinfo(),
		Price:           item.GetPrice(),
		PortionSize:     int16(item.GetPortionSize()),
		Available:       item.GetAvailable(),
		ModifiersConfig: item.GetModifiersConfig(),
	}
	modelItem, err := s.MenuService.CreateMenuItem(ctx, params)
	if err != nil {
		return nil, err
	}
	return modelMenuItemToProtoMenuItem(modelItem), nil
}

func (s *MenuServiceServer) GetMenuItem(ctx context.Context, req *proto.GetMenuItemRequest) (*proto.MenuItem, error) {
	id, err := model.ParseMenuItemID(req.GetId())
	if err != nil {
		return nil, err
	}
	item, err := s.MenuService.GetMenuItem(ctx, id)
	if err != nil {
		return nil, err
	}
	return modelMenuItemToProtoMenuItem(item), nil
}

func (s *MenuServiceServer) ListMenuItems(ctx context.Context, req *emptypb.Empty) (*proto.ListMenuItemsResponse, error) {
	items, err := s.MenuService.ListMenuItems(ctx)
	if err != nil {
		return nil, err
	}
	var protoItems []*proto.MenuItem
	for _, item := range items {
		protoItems = append(protoItems, modelMenuItemToProtoMenuItem(item))
	}
	resp := &proto.ListMenuItemsResponse{}
	resp.SetItems(protoItems)
	return resp, nil
}

func (s *MenuServiceServer) UpdateMenuItem(ctx context.Context, req *proto.UpdateMenuItemRequest) (*proto.MenuItem, error) {
	item := req.GetMenuItem()
	id, err := model.ParseMenuItemID(item.GetId())
	if err != nil {
		return nil, err
	}
	params := model.UpdateMenuItemParams{
		Name:            item.GetName(),
		Description:     item.GetDescription(),
		PhotoPath:       item.GetPhotoPathinfo(),
		Price:           item.GetPrice(),
		PortionSize:     int16(item.GetPortionSize()),
		Available:       item.GetAvailable(),
		ModifiersConfig: item.GetModifiersConfig(),
	}
	modelItem, err := s.MenuService.UpdateMenuItem(ctx, id, params)
	if err != nil {
		return nil, err
	}
	return modelMenuItemToProtoMenuItem(modelItem), nil
}

func (s *MenuServiceServer) DeleteMenuItem(ctx context.Context, req *proto.DeleteMenuItemRequest) (*emptypb.Empty, error) {
	id, err := model.ParseMenuItemID(req.GetId())
	if err != nil {
		return nil, err
	}
	if err := s.MenuService.DeleteMenuItem(ctx, id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func modelMenuItemToProtoMenuItem(item *model.MenuItem) *proto.MenuItem {
	mi := &proto.MenuItem{}
	mi.SetId(item.ID.String())
	mi.SetName(item.Name)
	mi.SetDescription(item.Description)
	mi.SetPhotoPathinfo(item.PhotoPathinfo)
	mi.SetPrice(item.Price)
	mi.SetPortionSize(int32(item.PortionSize))
	mi.SetAvailable(item.Available)
	mi.SetModifiersConfig(item.ModifiersConfig)
	var protoTags []*proto.MenuTag
	for _, tag := range item.MenuTags {
		protoTags = append(protoTags, modelMenuTagToProtoMenuTag(&tag))
	}
	mi.SetMenuTags(protoTags)
	mi.SetCreatedAt(timestamppb.New(item.CreatedAt))
	if item.DeletedAt != nil {
		mi.SetDeletedAt(timestamppb.New(*item.DeletedAt))
	}
	return mi
}

func modelMenuTagToProtoMenuTag(tag *model.MenuTag) *proto.MenuTag {
	mt := &proto.MenuTag{}
	mt.SetId(tag.ID.String())
	mt.SetValue(tag.Value)
	mt.SetDescription(tag.Description)
	mt.SetDimension(modelMenuTagDimensionToProto(&tag.Dimension))
	var protoPrereqs []*proto.MenuTag
	for _, pre := range tag.Prerequisites {
		protoPrereqs = append(protoPrereqs, modelMenuTagToProtoMenuTag(&pre))
	}
	mt.SetPrerequisites(protoPrereqs)
	mt.SetCreatedAt(timestamppb.New(tag.CreatedAt))
	mt.SetUpdatedAt(timestamppb.New(tag.UpdatedAt))
	return mt
}

func modelMenuTagDimensionToProto(dim *model.MenuTagDimension) *proto.MenuTagDimension {
	pd := &proto.MenuTagDimension{}
	pd.SetId(dim.ID.String())
	pd.SetValue(dim.Value)
	pd.SetDescription(dim.Description)
	pd.SetCreatedAt(timestamppb.New(dim.CreatedAt))
	pd.SetUpdatedAt(timestamppb.New(dim.UpdatedAt))
	return pd
}

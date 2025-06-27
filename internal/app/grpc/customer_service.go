package grpcapp

import (
	"context"

	"restaurant-ordering-system/api/proto"
	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CustomerServiceServer struct {
	proto.UnimplementedCustomerServiceServer
	CustomerService *service.CustomerService
}

func NewCustomerServiceServer(customerService *service.CustomerService) *CustomerServiceServer {
	return &CustomerServiceServer{CustomerService: customerService}
}

func (s *CustomerServiceServer) CreateCustomer(ctx context.Context, req *proto.CreateCustomerRequest) (*proto.Customer, error) {
	loginID, err := model.ParseLoginID(req.GetLoginId())
	if err != nil {
		return nil, err
	}
	params := model.CreateCustomerParams{
		LoginID:     loginID,
		Email:       req.GetEmail(),
		Password:    []byte(req.GetPassword()),
		Name:        req.GetName(),
		PhoneNumber: req.GetPhoneNumber(),
	}
	customer, err := s.CustomerService.CreateCustomer(ctx, params)
	if err != nil {
		return nil, err
	}
	return modelCustomerToProtoCustomer(&customer), nil
}

func (s *CustomerServiceServer) GetCustomerByID(ctx context.Context, req *proto.GetCustomerByIDRequest) (*proto.Customer, error) {
	id, err := model.ParseCustomerID(req.GetId())
	if err != nil {
		return nil, err
	}
	customer, err := s.CustomerService.GetCustomerByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return modelCustomerToProtoCustomer(&customer), nil
}

func modelCustomerToProtoCustomer(customer *model.Customer) *proto.Customer {
	pcust := &proto.Customer{}
	pcust.SetId(customer.ID.String())
	pcust.SetName(customer.Name)
	pcust.SetEmail(customer.Email)
	pcust.SetPhoneNumber(customer.PhoneNumber)
	pcust.SetCreatedAt(timestamppb.New(customer.CreatedAt))
	pcust.SetUpdatedAt(timestamppb.New(customer.UpdatedAt))
	return pcust
}

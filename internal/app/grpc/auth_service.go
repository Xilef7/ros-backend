package grpcapp

import (
	"context"

	"restaurant-ordering-system/api/proto"
	"restaurant-ordering-system/internal/pkg/model"
	"restaurant-ordering-system/internal/pkg/service"
)

type AuthServiceServer struct {
	proto.UnimplementedAuthServiceServer
	AuthService *service.AuthService
}

func NewAuthServiceServer(authService *service.AuthService) *AuthServiceServer {
	return &AuthServiceServer{AuthService: authService}
}

func (s *AuthServiceServer) GenerateToken(ctx context.Context, req *proto.GenerateTokenRequest) (*proto.GenerateTokenResponse, error) {
	loginID := model.LoginID(req.GetLoginId())
	password := req.GetPassword()
	token, err := s.AuthService.GenerateToken(ctx, loginID, password)
	if err != nil {
		return nil, err
	}
	resp := &proto.GenerateTokenResponse{}
	resp.SetAccessToken(token)
	resp.SetExpiresIn(86400) // TODO: set real expiry
	return resp, nil
}

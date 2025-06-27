package middleware

import (
	"context"
	"strings"

	"restaurant-ordering-system/internal/pkg/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var openMethods = map[string]bool{
	"/restaurant.CustomerService/CreateCustomer":            true,
	"/restaurant.AuthService/GenerateToken":                 true,
	"/restaurant.MenuService/GetMenuItem":                   true,
	"/restaurant.MenuService/ListMenuItems":                 true,
	"/restaurant.OrderService/CreateOrderItem":              true,
	"/restaurant.OrderService/DeleteOrderItem":              true,
	"/restaurant.OrderService/UpdateOrderItemModifiers":     true,
	"/restaurant.OrderService/UpdateOrderItemQuantity":      true,
	"/restaurant.OrderService/AddOrderItemGuestOwner":       true,
	"/restaurant.OrderService/RemoveOrderItemGuestOwner":    true,
	"/restaurant.OrderService/AddOrderItemCustomerOwner":    true,
	"/restaurant.OrderService/RemoveOrderItemCustomerOwner": true,
	"/restaurant.OrderService/SendOrder":                    true,
	"/restaurant.TabService/CreateGuest":                    true,
	"/restaurant.TabService/UpdateGuestName":                true,
	"/restaurant.TabService/GetOpenTab":                     true,
	"/restaurant.TabService/CloseTab":                       true,
}

var adminMethods = map[string]bool{
	"/restaurant.MenuService/CreateMenuItem": true,
	"/restaurant.MenuService/UpdateMenuItem": true,
	"/restaurant.MenuService/DeleteMenuItem": true,
	"/restaurant.TabService/CreateTab":       true,
}

func NewJWTUnaryInterceptor(parse auth.JWTParser) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if openMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		authorization := md["authorization"]
		if len(authorization) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization metadata missing")
		}

		tokenString := strings.TrimPrefix(authorization[0], "Bearer ")
		claims, err := parse(tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		if adminMethods[info.FullMethod] && claims.Role != auth.AdminRole {
			return nil, status.Error(codes.PermissionDenied, "not authorized")
		}

		ctx = auth.NewContext(ctx, claims)

		return handler(ctx, req)
	}
}

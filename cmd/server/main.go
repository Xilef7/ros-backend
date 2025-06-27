// Package main provides the entry point for the gRPC server
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"restaurant-ordering-system/api/proto"
	grpcapp "restaurant-ordering-system/internal/app/grpc"
	"restaurant-ordering-system/internal/pkg/auth"
	"restaurant-ordering-system/internal/pkg/config"
	"restaurant-ordering-system/internal/pkg/middleware"
	"restaurant-ordering-system/internal/pkg/service"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Load configuration
	cfg, err := config.LoadConfig("configs/config.json")
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	cert, err := tls.LoadX509KeyPair("certs/server_cert.pem", "certs/server_key.pem")
	if err != nil {
		logger.Error("Failed to load key pair", "error", err)
		os.Exit(1)
	}

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)
	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("Failed to create dbpool", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	// Initialize JWT generator
	jwtGenerator := auth.NewCustomerJWTGenerator([]byte(cfg.JWT.Secret), cfg.JWT.Expiry)

	// Initialize services
	customerService := service.NewCustomerService(dbpool)
	authService := service.NewAuthService(dbpool, jwtGenerator)
	menuService := service.NewMenuService(dbpool)
	orderService := service.NewOrderService(dbpool)
	tabService := service.NewTabService(dbpool)

	// Initialize JWT parser
	jwtParser := auth.NewJWTParser([]byte(cfg.JWT.Secret))

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.NewJWTUnaryInterceptor(jwtParser),
			middleware.UnaryServerInterceptor(logger),
		),
		grpc.StreamInterceptor(middleware.StreamServerInterceptor(logger)),
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	)

	// Register services
	grpcappCustomerService := grpcapp.NewCustomerServiceServer(customerService)
	proto.RegisterCustomerServiceServer(grpcServer, grpcappCustomerService)
	grpcappAuthService := grpcapp.NewAuthServiceServer(authService)
	proto.RegisterAuthServiceServer(grpcServer, grpcappAuthService)
	grpcappMenuService := grpcapp.NewMenuServiceServer(menuService)
	proto.RegisterMenuServiceServer(grpcServer, grpcappMenuService)
	grpcappOrderService := grpcapp.NewOrderServiceServer(orderService)
	proto.RegisterOrderServiceServer(grpcServer, grpcappOrderService)
	grpcappTabService := grpcapp.NewTabServiceServer(tabService)
	proto.RegisterTabServiceServer(grpcServer, grpcappTabService)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("Failed to listen", "error", err)
		dbpool.Close()
		os.Exit(1)
	}

	go func() {
		logger.Info("Starting gRPC server", "address", addr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("Failed to serve", "error", err)
			dbpool.Close()
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown
	logger.Info("Shutting down gRPC server")
	grpcServer.GracefulStop()
}

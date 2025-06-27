// Package middleware provides middleware functions for the gRPC server
package middleware

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor creates a new unary server interceptor for logging and tracing
func UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		md, _ := metadata.FromIncomingContext(ctx)

		// Start a new span
		tracer := otel.Tracer("grpc.server")
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithAttributes())
		defer span.End()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		attrs := []any{
			"method", info.FullMethod,
			"duration", duration,
			"metadata", md,
		}

		if err != nil {
			st, _ := status.FromError(err)
			attrs = append(attrs,
				"code", st.Code().String(),
				"error", st.Message(),
			)
			logger.Error("gRPC request failed", attrs...)
		} else {
			attrs = append(attrs, "code", codes.OK.String())
			logger.Info("gRPC request successful", attrs...)
		}

		return resp, err
	}
}

// StreamServerInterceptor creates a new stream server interceptor for logging and tracing
func StreamServerInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		md, _ := metadata.FromIncomingContext(ss.Context())

		// Start a new span
		tracer := otel.Tracer("grpc.server")
		ctx, span := tracer.Start(ss.Context(), info.FullMethod, trace.WithAttributes())
		defer span.End()

		// Wrap the stream to propagate the context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Call the handler
		err := handler(srv, wrapped)

		// Log the stream
		duration := time.Since(start)
		attrs := []any{
			"method", info.FullMethod,
			"duration", duration,
			"metadata", md,
		}

		if err != nil {
			st, _ := status.FromError(err)
			attrs = append(attrs,
				"code", st.Code().String(),
				"error", st.Message(),
			)
			logger.Error("gRPC stream failed", attrs...)
		} else {
			attrs = append(attrs, "code", codes.OK.String())
			logger.Info("gRPC stream successful", attrs...)
		}

		return err
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

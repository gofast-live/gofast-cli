package grpc

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func SlogStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startTime := time.Now()
		ctx := ss.Context()

		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		var userAgent string
		var authTokenPresent bool
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ua := md.Get("user-agent"); len(ua) > 0 {
				userAgent = ua[0]
			}
			if auth := md.Get("authorization"); len(auth) > 0 && auth[0] != "" {
				authTokenPresent = true
			}
		}

		slog.Info("gRPC stream started",
			slog.String("method", info.FullMethod),
			slog.Bool("is_client_stream", info.IsClientStream),
			slog.Bool("is_server_stream", info.IsServerStream),
			slog.String("remote_addr", clientIP),
			slog.String("user_agent", userAgent),
			slog.Bool("auth_token_present", authTokenPresent),
		)

		err := handler(srv, ss)

		duration := time.Since(startTime)
		statusCode := status.Code(err)

		attrs := []slog.Attr{
			slog.String("grpc_method", info.FullMethod),
			slog.String("grpc_status", statusCode.String()),
			slog.Duration("duration", duration),
			slog.Bool("is_client_stream", info.IsClientStream),
			slog.Bool("is_server_stream", info.IsServerStream),
		}

		logLevel := slog.LevelInfo
		if err != nil {
			if statusCode >= codes.Internal {
				logLevel = slog.LevelError
			} else if statusCode >= codes.InvalidArgument && statusCode < codes.Internal {
				logLevel = slog.LevelWarn
			}
		}
		if err != nil {
			attrs = append(attrs, slog.Any("grpc_error", err.Error()))
		}

		slog.LogAttrs(ctx, logLevel, "gRPC stream finished", attrs...)

		return err
	}
}

func SlogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		startTime := time.Now()

		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		var userAgent string
		var authTokenPresent bool
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ua := md.Get("user-agent"); len(ua) > 0 {
				userAgent = ua[0]
			}
			if auth := md.Get("authorization"); len(auth) > 0 && auth[0] != "" {
				authTokenPresent = true
			}
		}

		slog.Info("gRPC call started",
			slog.String("grpc_method", info.FullMethod),
			slog.String("remote_addr", clientIP),
			slog.String("user_agent", userAgent),
			slog.Bool("auth_token_present", authTokenPresent),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		statusCode := status.Code(err)

		attrs := []slog.Attr{
			slog.String("grpc_method", info.FullMethod),
			slog.String("grpc_status", statusCode.String()),
			slog.Duration("duration", duration),
			slog.String("remote_addr", clientIP),
			slog.String("user_agent", userAgent),
			slog.Bool("auth_token_present", authTokenPresent),
		}

		logLevel := slog.LevelInfo
		if err != nil {
			if statusCode >= codes.Internal {
				logLevel = slog.LevelError
			} else if statusCode >= codes.InvalidArgument && statusCode < codes.Internal {
				logLevel = slog.LevelWarn
			}
		}
		if err != nil {
			attrs = append(attrs, slog.Any("grpc_error", err.Error()))
		}
		slog.LogAttrs(ctx, logLevel, "gRPC call finished", attrs...)

		return resp, err
	}
}

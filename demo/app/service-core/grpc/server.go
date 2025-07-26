package grpc

import (
	"app/pkg"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "service-core/proto"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer

	handler *Handler
}

type userServer struct {
	pb.UnimplementedUserServiceServer

	handler *Handler
}
type noteServer struct {
	pb.UnimplementedNoteServiceServer

	handler *Handler
}

func Run(handler *Handler) {
	cfg := handler.cfg
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.GRPCPort))
	if err != nil {
		slog.Error("Error listening on gRPC port", "error", err)
		panic(err)
	}
	unaryLogger := SlogUnaryServerInterceptor()
	streamLogger := SlogStreamServerInterceptor()
	s := grpc.NewServer(grpc.UnaryInterceptor(unaryLogger), grpc.StreamInterceptor(streamLogger))
	pb.RegisterAuthServiceServer(s, &authServer{
		UnimplementedAuthServiceServer: pb.UnimplementedAuthServiceServer{},
		handler:                        handler,
	})
	pb.RegisterUserServiceServer(s, &userServer{
		UnimplementedUserServiceServer: pb.UnimplementedUserServiceServer{},
		handler:                        handler,
	})
	pb.RegisterNoteServiceServer(s, &noteServer{
		UnimplementedNoteServiceServer: pb.UnimplementedNoteServiceServer{},
		handler:                        handler,
	})
	go func() {
		slog.Info("gRPC server listening on", "port", cfg.GRPCPort)
		err = s.Serve(lis)
		if err != nil {
			slog.Error("Error serving gRPC", "error", err)
			panic(err)
		}
	}()
}

func getToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	token := strings.Join(md.Get("Authorization"), "")
	return token
}

func writeResponse(err error) error {
	if err != nil {
		var unauthorizedError pkg.UnauthorizedError
		if errors.As(err, &unauthorizedError) {
			return status.Errorf(codes.Unauthenticated, "Unauthorized")
		}
		var internalError pkg.InternalError
		if errors.As(err, &internalError) {
			return status.Errorf(codes.Internal, "Internal error: %s", internalError.Message)
		}
		var notFoundError pkg.NotFoundError
		if errors.As(err, &notFoundError) {
			return status.Errorf(codes.NotFound, "Not found: %s", notFoundError.Message)
		}
		var validationErrors pkg.ValidationErrors
		if errors.As(err, &validationErrors) {
			return status.Errorf(codes.InvalidArgument, "%s", validationErrors.Error())
		}
		return status.Errorf(codes.Internal, "Internal error: %s", err.Error())
	}
	return nil
}

package grpc

import (
	"app/pkg/auth"
	"context"
	"time"

	pb "service-core/proto"
	"service-core/storage/query"
)

func toProtoUser(user *query.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Id:                 user.ID.String(),
		Created:            user.Created.Format(time.RFC3339),
		Updated:            user.Updated.Format(time.RFC3339),
		Email:              user.Email,
		Access:             user.Access,
		Sub:                user.Sub,
		Avatar:             user.Avatar,
		SubscriptionId:     user.SubscriptionID,
		SubscriptionEnd:    user.SubscriptionEnd.Format(time.RFC3339),
		SubscriptionActive: user.SubscriptionEnd.After(time.Now()),
	}
}

func (s *userServer) GetAllUsers(_ *pb.Empty, stream pb.UserService_GetAllUsersServer) error {
	token := getToken(stream.Context())
	_, err := s.handler.authService.Auth(token, auth.GetUsers)
	if err != nil {
		return writeResponse(err)
	}
	var process = func(_ context.Context, user *query.User) error {
		if user == nil {
			return nil
		}
		err := stream.Send(toProtoUser(user))
		if err != nil {
			return writeResponse(err)
		}
		return nil
	}
	r := s.handler.userService.GetAllUsers(stream.Context(), process)
	return writeResponse(r)
}

func (s *userServer) GetUserByID(ctx context.Context, id *pb.ID) (*pb.User, error) {
	token := getToken(ctx)
	_, err := s.handler.authService.Auth(token, auth.GetUsers)
	if err != nil {
		return nil, writeResponse(err)
	}
	user, err := s.handler.userService.GetUserByID(ctx, id.GetId())
	if err != nil {
		return nil, writeResponse(err)
	}
	return toProtoUser(user), nil
}

func (s *userServer) EditUserAccess(ctx context.Context, in *pb.User) (*pb.User, error) {
	token := getToken(ctx)
	_, err := s.handler.authService.Auth(token, auth.EditUser)
	if err != nil {
		return nil, writeResponse(err)
	}
	user, err := s.handler.userService.EditUserAccess(ctx, in.GetId(), in.GetAccess())
	if err != nil {
		return nil, writeResponse(err)
	}
	return toProtoUser(user), nil
}

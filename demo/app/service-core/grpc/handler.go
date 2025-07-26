package grpc

import (
	"app/pkg/auth"
	"service-core/domain/user"
	"service-core/config"
)

type Handler struct {
	cfg          *config.Config
	authService  *auth.Service
	userService  *user.Service
}

func NewHandler(
	cfg *config.Config,
	authService *auth.Service,
	userService *user.Service,
) *Handler {
	return &Handler{
		cfg:          cfg,
		authService:  authService,
		userService:  userService,
	}
}

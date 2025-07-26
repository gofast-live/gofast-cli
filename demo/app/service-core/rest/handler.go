package rest

import (
	"app/pkg/auth"
	"service-core/config"
	"service-core/domain/user"
	"service-core/storage/query"
)

type Handler struct {
	Cfg         *config.Config
	Store       *query.Queries
	AuthService *auth.Service
	UserService *user.Service
}

func NewHandler(
	config *config.Config,
	store *query.Queries,
	authService *auth.Service,
	userService *user.Service,
) *Handler {
	return &Handler{
		Cfg:         config,
		Store:       store,
		AuthService: authService,
		UserService: userService,
	}
}

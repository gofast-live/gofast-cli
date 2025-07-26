package server

import (
	"app/pkg/auth"
	"service-core/config"
	"service-core/domain/user"
	"service-core/grpc"
	"service-core/rest"
	"service-core/storage"
	"service-core/storage/query"
)

type Server struct {
	Config      *config.Config
	Storage     *storage.Storage
	GRPCServer  *grpc.Handler
	RESTServer  *rest.Handler
	AuthService *auth.Service
	UserService *user.Service
}

func New(cfg *config.Config, s *storage.Storage) *Server {
	store := query.New(s.Conn)
	authService := auth.NewService()
	userService := user.NewService(cfg, store)

	return &Server{
		Config:      cfg,
		Storage:     s,
		GRPCServer:  grpc.NewHandler(cfg, authService, userService),
		RESTServer:  rest.NewHandler(cfg, store, authService, userService),
		AuthService: authService,
		UserService: userService,
	}
}

func (s *Server) Start() {
	rest.Run(s.RESTServer)
	grpc.Run(s.GRPCServer)
}

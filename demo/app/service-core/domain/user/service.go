package user

import (
	"app/pkg"
	"context"
	"service-core/config"
	"service-core/storage/query"

	"github.com/google/uuid"
)

type store interface {
	SelectUsers(ctx context.Context) ([]query.User, error)
	SelectUser(ctx context.Context, id uuid.UUID) (query.User, error)
	UpdateUserAccess(ctx context.Context, params query.UpdateUserAccessParams) (query.User, error)
}

type Service struct {
	cfg   *config.Config
	store store
}

func NewService(
	cfg *config.Config,
	store store,
) *Service {
	return &Service{
		cfg:   cfg,
		store: store,
	}
}

func (s *Service) GetAllUsers(ctx context.Context, process func(ctx context.Context, user *query.User) error) error {
	users, err := s.store.SelectUsers(ctx)
	if err != nil {
		return pkg.NotFoundError{Message: "Error selecting users", Err: err}
	}
	for _, user := range users {
		err := process(ctx, &user)
		if err != nil {
			return pkg.InternalError{Message: "Error processing user", Err: err}
		}
	}
	return nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*query.User, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, pkg.InternalError{Message: "Error parsing UUID", Err: err}
	}
	user, err := s.store.SelectUser(ctx, uuid)
	if err != nil {
		return nil, pkg.NotFoundError{Message: "Error selecting user by ID", Err: err}
	}
	return &user, nil
}

func (s *Service) EditUserAccess(ctx context.Context, id string, access int64) (*query.User, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, pkg.InternalError{Message: "Error parsing UUID", Err: err}
	}
	params := query.UpdateUserAccessParams{
		ID:     uuid,
		Access: access,
	}
	user, err := s.store.UpdateUserAccess(ctx, params)
	if err != nil {
		return nil, pkg.NotFoundError{Message: "Error updating user access by ID", Err: err}
	}
	return &user, nil
}

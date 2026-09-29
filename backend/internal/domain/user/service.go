package user

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo IUserRepository
}

func NewService(repo IUserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

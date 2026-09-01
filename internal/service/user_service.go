// Package service: business logic, unaware of both HTTP and SQL.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type userService struct {
	userRepo     domain.UserRepository
	hasher       domain.PasswordHasher
	auditLogRepo domain.AuditLogRepository
}

func NewUserService(userRepo domain.UserRepository, hasher domain.PasswordHasher, auditLogRepo domain.AuditLogRepository) domain.UserService {
	return &userService{userRepo: userRepo, hasher: hasher, auditLogRepo: auditLogRepo}
}

func (s *userService) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// An unknown role id is rejected by the foreign key, not here.
	if input.RoleID == uuid.Nil {
		input.RoleID = domain.RoleIDUser
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: email is already registered", domain.ErrConflict)
	}

	hashed, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	user := &domain.User{
		ID:           id,
		Email:        email,
		FullName:     strings.TrimSpace(input.FullName),
		PasswordHash: hashed,
		RoleID:       input.RoleID,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return s.userRepo.List(ctx, filter)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateUserInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.FullName != nil {
		name := strings.TrimSpace(*input.FullName)
		if name == "" {
			return nil, fmt.Errorf("%w: full_name must not be empty", domain.ErrInvalidInput)
		}
		user.FullName = name
	}
	if input.RoleID != nil {
		if *input.RoleID == uuid.Nil {
			return nil, fmt.Errorf("%w: role_id must not be empty", domain.ErrInvalidInput)
		}
		user.RoleID = *input.RoleID
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type roleService struct {
	roleRepo     domain.RoleRepository
	auditLogRepo domain.AuditLogRepository
}

func NewRoleService(roleRepo domain.RoleRepository, auditLogRepo domain.AuditLogRepository) domain.RoleService {
	return &roleService{roleRepo: roleRepo, auditLogRepo: auditLogRepo}
}

func (s *roleService) List(ctx context.Context, filter domain.RoleFilter) ([]domain.Role, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return s.roleRepo.List(ctx, filter)
}

func (s *roleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	permissions, err := s.roleRepo.GetRolePermissions(ctx, id)
	if err != nil {
		return nil, err
	}

	role.Permissions = permissions
	return role, nil
}

func (s *roleService) Create(ctx context.Context, input domain.CreateRoleInput) (*domain.Role, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.InvalidInput(domain.CodeRoleInvalidData, "role name is required").WithField("name")
	}

	permissions, err := buildPermissions(input.Permissions)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	role := &domain.Role{
		ID:          id,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		Permissions: permissions,
	}
	// A duplicate name is rejected by the unique index; the repository writes both in one transaction.
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func buildPermissions(input []domain.CreateRoleMenuPermissionInput) ([]domain.RoleMenuPermission, error) {
	if len(input) == 0 {
		return nil, nil
	}

	out := make([]domain.RoleMenuPermission, 0, len(input))
	seen := make(map[uuid.UUID]struct{}, len(input))
	for _, p := range input {
		if p.MenuID == uuid.Nil {
			return nil, domain.InvalidInput(domain.CodeRoleInvalidMenu, "menu_id must not be empty").WithField("permissions")
		}
		if _, duplicate := seen[p.MenuID]; duplicate {
			return nil, domain.InvalidInput(domain.CodeRoleInvalidMenu, fmt.Sprintf("menu %s is listed twice", p.MenuID)).WithField("permissions")
		}
		seen[p.MenuID] = struct{}{}

		out = append(out, domain.RoleMenuPermission{
			MenuID:     p.MenuID,
			CanCreate:  p.CanCreate,
			CanRead:    p.CanRead,
			CanUpdate:  p.CanUpdate,
			CanDelete:  p.CanDelete,
			CanApprove: p.CanApprove,
		})
	}
	return out, nil
}

func (s *roleService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateRoleInput) (*domain.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil && strings.TrimSpace(*input.Name) != role.Name {
		if role.IsSystem {
			return nil, domain.Forbidden(domain.CodeRoleSystemImmutable, "a built-in role cannot be renamed").WithField("name")
		}
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.InvalidInput(domain.CodeRoleInvalidData, "role name must not be empty").WithField("name")
		}
		role.Name = name
	}
	// Description stays editable on built-in roles; only the name is fixed.
	if input.Description != nil {
		role.Description = strings.TrimSpace(*input.Description)
	}

	// nil means the request left permissions alone; an empty slice clears them.
	var permissions []domain.RoleMenuPermission
	if input.Permissions != nil {
		permissions, err = buildPermissions(*input.Permissions)
		if err != nil {
			return nil, err
		}
		if permissions == nil {
			permissions = []domain.RoleMenuPermission{}
		}
	}

	if err := s.roleRepo.Update(ctx, role, permissions); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Delete(ctx context.Context, id uuid.UUID) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Roles are soft-deleted, so ON DELETE RESTRICT never fires and the guard lives here.
	if role.IsSystem {
		return domain.Forbidden(domain.CodeRoleSystemImmutable, "a built-in role cannot be deleted")
	}
	return s.roleRepo.Delete(ctx, id)
}

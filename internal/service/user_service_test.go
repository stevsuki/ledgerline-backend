package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/mocks"
	"github.com/stevensuki/ledgerline-backend/internal/service"
)

func TestUserService_Create(t *testing.T) {
	t.Parallel()

	input := domain.CreateUserInput{
		Email:    "Budi@Example.com",
		FullName: "Budi Santoso",
		Password: "Rahasia123!",
	}

	// Table-driven test: one table of cases, one execution loop.
	tests := []struct {
		name      string
		setupMock func(*mocks.UserRepository, *mocks.PasswordHasher)
		wantErr   error
		assert    func(*testing.T, *domain.User)
	}{
		{
			name: "creates a user with the default role",
			setupMock: func(repo *mocks.UserRepository, hasher *mocks.PasswordHasher) {
				repo.On("ExistsByEmail", mock.Anything, "budi@example.com").Return(false, nil)
				hasher.On("Hash", input.Password).Return("hashed", nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
			},
			assert: func(t *testing.T, user *domain.User) {
				assert.Equal(t, "budi@example.com", user.Email, "email must be normalized to lowercase")
				assert.Equal(t, domain.RoleIDUser, user.RoleID)
				assert.Equal(t, "hashed", user.PasswordHash)
				assert.NotEqual(t, uuid.Nil, user.ID)
			},
		},
		{
			name: "fails because the email is already registered",
			setupMock: func(repo *mocks.UserRepository, _ *mocks.PasswordHasher) {
				repo.On("ExistsByEmail", mock.Anything, "budi@example.com").Return(true, nil)
			},
			wantErr: domain.ErrConflict,
		},
		{
			name: "fails because of a repository error",
			setupMock: func(repo *mocks.UserRepository, hasher *mocks.PasswordHasher) {
				repo.On("ExistsByEmail", mock.Anything, "budi@example.com").Return(false, nil)
				hasher.On("Hash", input.Password).Return("hashed", nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(errors.New("db down"))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mocks.UserRepository)
			hasher := new(mocks.PasswordHasher)
			tt.setupMock(repo, hasher)

			user, err := service.NewUserService(repo, hasher, new(mocks.AuditLogRepository)).Create(context.Background(), input)

			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)
			case tt.assert != nil:
				require.NoError(t, err)
				tt.assert(t, user)
			default:
				require.Error(t, err)
			}
			repo.AssertExpectations(t)
			hasher.AssertExpectations(t)
		})
	}
}

func TestUserService_Update(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	// Each subtest gets its own instance so they never mutate shared data.
	newExisting := func() *domain.User {
		return &domain.User{ID: id, Email: "budi@example.com", FullName: "Budi", RoleID: domain.RoleIDUser}
	}
	newName := "Budi Santoso"
	emptyRole := uuid.Nil

	t.Run("successfully changes the name", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		repo.On("GetByID", mock.Anything, id).Return(newExisting(), nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := service.NewUserService(repo, new(mocks.PasswordHasher), new(mocks.AuditLogRepository)).
			Update(context.Background(), id, domain.UpdateUserInput{FullName: &newName})

		require.NoError(t, err)
		assert.Equal(t, newName, user.FullName)
		repo.AssertExpectations(t)
	})

	t.Run("successfully changes the role", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		repo.On("GetByID", mock.Anything, id).Return(newExisting(), nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

		adminRole := domain.RoleIDAdmin
		user, err := service.NewUserService(repo, new(mocks.PasswordHasher), new(mocks.AuditLogRepository)).
			Update(context.Background(), id, domain.UpdateUserInput{RoleID: &adminRole})

		require.NoError(t, err)
		assert.Equal(t, domain.RoleIDAdmin, user.RoleID)
		repo.AssertExpectations(t)
	})

	// An unknown role id is caught by the foreign key; an empty one never
	// reaches the database, so the service has to reject it itself.
	t.Run("fails because role_id is empty", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		repo.On("GetByID", mock.Anything, id).Return(newExisting(), nil)

		_, err := service.NewUserService(repo, new(mocks.PasswordHasher), new(mocks.AuditLogRepository)).
			Update(context.Background(), id, domain.UpdateUserInput{RoleID: &emptyRole})

		require.ErrorIs(t, err, domain.ErrInvalidInput)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("fails because the user is not found", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		repo.On("GetByID", mock.Anything, id).Return(nil, domain.ErrNotFound)

		_, err := service.NewUserService(repo, new(mocks.PasswordHasher), new(mocks.AuditLogRepository)).
			Update(context.Background(), id, domain.UpdateUserInput{FullName: &newName})

		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/mocks"
	"github.com/stevensuki/ledgerline-backend/internal/service"
)

func TestCategoryService_Create(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name      string
		input     domain.CreateCategoryInput
		setupMock func(*mocks.CategoryRepository)
		wantErr   error
	}{
		{
			name:  "succeeds",
			input: domain.CreateCategoryInput{Name: "  Gaji  ", Type: domain.CategoryTypeIncome},
			setupMock: func(repo *mocks.CategoryRepository) {
				repo.On("ExistsByName", mock.Anything, userID, "Gaji").Return(false, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Category")).Return(nil)
			},
		},
		{
			name:      "an invalid type is rejected before reaching the repository",
			input:     domain.CreateCategoryInput{Name: "Gaji", Type: domain.CategoryType("invest")},
			setupMock: func(*mocks.CategoryRepository) {},
			wantErr:   domain.ErrInvalidInput,
		},
		{
			name:  "a duplicate name is rejected",
			input: domain.CreateCategoryInput{Name: "Gaji", Type: domain.CategoryTypeIncome},
			setupMock: func(repo *mocks.CategoryRepository) {
				repo.On("ExistsByName", mock.Anything, userID, "Gaji").Return(true, nil)
			},
			wantErr: domain.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mocks.CategoryRepository)
			tt.setupMock(repo)

			category, err := service.NewCategoryService(repo).Create(context.Background(), userID, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, category)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "Gaji", category.Name, "whitespace in the name must be trimmed")
				assert.Equal(t, userID, category.UserID)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_Update(t *testing.T) {
	t.Parallel()

	userID, id := uuid.New(), uuid.New()
	// Each subtest gets its own instance so they never mutate shared data.
	newExisting := func() *domain.Category {
		return &domain.Category{ID: id, UserID: userID, Name: "Gaji", Type: domain.CategoryTypeIncome}
	}
	newName := "Gaji Bulanan"

	t.Run("an unchanged name does not trigger the duplicate check", func(t *testing.T) {
		t.Parallel()

		sameName := "gaji"
		repo := new(mocks.CategoryRepository)
		repo.On("GetByID", mock.Anything, id, userID).Return(newExisting(), nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Category")).Return(nil)

		_, err := service.NewCategoryService(repo).
			Update(context.Background(), userID, id, domain.UpdateCategoryInput{Name: &sameName})

		require.NoError(t, err)
		repo.AssertNotCalled(t, "ExistsByName", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("a conflicting new name is rejected", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.CategoryRepository)
		repo.On("GetByID", mock.Anything, id, userID).Return(newExisting(), nil)
		repo.On("ExistsByName", mock.Anything, userID, newName).Return(true, nil)

		_, err := service.NewCategoryService(repo).
			Update(context.Background(), userID, id, domain.UpdateCategoryInput{Name: &newName})

		require.ErrorIs(t, err, domain.ErrConflict)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("a category owned by another user is treated as missing", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.CategoryRepository)
		repo.On("GetByID", mock.Anything, id, userID).Return(nil, domain.ErrNotFound)

		_, err := service.NewCategoryService(repo).
			Update(context.Background(), userID, id, domain.UpdateCategoryInput{Name: &newName})

		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

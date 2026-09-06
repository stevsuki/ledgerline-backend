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

func TestCategoryService_Delete(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	categoryID := uuid.New()

	tests := []struct {
		name      string
		setupMock func(*mocks.CategoryRepository, *mocks.BudgetRepository)
		wantCode  string
	}{
		{
			name: "deletes a category no budget limits",
			setupMock: func(categories *mocks.CategoryRepository, budgets *mocks.BudgetRepository) {
				budgets.On("ExistsByCategory", mock.Anything, categoryID, userID).Return(false, nil)
				categories.On("Delete", mock.Anything, categoryID, userID).Return(nil)
			},
		},
		{
			name: "refuses a category a budget still limits",
			setupMock: func(_ *mocks.CategoryRepository, budgets *mocks.BudgetRepository) {
				budgets.On("ExistsByCategory", mock.Anything, categoryID, userID).Return(true, nil)
			},
			wantCode: domain.CodeCategoryInUse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			categories := &mocks.CategoryRepository{}
			budgets := &mocks.BudgetRepository{}
			tt.setupMock(categories, budgets)

			err := service.NewCategoryService(categories, budgets).
				Delete(context.Background(), userID, categoryID)

			if tt.wantCode == "" {
				require.NoError(t, err)
				categories.AssertExpectations(t)
				return
			}

			var domainErr *domain.Error
			require.ErrorAs(t, err, &domainErr)
			assert.Equal(t, tt.wantCode, domainErr.Code)
			categories.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

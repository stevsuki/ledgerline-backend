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

	type repos struct {
		categories   *mocks.CategoryRepository
		budgets      *mocks.BudgetRepository
		transactions *mocks.TransactionRepository
	}

	tests := []struct {
		name      string
		setupMock func(repos)
		wantCode  string
	}{
		{
			name: "deletes a category nothing points at",
			setupMock: func(r repos) {
				r.budgets.On("ExistsByCategory", mock.Anything, categoryID, userID).Return(false, nil)
				r.transactions.On("ExistsByCategory", mock.Anything, categoryID, userID).Return(false, nil)
				r.categories.On("Delete", mock.Anything, categoryID, userID).Return(nil)
			},
		},
		{
			name: "refuses a category a budget still limits",
			setupMock: func(r repos) {
				r.budgets.On("ExistsByCategory", mock.Anything, categoryID, userID).Return(true, nil)
			},
			wantCode: domain.CodeCategoryInUse,
		},
		{
			// The half that used to slip through: nothing refused it, and the
			// dashboard then counted its spending in the total while leaving it
			// out of the ring.
			name: "refuses a category transactions are still filed under",
			setupMock: func(r repos) {
				r.budgets.On("ExistsByCategory", mock.Anything, categoryID, userID).Return(false, nil)
				r.transactions.On("ExistsByCategory", mock.Anything, categoryID, userID).Return(true, nil)
			},
			wantCode: domain.CodeCategoryInUse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			categories := &mocks.CategoryRepository{}
			budgets := &mocks.BudgetRepository{}
			transactions := &mocks.TransactionRepository{}
			tt.setupMock(repos{categories, budgets, transactions})

			err := service.NewCategoryService(categories, budgets, transactions).
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

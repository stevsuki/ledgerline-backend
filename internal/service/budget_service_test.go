package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/mocks"
)

func usage(name string, limit, spent int64, threshold int, isFixed bool) domain.BudgetUsage {
	return domain.BudgetUsage{
		BudgetID:              uuid.New(),
		CategoryID:            uuid.New(),
		CategoryName:          name,
		Currency:              domain.BaseCurrency,
		MonthlyLimit:          limit,
		Spent:                 spent,
		AlertThresholdPercent: threshold,
		IsFixed:               isFixed,
	}
}

func TestBudgetOverviewOf(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, time.August, 27, 10, 0, 0, 0, time.UTC)

	usages := []domain.BudgetUsage{
		usage("Housing", 6_000_000, 6_000_000, 100, true),
		usage("Food & Drink", 3_000_000, 2_700_000, 80, false),
		usage("Transport", 1_000_000, 1_200_000, 80, false),
	}

	overview := budgetOverviewOf(usages, now)

	assert.Equal(t, domain.BaseCurrency, overview.Currency)
	assert.Equal(t, int64(10_000_000), overview.TotalAllocated)
	assert.Equal(t, int64(9_900_000), overview.TotalSpent)
	assert.Equal(t, int64(100_000), overview.TotalLeft)
	assert.Equal(t, 3, overview.CategoryCount)
	assert.Equal(t, 99, overview.UsedPercent)
	assert.False(t, overview.IsOver)
	assert.Equal(t, 4, overview.DaysLeft)
	assert.Equal(t, 87, overview.CycleElapsedPercent)
}

func TestBudgetOverviewOf_SharesAlwaysCloseOn100(t *testing.T) {
	t.Parallel()

	usages := []domain.BudgetUsage{
		usage("A", 1_000_000, 0, 80, false),
		usage("B", 1_000_000, 0, 80, false),
		usage("C", 1_000_000, 0, 80, false),
	}

	shares := budgetOverviewOf(usages, time.Now()).Shares
	require.Len(t, shares, 3)

	total := 0
	for _, share := range shares {
		total += share.Percent
	}
	assert.Equal(t, 100, total)
}

func TestBudgetAttention(t *testing.T) {
	t.Parallel()

	usages := []domain.BudgetUsage{
		usage("Under", 1_000_000, 500_000, 80, false),
		usage("Near", 1_000_000, 850_000, 80, false),
		usage("Over", 1_000_000, 1_200_000, 80, false),
		usage("Fixed unpaid", 1_000_000, 0, 100, true),
		usage("Fixed paid", 1_000_000, 1_000_000, 100, true),
	}

	items := budgetAttention(usages)

	require.Len(t, items, 2)
	assert.Equal(t, "Over", items[0].CategoryName)
	assert.True(t, items[0].IsOver)
	assert.Equal(t, int64(-200_000), items[0].Remaining)
	assert.Equal(t, "Near", items[1].CategoryName)
	assert.False(t, items[1].IsOver)
}

func TestBudgetOverviewOf_NoBudgets(t *testing.T) {
	t.Parallel()

	overview := budgetOverviewOf(nil, time.Now())

	assert.Zero(t, overview.UsedPercent)
	assert.Empty(t, overview.Shares)
	assert.Empty(t, overview.NeedAttention)
}

func TestCheckBudgetLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		monthlyLimit int64
		threshold    int
		isFixed      bool
		wantCode     string
		wantField    string
	}{
		{name: "a plain budget", monthlyLimit: 10_000_000, threshold: 80},
		{name: "a fixed budget at 100", monthlyLimit: 10_000_000, threshold: 100, isFixed: true},
		{
			name: "a zero limit", monthlyLimit: 0, threshold: 80,
			wantCode: domain.CodeBudgetInvalidLimit, wantField: "monthly_limit",
		},
		{
			name: "a threshold over 100", monthlyLimit: 10_000_000, threshold: 101,
			wantCode: domain.CodeBudgetInvalidThreshold, wantField: "alert_threshold_percent",
		},
		{
			name: "a threshold under 1", monthlyLimit: 10_000_000, threshold: 0,
			wantCode: domain.CodeBudgetInvalidThreshold, wantField: "alert_threshold_percent",
		},
		{
			name: "a fixed budget alerting early", monthlyLimit: 10_000_000, threshold: 10, isFixed: true,
			wantCode: domain.CodeBudgetInvalidFixed, wantField: "alert_threshold_percent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := checkBudgetLimits(tt.monthlyLimit, tt.threshold, tt.isFixed)
			if tt.wantCode == "" {
				assert.NoError(t, err)
				return
			}

			var domainErr *domain.Error
			require.ErrorAs(t, err, &domainErr)
			assert.Equal(t, tt.wantCode, domainErr.Code)
			assert.Equal(t, tt.wantField, domainErr.Field)
		})
	}
}

func TestBudgetService_CreateChecksCategory(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	categoryID := uuid.New()
	input := domain.CreateBudgetInput{
		CategoryID:            categoryID,
		Currency:              domain.CurrencyIDR,
		MonthlyLimit:          10_000_000,
		AlertThresholdPercent: 80,
	}

	tests := []struct {
		name      string
		setupMock func(*mocks.BudgetRepository, *mocks.CategoryRepository)
		wantCode  string
	}{
		{
			name: "rejects a category the user does not own",
			setupMock: func(_ *mocks.BudgetRepository, categories *mocks.CategoryRepository) {
				categories.On("ResolveForUser", mock.Anything, userID, categoryID).
					Return(nil, domain.NotFound(domain.CodeCategoryNotFound, "category not found"))
			},
			wantCode: domain.CodeBudgetInvalidCategory,
		},
		{
			name: "rejects an income category",
			setupMock: func(_ *mocks.BudgetRepository, categories *mocks.CategoryRepository) {
				categories.On("ResolveForUser", mock.Anything, userID, categoryID).
					Return(&domain.Category{ID: categoryID, Type: domain.CategoryTypeIncome}, nil)
			},
			wantCode: domain.CodeBudgetInvalidCategory,
		},
		{
			name: "accepts an expense category the user owns",
			setupMock: func(budgets *mocks.BudgetRepository, categories *mocks.CategoryRepository) {
				categories.On("ResolveForUser", mock.Anything, userID, categoryID).
					Return(&domain.Category{ID: categoryID, Type: domain.CategoryTypeExpense}, nil)
				budgets.On("Create", mock.Anything, mock.AnythingOfType("*domain.Budget")).Return(nil)
				budgets.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID"), userID).
					Return(&domain.Budget{CategoryID: categoryID, CategoryName: "Food & Drink"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			budgets := &mocks.BudgetRepository{}
			categories := &mocks.CategoryRepository{}
			tt.setupMock(budgets, categories)

			budget, err := NewBudgetService(budgets, categories).
				Create(context.Background(), userID, input)

			if tt.wantCode == "" {
				require.NoError(t, err)
				assert.Equal(t, "Food & Drink", budget.CategoryName)
				budgets.AssertExpectations(t)
				return
			}

			var domainErr *domain.Error
			require.ErrorAs(t, err, &domainErr)
			assert.Equal(t, tt.wantCode, domainErr.Code)
			assert.Equal(t, "category_id", domainErr.Field)
			budgets.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}

func TestBudgetOverviewOf_LeavesOtherCurrenciesOut(t *testing.T) {
	t.Parallel()

	inDollars := usage("Travel", 2_000_000, 0, 80, false)
	inDollars.Currency = domain.CurrencyUSD

	overview := budgetOverviewOf([]domain.BudgetUsage{
		usage("Housing", 6_000_000, 0, 80, false),
		inDollars,
	}, time.Now())

	assert.Equal(t, int64(6_000_000), overview.TotalAllocated)
	assert.Equal(t, 1, overview.CategoryCount)
	assert.Equal(t, 1, overview.UncountedBudgets)
	require.Len(t, overview.Shares, 1)
	assert.Equal(t, 100, overview.Shares[0].Percent)
}

func TestBudgetOverviewOf_CarryWidensWhatIsLeft(t *testing.T) {
	t.Parallel()

	carried := usage("Food & Drink", 3_000_000, 3_400_000, 80, false)
	carried.CarriedOver = 800_000

	overview := budgetOverviewOf([]domain.BudgetUsage{carried}, time.Now())

	// Allocated stays the plan; only what is left counts the carry.
	assert.Equal(t, int64(3_000_000), overview.TotalAllocated)
	assert.Equal(t, int64(800_000), overview.TotalCarriedOver)
	assert.Equal(t, int64(400_000), overview.TotalLeft)
	assert.Equal(t, 89, overview.UsedPercent)
	assert.False(t, overview.IsOver)
}

func TestBudgetAttention_CarryKeepsABudgetQuiet(t *testing.T) {
	t.Parallel()

	carried := usage("Food & Drink", 1_000_000, 1_200_000, 80, false)
	carried.CarriedOver = 600_000

	assert.Empty(t, budgetAttention([]domain.BudgetUsage{carried}))

	// The same spend without a carry is over its limit.
	carried.CarriedOver = 0
	items := budgetAttention([]domain.BudgetUsage{carried})
	require.Len(t, items, 1)
	assert.True(t, items[0].IsOver)
}

package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type BudgetService struct {
	budgetRepo   domain.BudgetRepository
	categoryRepo domain.CategoryRepository
}

func NewBudgetService(
	budgetRepo domain.BudgetRepository,
	categoryRepo domain.CategoryRepository,
) domain.BudgetService {
	return &BudgetService{budgetRepo: budgetRepo, categoryRepo: categoryRepo}
}

func (s *BudgetService) checkCategory(ctx context.Context, userID, categoryID uuid.UUID) error {
	category, err := s.categoryRepo.GetByID(ctx, categoryID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.InvalidInput(domain.CodeBudgetInvalidCategory,
				"category_id does not refer to one of your categories").WithField("category_id")
		}
		return err
	}

	if category.Type != domain.CategoryTypeExpense {
		return domain.InvalidInput(domain.CodeBudgetInvalidCategory,
			"a budget can only limit an expense category").WithField("category_id")
	}
	return nil
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	return s.budgetRepo.List(ctx, userID)
}

func (s *BudgetService) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Budget, error) {
	return s.budgetRepo.GetByID(ctx, id, userID)
}

func checkBudgetLimits(monthlyLimit int64, threshold int, isFixed bool) error {
	if monthlyLimit <= 0 {
		return domain.InvalidInput(domain.CodeBudgetInvalidLimit,
			"monthly_limit must be greater than 0").WithField("monthly_limit")
	}

	if threshold < domain.MinAlertThresholdPercent || threshold > domain.MaxAlertThresholdPercent {
		return domain.InvalidInput(domain.CodeBudgetInvalidThreshold,
			fmt.Sprintf("alert_threshold_percent must be between %d and %d",
				domain.MinAlertThresholdPercent, domain.MaxAlertThresholdPercent)).
			WithField("alert_threshold_percent")
	}

	// A fixed commitment lands on its whole limit in one payment, so a warning
	// before that would fire every month and mean nothing.
	if isFixed && threshold != domain.MaxAlertThresholdPercent {
		return domain.InvalidInput(domain.CodeBudgetInvalidFixed,
			fmt.Sprintf("a fixed budget alerts at %d percent only", domain.MaxAlertThresholdPercent)).
			WithField("alert_threshold_percent")
	}
	return nil
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateBudgetInput) (*domain.Budget, error) {
	if !input.Currency.Valid() {
		return nil, domain.InvalidInput(domain.CodeBudgetInvalidCurrency, "budget currency must be IDR, USD, or SGD").WithField("currency")
	}

	if err := checkBudgetLimits(input.MonthlyLimit, input.AlertThresholdPercent, input.IsFixed); err != nil {
		return nil, err
	}

	if err := s.checkCategory(ctx, userID, input.CategoryID); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	budget := &domain.Budget{
		ID:                    id,
		UserID:                userID,
		CategoryID:            input.CategoryID,
		Currency:              input.Currency,
		MonthlyLimit:          input.MonthlyLimit,
		AlertThresholdPercent: input.AlertThresholdPercent,
		IsFixed:               input.IsFixed,
		Rollover:              input.Rollover,
	}

	if err := s.budgetRepo.Create(ctx, budget); err != nil {
		return nil, err
	}

	// Read back: the category name and icon come from the join, not from the write.
	return s.budgetRepo.GetByID(ctx, budget.ID, userID)
}

func (s *BudgetService) Update(ctx context.Context, userID, id uuid.UUID, input domain.UpdateBudgetInput) (*domain.Budget, error) {
	budget, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if input.CategoryID != nil {
		if err := s.checkCategory(ctx, userID, *input.CategoryID); err != nil {
			return nil, err
		}
		budget.CategoryID = *input.CategoryID
	}

	if input.Currency != nil {
		if !input.Currency.Valid() {
			return nil, domain.InvalidInput(domain.CodeBudgetInvalidCurrency, "budget currency must be IDR, USD, or SGD").WithField("currency")
		}
		budget.Currency = *input.Currency
	}

	if input.MonthlyLimit != nil {
		budget.MonthlyLimit = *input.MonthlyLimit
	}

	if input.AlertThresholdPercent != nil {
		budget.AlertThresholdPercent = *input.AlertThresholdPercent
	}

	if input.IsFixed != nil {
		budget.IsFixed = *input.IsFixed
	}

	if input.Rollover != nil {
		budget.Rollover = *input.Rollover
	}

	if err := checkBudgetLimits(budget.MonthlyLimit, budget.AlertThresholdPercent, budget.IsFixed); err != nil {
		return nil, err
	}

	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, err
	}
	return s.budgetRepo.GetByID(ctx, budget.ID, userID)
}

func (s *BudgetService) Overview(ctx context.Context, userID uuid.UUID) (domain.BudgetOverview, error) {
	usages, err := s.budgetRepo.Usage(ctx, userID)
	if err != nil {
		return domain.BudgetOverview{}, err
	}
	return budgetOverviewOf(usages, time.Now()), nil
}

func (s *BudgetService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.budgetRepo.Delete(ctx, id, userID)
}

func budgetOverviewOf(all []domain.BudgetUsage, now time.Time) domain.BudgetOverview {
	// Only the base currency reaches the headline; the rest are reported as a
	// count so the panel never quietly adds rupiah to dollars.
	usages := make([]domain.BudgetUsage, 0, len(all))
	uncounted := 0
	for _, usage := range all {
		if usage.Currency != domain.BaseCurrency {
			uncounted++
			continue
		}
		usages = append(usages, usage)
	}

	overview := domain.BudgetOverview{
		Currency:            domain.BaseCurrency,
		CategoryCount:       len(usages),
		UncountedBudgets:    uncounted,
		DaysLeft:            daysLeftInCycle(now),
		CycleElapsedPercent: cycleElapsedPercent(now),
		Shares:              budgetShares(usages),
		NeedAttention:       budgetAttention(usages),
	}

	for _, usage := range usages {
		overview.TotalAllocated += usage.MonthlyLimit
		overview.TotalSpent += usage.Spent
	}
	overview.TotalLeft = overview.TotalAllocated - overview.TotalSpent
	overview.UsedPercent = domain.PercentOf(overview.TotalSpent, overview.TotalAllocated)
	overview.IsOver = overview.TotalSpent > overview.TotalAllocated

	return overview
}

func budgetShares(usages []domain.BudgetUsage) []domain.BudgetShare {
	shares := make([]domain.BudgetShare, 0, len(usages))

	var allocated int64
	for _, usage := range usages {
		allocated += usage.MonthlyLimit
	}
	if allocated <= 0 {
		return shares
	}

	claimed := 0
	for i, usage := range usages {
		percent := domain.PercentOf(usage.MonthlyLimit, allocated)
		if i == len(usages)-1 {
			percent = domain.FullPercent - claimed
		}
		claimed += percent

		shares = append(shares, domain.BudgetShare{
			CategoryID:   usage.CategoryID,
			CategoryName: usage.CategoryName,
			Color:        usage.Color,
			Percent:      percent,
		})
	}
	return shares
}

func budgetAttention(usages []domain.BudgetUsage) []domain.BudgetAttention {
	items := make([]domain.BudgetAttention, 0, len(usages))

	for _, usage := range usages {
		used := domain.PercentOf(usage.Spent, usage.MonthlyLimit)
		isOver := usage.Spent > usage.MonthlyLimit
		if !isOver && (usage.IsFixed || used < usage.AlertThresholdPercent) {
			continue
		}

		items = append(items, domain.BudgetAttention{
			BudgetID:              usage.BudgetID,
			CategoryID:            usage.CategoryID,
			CategoryName:          usage.CategoryName,
			Icon:                  usage.Icon,
			Color:                 usage.Color,
			MonthlyLimit:          usage.MonthlyLimit,
			Spent:                 usage.Spent,
			Remaining:             usage.MonthlyLimit - usage.Spent,
			UsedPercent:           used,
			AlertThresholdPercent: usage.AlertThresholdPercent,
			IsFixed:               usage.IsFixed,
			IsOver:                isOver,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UsedPercent > items[j].UsedPercent
	})
	return items
}

// lastDayOfMonth: day 0 of next month is the last day of this one.
func lastDayOfMonth(now time.Time) int {
	return time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
}

func daysLeftInCycle(now time.Time) int {
	return lastDayOfMonth(now) - now.Day()
}

func cycleElapsedPercent(now time.Time) int {
	return domain.PercentOf(int64(now.Day()), int64(lastDayOfMonth(now)))
}

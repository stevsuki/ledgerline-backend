package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

const defaultRecurringNote = "Recurring payment"

type RecurringService struct {
	recurringRepo      domain.RecurringRepository
	categoryRepo       domain.CategoryRepository
	walletRepo         domain.WalletRepository
	transactionService domain.TransactionService
	txManager          domain.TxManager
	auditLogRepo       domain.AuditLogRepository
}

func NewRecurringService(
	recurringRepo domain.RecurringRepository,
	categoryRepo domain.CategoryRepository,
	walletRepo domain.WalletRepository,
	transactionService domain.TransactionService,
	txManager domain.TxManager,
	auditLogRepo domain.AuditLogRepository,
) domain.RecurringService {
	return &RecurringService{
		recurringRepo:      recurringRepo,
		categoryRepo:       categoryRepo,
		walletRepo:         walletRepo,
		transactionService: transactionService,
		txManager:          txManager,
		auditLogRepo:       auditLogRepo,
	}
}

func (s *RecurringService) List(ctx context.Context, userID uuid.UUID) ([]domain.Recurring, error) {
	return s.recurringRepo.List(ctx, userID)
}

func (s *RecurringService) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Recurring, error) {
	return s.recurringRepo.GetByID(ctx, id, userID)
}

func (s *RecurringService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateRecurringInput) (*domain.Recurring, error) {
	name := strings.TrimSpace(input.Name)
	if err := checkRecurring(name, input.Amount, input.Frequency, input.NextDueDate); err != nil {
		return nil, err
	}

	wallet, err := s.checkWallet(ctx, userID, input.WalletID)
	if err != nil {
		return nil, err
	}

	categoryID, err := s.checkCategory(ctx, userID, input.CategoryID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	recurring := &domain.Recurring{
		ID:          id,
		UserID:      userID,
		CategoryID:  categoryID,
		WalletID:    wallet.ID,
		Name:        name,
		Amount:      input.Amount,
		Currency:    wallet.Currency,
		Frequency:   input.Frequency,
		NextDueDate: dateOnly(input.NextDueDate),
		AnchorDay:   input.NextDueDate.Day(),
		IsPaused:    input.IsPaused,
	}

	if err := s.recurringRepo.Create(ctx, recurring); err != nil {
		return nil, err
	}

	created, err := s.recurringRepo.GetByID(ctx, recurring.ID, userID)
	if err != nil {
		return nil, err
	}

	recordAction(ctx, s.auditLogRepo, domain.AuditModuleRecurring, domain.AuditVerbCreate, recurringRecordOf(created))
	return created, nil
}

func (s *RecurringService) Update(ctx context.Context, userID, id uuid.UUID, input domain.UpdateRecurringInput) (*domain.Recurring, error) {
	recurring, err := s.recurringRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if input.WalletID != nil {
		wallet, err := s.checkWallet(ctx, userID, *input.WalletID)
		if err != nil {
			return nil, err
		}
		recurring.WalletID = wallet.ID
		recurring.Currency = wallet.Currency
	}

	if input.CategoryID != nil {
		categoryID, err := s.checkCategory(ctx, userID, *input.CategoryID)
		if err != nil {
			return nil, err
		}
		recurring.CategoryID = categoryID
	}

	if input.Name != nil {
		recurring.Name = strings.TrimSpace(*input.Name)
	}
	if input.Amount != nil {
		recurring.Amount = *input.Amount
	}
	if input.Frequency != nil {
		recurring.Frequency = *input.Frequency
	}
	if input.NextDueDate != nil {
		recurring.NextDueDate = dateOnly(*input.NextDueDate)
		recurring.AnchorDay = input.NextDueDate.Day()
	}
	if input.IsPaused != nil {
		recurring.IsPaused = *input.IsPaused
	}

	if err := checkRecurring(recurring.Name, recurring.Amount, recurring.Frequency, recurring.NextDueDate); err != nil {
		return nil, err
	}

	if err := s.recurringRepo.Update(ctx, recurring); err != nil {
		return nil, err
	}

	updated, err := s.recurringRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	recordAction(ctx, s.auditLogRepo, domain.AuditModuleRecurring, domain.AuditVerbUpdate, recurringRecordOf(updated))
	return updated, nil
}

func (s *RecurringService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	recurring, err := s.recurringRepo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if err := s.recurringRepo.Delete(ctx, id, userID); err != nil {
		return err
	}

	recordAction(ctx, s.auditLogRepo, domain.AuditModuleRecurring, domain.AuditVerbDelete, recurringRecordOf(recurring))
	return nil
}

func (s *RecurringService) Overview(ctx context.Context, userID uuid.UUID, today time.Time) (domain.RecurringOverview, error) {
	items, err := s.recurringRepo.List(ctx, userID)
	if err != nil {
		return domain.RecurringOverview{}, err
	}
	return recurringOverviewOf(items, today), nil
}

// Confirm records the payment and moves the schedule on, as one unit.
func (s *RecurringService) Confirm(ctx context.Context, userID, id uuid.UUID, input domain.ConfirmRecurringInput) (*domain.Recurring, error) {
	recurring, err := s.recurringRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if recurring.IsPaused {
		return nil, domain.Conflict(domain.CodeRecurringPaused, "resume this item before recording a payment for it")
	}

	amount := recurring.Amount
	if input.Amount != nil {
		amount = *input.Amount
	}
	if amount <= 0 {
		return nil, domain.InvalidInput(domain.CodeRecurringInvalidAmount, "amount must be greater than 0").WithField("amount")
	}

	occurredAt := time.Now()
	if input.OccurredAt != nil {
		occurredAt = *input.OccurredAt
	}

	note := strings.TrimSpace(input.Note)
	if note == "" {
		note = defaultRecurringNote
	}

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		_, err := s.transactionService.Create(ctx, userID, domain.CreateTransactionInput{
			CategoryID: recurring.CategoryID,
			WalletID:   recurring.WalletID,
			Name:       recurring.Name,
			Note:       note,
			Type:       domain.TransactionTypeExpense,
			Amount:     -amount,
			Currency:   recurring.Currency,
			OccurredAt: occurredAt,
		})
		if err != nil {
			return err
		}

		recurring.NextDueDate = recurring.Frequency.Next(recurring.NextDueDate, recurring.AnchorDay)
		return s.recurringRepo.Update(ctx, recurring)
	})
	if err != nil {
		return nil, err
	}

	return s.recurringRepo.GetByID(ctx, id, userID)
}

func (s *RecurringService) checkWallet(ctx context.Context, userID, walletID uuid.UUID) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.InvalidInput(domain.CodeRecurringInvalidWallet,
				"wallet_id does not refer to one of your wallets").WithField("wallet_id")
		}
		return nil, err
	}
	return wallet, nil
}

// checkCategory: a recurring item is a payment, so only an expense category can carry it.
func (s *RecurringService) checkCategory(ctx context.Context, userID, categoryID uuid.UUID) (uuid.UUID, error) {
	category, err := s.categoryRepo.ResolveForUser(ctx, userID, categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return uuid.Nil, domain.InvalidInput(domain.CodeRecurringInvalidCategory,
				"category_id does not refer to one of your categories").WithField("category_id")
		}
		return uuid.Nil, err
	}

	if category.Type != domain.CategoryTypeExpense {
		return uuid.Nil, domain.InvalidInput(domain.CodeRecurringInvalidCategory,
			"a recurring item can only be filed under an expense category").WithField("category_id")
	}
	return category.ID, nil
}

func checkRecurring(name string, amount int64, frequency domain.RecurringFrequency, nextDueDate time.Time) error {
	if name == "" {
		return domain.InvalidInput(domain.CodeRecurringInvalidData, "name is required").WithField("name")
	}
	if amount <= 0 {
		return domain.InvalidInput(domain.CodeRecurringInvalidAmount, "amount must be greater than 0").WithField("amount")
	}
	if !frequency.Valid() {
		return domain.InvalidInput(domain.CodeRecurringInvalidFrequency,
			"frequency must be weekly, monthly, quarterly, or annual").WithField("frequency")
	}
	if nextDueDate.IsZero() {
		return domain.InvalidInput(domain.CodeRecurringInvalidDate, "next_due_date is required").WithField("next_due_date")
	}
	return nil
}

func recurringRecordOf(recurring *domain.Recurring) domain.RecordDetail {
	detail := domain.NewRecordDetail(domain.RecordTypeRecurring, recurring.Name)
	detail.Note = fmt.Sprintf("%s %s", domain.Money{Amount: recurring.Amount, Currency: recurring.Currency}.String(), recurring.Frequency)
	return detail
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func recurringOverviewOf(items []domain.Recurring, today time.Time) domain.RecurringOverview {
	overview := domain.RecurringOverview{Currency: domain.BaseCurrency}
	today = dateOnly(today)
	dueSoonUntil := today.AddDate(0, 0, domain.RecurringDueSoonDays)

	for _, item := range items {
		if item.IsPaused {
			overview.PausedCount++
			continue
		}
		overview.ActiveCount++

		if item.NextDueDate.Before(today) {
			overview.OverdueCount++
		}
		if item.Currency != domain.BaseCurrency {
			overview.UncountedItems++
			continue
		}

		overview.MonthlyCommitment += item.Frequency.MonthlyAmount(item.Amount)
		if !item.NextDueDate.After(dueSoonUntil) {
			overview.DueSoonTotal += item.Amount
			overview.DueSoonCount++
		}
	}
	return overview
}

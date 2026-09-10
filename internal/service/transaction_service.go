package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type TransactionService struct {
	transactionRepo domain.TransactionRepository
	categoryRepo    domain.CategoryRepository
	walletRepo      domain.WalletRepository
}

func NewTransactionService(transactionRepo domain.TransactionRepository, categoryRepo domain.CategoryRepository, walletRepo domain.WalletRepository) domain.TransactionService {
	return &TransactionService{transactionRepo: transactionRepo, categoryRepo: categoryRepo, walletRepo: walletRepo}
}

func (t *TransactionService) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return t.transactionRepo.List(ctx, userID, filter)
}

func (t *TransactionService) Summary(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) (domain.TransactionSummary, error) {
	return t.transactionRepo.Summary(ctx, userID, filter)
}

func (t *TransactionService) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	return t.transactionRepo.GetByID(ctx, id, userID)
}

func (t *TransactionService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction name is required").WithField("name")
	}

	note := strings.TrimSpace(input.Note)
	if note == "" {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction note is required").WithField("note")
	}

	if !input.Type.Valid() {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidType, "transaction type must be debit or credit").WithField("type")
	}

	if err := checkAmountSign(input.Amount, input.Type); err != nil {
		return nil, err
	}

	_, err := t.walletRepo.GetByID(ctx, input.WalletID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidWallet,
				"wallet_id does not refer to one of your wallets").WithField("wallet_id")
		}
		return nil, err
	}

	_, err = t.categoryRepo.GetByID(ctx, input.CategoryID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidCategory,
				"category_id does not refer to one of your categories").WithField("category_id")
		}
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	transaction := &domain.Transaction{
		ID:         id,
		UserID:     userID,
		WalletID:   input.WalletID,
		CategoryID: input.CategoryID,
		Name:       name,
		Note:       note,
		Type:       input.Type,
		Amount:     input.Amount,
		Currency:   input.Currency,
		OccurredAt: input.OccurredAt,
	}

	if err := t.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (t *TransactionService) Update(ctx context.Context, id, userID uuid.UUID, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	transaction, err := t.transactionRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if input.WalletID != nil {
		if _, err := t.walletRepo.GetByID(ctx, *input.WalletID, userID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, domain.InvalidInput(domain.CodeTransactionInvalidWallet,
					"wallet_id does not refer to one of your wallets").WithField("wallet_id")
			}
			return nil, err
		}
		transaction.WalletID = *input.WalletID
	}

	if input.CategoryID != nil {
		if _, err := t.categoryRepo.GetByID(ctx, *input.CategoryID, userID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, domain.InvalidInput(domain.CodeTransactionInvalidCategory,
					"category_id does not refer to one of your categories").WithField("category_id")
			}
			return nil, err
		}
		transaction.CategoryID = *input.CategoryID
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction name must not be empty").WithField("name")
		}
		transaction.Name = name
	}

	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		if note == "" {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction note must not be empty").WithField("note")
		}
		transaction.Note = note
	}

	if input.Type != nil {
		if !input.Type.Valid() {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidType, "transaction type must be income or expense").WithField("type")
		}
		transaction.Type = *input.Type
	}

	if input.Currency != nil {
		if !input.Currency.Valid() {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction currency must be IDR, USD, or SGD").WithField("currency")
		}
		transaction.Currency = *input.Currency
	}

	if input.Amount != nil {
		transaction.Amount = *input.Amount
	}

	if input.OccurredAt != nil {
		transaction.OccurredAt = *input.OccurredAt
	}

	if err := checkAmountSign(transaction.Amount, transaction.Type); err != nil {
		return nil, err
	}

	if err := t.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (t *TransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return t.transactionRepo.Delete(ctx, id, userID)
}

func checkAmountSign(amount int64, transactionType domain.TransactionType) error {
	switch {
	case amount == 0:
		return domain.InvalidInput(domain.CodeTransactionInvalidAmount,
			"amount must not be zero").WithField("amount")
	case transactionType == domain.TransactionTypeIncome && amount < 0:
		return domain.InvalidInput(domain.CodeTransactionInvalidAmount,
			"income transaction must have positive amount").WithField("amount")
	case transactionType == domain.TransactionTypeExpense && amount > 0:
		return domain.InvalidInput(domain.CodeTransactionInvalidAmount,
			"expense transaction must have negative amount").WithField("amount")
	}
	return nil
}

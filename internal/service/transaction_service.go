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

	if !input.Currency.Valid() {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidCurrency,
			"transaction currency must be IDR, USD, or SGD").WithField("currency")
	}

	if err := checkAmountSign(input.Amount, input.Type); err != nil {
		return nil, err
	}

	wallet, err := t.walletRepo.GetByID(ctx, input.WalletID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidWallet,
				"wallet_id does not refer to one of your wallets").WithField("wallet_id")
		}
		return nil, err
	}
	if err := checkWalletCurrency(input.Currency, wallet); err != nil {
		return nil, err
	}

	category, err := t.categoryRepo.GetByID(ctx, input.CategoryID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidCategory,
				"category_id does not refer to one of your categories").WithField("category_id")
		}
		return nil, err
	}
	if err := checkCategoryType(input.Type, category); err != nil {
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

	// Kept for the pairing checks at the end, which only read once every field the
	// patch touches has been applied.
	var wallet *domain.Wallet
	var category *domain.Category

	if input.WalletID != nil {
		wallet, err = t.walletRepo.GetByID(ctx, *input.WalletID, userID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, domain.InvalidInput(domain.CodeTransactionInvalidWallet,
					"wallet_id does not refer to one of your wallets").WithField("wallet_id")
			}
			return nil, err
		}
		transaction.WalletID = *input.WalletID
	}

	if input.CategoryID != nil {
		category, err = t.categoryRepo.GetByID(ctx, *input.CategoryID, userID)
		if err != nil {
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
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidCurrency,
				"transaction currency must be IDR, USD, or SGD").WithField("currency")
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

	// A currency change has to answer to the wallet it sits in even when the wallet
	// itself was not touched, and the same the other way around — so whichever side
	// the patch left unread is read now.
	if input.Currency != nil && wallet == nil {
		if wallet, err = t.walletRepo.GetByID(ctx, transaction.WalletID, userID); err != nil {
			return nil, err
		}
	}
	if wallet != nil {
		if err := checkWalletCurrency(transaction.Currency, wallet); err != nil {
			return nil, err
		}
	}

	if input.Type != nil && category == nil {
		if category, err = t.categoryRepo.GetByID(ctx, transaction.CategoryID, userID); err != nil {
			return nil, err
		}
	}
	if category != nil {
		if err := checkCategoryType(transaction.Type, category); err != nil {
			return nil, err
		}
	}

	if err := t.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (t *TransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return t.transactionRepo.Delete(ctx, id, userID)
}

// checkWalletCurrency: a wallet holds one currency, and its balance sums only the rows
// kept in that currency — a row in another one would be money the wallet can never
// account for, so it is refused instead of filed where nothing will count it.
func checkWalletCurrency(currency domain.Currency, wallet *domain.Wallet) error {
	if currency == wallet.Currency {
		return nil
	}
	return domain.InvalidInput(domain.CodeTransactionCurrencyMismatch,
		fmt.Sprintf("%s holds %s, so this transaction cannot be kept in %s",
			wallet.Name, wallet.Currency, currency)).WithField("currency")
}

// checkCategoryType: income belongs under an income category and expense under an
// expense one, or the same row would read as earning on one screen and spending on
// the next — and a budget would measure against a category that cannot be spent on.
func checkCategoryType(transactionType domain.TransactionType, category *domain.Category) error {
	if transactionType.MatchesCategoryType(category.Type) {
		return nil
	}
	return domain.InvalidInput(domain.CodeTransactionCategoryMismatch,
		fmt.Sprintf("%s is a %s category, so it cannot carry a %s transaction",
			category.Name, category.Type, transactionType)).WithField("category_id")
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

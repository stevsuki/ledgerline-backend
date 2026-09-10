package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

func TestTransactionTypeMatchesCategoryType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		transactionType domain.TransactionType
		categoryType    string
		want            bool
	}{
		{name: "expense under an expense category", transactionType: domain.TransactionTypeExpense, categoryType: domain.CategoryTypeExpense, want: true},
		{name: "income under an income category", transactionType: domain.TransactionTypeIncome, categoryType: domain.CategoryTypeIncome, want: true},
		{name: "expense under an income category", transactionType: domain.TransactionTypeExpense, categoryType: domain.CategoryTypeIncome},
		{name: "income under an expense category", transactionType: domain.TransactionTypeIncome, categoryType: domain.CategoryTypeExpense},
		{name: "a category type nothing recognises", transactionType: domain.TransactionTypeExpense, categoryType: "transfer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.transactionType.MatchesCategoryType(tt.categoryType))
		})
	}
}

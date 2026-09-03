package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Money: an amount in the currency's smallest unit, integer because float64 loses cents.
type Money struct {
	Amount   int64    `json:"amount"`
	Currency Currency `json:"currency"`
}

// Currency: the currencies the app accepts, matching the CHECK constraints in the schema.
type Currency string

const (
	CurrencyIDR Currency = "IDR"
	CurrencyUSD Currency = "USD"
	CurrencySGD Currency = "SGD"
)

func (c Currency) Valid() bool {
	switch c {
	case CurrencyIDR, CurrencyUSD, CurrencySGD:
		return true
	}
	return false
}

// BaseCurrency: the one currency a headline total may be stated in.
const BaseCurrency = CurrencyIDR

func IDR(amount int64) Money { return Money{Amount: amount, Currency: CurrencyIDR} }

// String prints whole rupiah for IDR, two decimals for anything else.
func (m Money) String() string {
	if m.Currency == CurrencyIDR {
		return "Rp" + groupThousands(m.Amount)
	}
	return fmt.Sprintf("%s %s.%02d", m.Currency, groupThousands(m.Amount/100), abs(m.Amount%100))
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// groupThousands writes 2500000 as 2.500.000.
func groupThousands(n int64) string {
	digits := strconv.FormatInt(abs(n), 10)

	var b strings.Builder
	if n < 0 {
		b.WriteByte('-')
	}
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	return b.String()
}

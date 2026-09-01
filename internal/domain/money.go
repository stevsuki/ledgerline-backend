package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Money: an amount in the currency's smallest unit — rupiah for IDR, cents for
// USD. Integer on purpose: float64 cannot hold money exactly.
type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

const CurrencyIDR = "IDR"

func IDR(amount int64) Money { return Money{Amount: amount, Currency: CurrencyIDR} }

// String renders the amount for display. IDR has no minor unit, so it prints
// whole rupiah; anything else falls back to two decimals.
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

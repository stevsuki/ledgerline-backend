package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

func TestRecurringFrequencyNext(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		frequency domain.RecurringFrequency
		from      time.Time
		anchorDay int
		want      time.Time
	}{
		{"weekly", domain.RecurringWeekly, day(2026, time.August, 27), 27, day(2026, time.September, 3)},
		{"monthly", domain.RecurringMonthly, day(2026, time.September, 1), 1, day(2026, time.October, 1)},
		{"monthly clamps to a short month", domain.RecurringMonthly, day(2026, time.January, 31), 31, day(2026, time.February, 28)},
		{"monthly returns to its anchor", domain.RecurringMonthly, day(2026, time.February, 28), 31, day(2026, time.March, 31)},
		{"monthly over a year end", domain.RecurringMonthly, day(2026, time.December, 15), 15, day(2027, time.January, 15)},
		{"quarterly", domain.RecurringQuarterly, day(2026, time.November, 30), 30, day(2027, time.February, 28)},
		{"annual", domain.RecurringAnnual, day(2024, time.February, 29), 29, day(2025, time.February, 28)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.frequency.Next(tc.from, tc.anchorDay))
		})
	}
}

func TestRecurringOverviewOf(t *testing.T) {
	t.Parallel()

	today := day(2026, time.August, 27)
	item := func(amount int64, frequency domain.RecurringFrequency, due time.Time) domain.Recurring {
		return domain.Recurring{Amount: amount, Currency: domain.BaseCurrency, Frequency: frequency, NextDueDate: due}
	}

	paused := item(186_000, domain.RecurringMonthly, day(2026, time.September, 1))
	paused.IsPaused = true
	dollars := item(2_000, domain.RecurringMonthly, day(2026, time.August, 28))
	dollars.Currency = domain.CurrencyUSD

	overview := recurringOverviewOf([]domain.Recurring{
		item(4_500_000, domain.RecurringMonthly, day(2026, time.September, 1)),
		item(300_000, domain.RecurringQuarterly, day(2026, time.August, 20)),
		item(1_200_000, domain.RecurringAnnual, day(2026, time.December, 1)),
		item(120_000, domain.RecurringWeekly, day(2026, time.September, 3)),
		paused,
		dollars,
	}, today)

	assert.Equal(t, int64(4_500_000+100_000+100_000+520_000), overview.MonthlyCommitment)
	assert.Equal(t, int64(4_500_000+300_000+120_000), overview.DueSoonTotal)
	assert.Equal(t, 3, overview.DueSoonCount)
	assert.Equal(t, 1, overview.OverdueCount)
	assert.Equal(t, 5, overview.ActiveCount)
	assert.Equal(t, 1, overview.PausedCount)
	assert.Equal(t, 1, overview.UncountedItems)
}

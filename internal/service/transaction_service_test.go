package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

func TestTrendBuckets_Weekly(t *testing.T) {
	t.Parallel()

	// September 2026 starts on a Tuesday, so its first week starts in August.
	now := time.Date(2026, time.September, 10, 19, 0, 0, 0, time.UTC)

	buckets, bucket := trendBuckets(domain.TrendRangeWeekly, now)

	assert.Equal(t, domain.TrendBucketWeek, bucket)
	require.Len(t, buckets, 5)
	assert.Equal(t, time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC), buckets[0])
	assert.Equal(t, time.Date(2026, time.September, 28, 0, 0, 0, 0, time.UTC), buckets[4])

	for i, start := range buckets {
		assert.Equal(t, time.Monday, start.Weekday(), "bucket %d starts on a Monday", i)
	}
}

func TestTrendBuckets_Monthly(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 10, 19, 0, 0, 0, time.UTC)

	buckets, bucket := trendBuckets(domain.TrendRangeMonthly, now)

	assert.Equal(t, domain.TrendBucketMonth, bucket)
	require.Len(t, buckets, domain.TrendMonths)
	// Six months to September, this one included, earliest first.
	assert.Equal(t, time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC), buckets[0])
	assert.Equal(t, time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), buckets[5])
}

func TestFillTrend_KeepsEmptyPeriods(t *testing.T) {
	t.Parallel()

	buckets := []time.Time{
		time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
	}
	// Only August has rows, and the database answered in another zone.
	jakarta := time.FixedZone("WIB", 7*60*60)
	points := []domain.TransactionTrendPoint{
		{
			Start:    time.Date(2026, time.August, 1, 7, 0, 0, 0, jakarta),
			MoneyIn:  5_000_000,
			MoneyOut: -2_000_000,
		},
	}

	filled := fillTrend(buckets, domain.TrendBucketMonth, points)

	require.Len(t, filled, 3)
	assert.Zero(t, filled[0].MoneyIn)
	assert.Zero(t, filled[0].MoneyOut)
	assert.Equal(t, int64(5_000_000), filled[1].MoneyIn)
	assert.Equal(t, int64(-2_000_000), filled[1].MoneyOut)
	assert.Zero(t, filled[2].MoneyIn)

	// Every bar carries the period that was asked for, not the one the row came back with.
	for i, point := range filled {
		assert.Equal(t, buckets[i], point.Start)
	}
}

func TestMonthStart(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, time.September, 30, 23, 59, 59, 0, time.UTC)

	assert.Equal(t, time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), monthStart(at))
}

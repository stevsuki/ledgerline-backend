package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

func TestParams_Normalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		in          pagination.Params
		wantLimit   int
		wantOffset  int
		wantPageNum int
	}{
		{"empty values fall back to defaults", pagination.Params{}, 10, 0, 1},
		{"second page", pagination.Params{Page: 2, PerPage: 25}, 25, 25, 2},
		{"per_page above the cap is clamped", pagination.Params{Page: 1, PerPage: 500}, 100, 0, 1},
		{"negative values are normalized", pagination.Params{Page: -3, PerPage: -10}, 10, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantLimit, tt.in.Limit())
			assert.Equal(t, tt.wantOffset, tt.in.Offset())
			assert.Equal(t, tt.wantPageNum, tt.in.Normalize().Page)
		})
	}
}

func TestTotalPages(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 5, pagination.TotalPages(42, 10))
	assert.Equal(t, 1, pagination.TotalPages(10, 10))
	assert.Equal(t, 0, pagination.TotalPages(0, 10))
	assert.Equal(t, 0, pagination.TotalPages(10, 0))
}

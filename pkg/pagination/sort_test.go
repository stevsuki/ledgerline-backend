package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

var sortable = pagination.Sortable{
	Allowed: pagination.Whitelist{
		"name":       "name",
		"type":       "type",
		"created_at": "created_at",
		"updated_at": "updated_at",
	},
	Default:    "-created_at",
	TieBreaker: "id",
}

func TestSortable_OrderBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty falls back to default", "", "created_at DESC, id ASC"},
		{"whitespace counts as empty", "   ", "created_at DESC, id ASC"},
		{"ascending without a prefix", "name", "name ASC, id ASC"},
		{"descending with a minus prefix", "-name", "name DESC, id ASC"},
		{"multiple columns follow input order", "type,-created_at", "type ASC, created_at DESC, id ASC"},
		{"whitespace between columns is cleaned up", " type , -name ", "type ASC, name DESC, id ASC"},
		{"tie breaker is not duplicated", "-created_at", "created_at DESC, id ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sortable.OrderBy(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSortable_OrderBy_Rejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{"a column outside the whitelist", "password"},
		{"an sql injection attempt", "name; DROP TABLE users"},
		{"subquery", "(SELECT 1)"},
		{"a duplicated column", "name,-name"},
		{"exceeds the column limit", "name,type,created_at,updated_at"},
		{"a minus prefix with no column", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sortable.OrderBy(tt.raw)
			require.ErrorIs(t, err, pagination.ErrInvalidSort)
			assert.Empty(t, got)
		})
	}
}

// Without TieBreaker, the result is purely the requested columns.
func TestSortable_WithoutTieBreaker(t *testing.T) {
	t.Parallel()

	s := pagination.Sortable{Allowed: pagination.Whitelist{"name": "name"}, Default: "name"}

	got, err := s.OrderBy("")
	require.NoError(t, err)
	assert.Equal(t, "name ASC", got)
}

func TestSortable_Fields(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"created_at", "name", "type", "updated_at"}, sortable.Fields())
}

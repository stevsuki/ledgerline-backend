package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

func TestActorFrom(t *testing.T) {
	t.Parallel()

	actor := uuid.MustParse("6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b")

	tests := []struct {
		name string
		ctx  context.Context
		want *uuid.UUID
	}{
		{"a signed-in actor", domain.WithActor(context.Background(), actor), &actor},
		{"no actor at all", context.Background(), nil},
		{"the nil uuid counts as no actor", domain.WithActor(context.Background(), uuid.Nil), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := domain.ActorFrom(tt.ctx)
			switch {
			case tt.want == nil && got != nil:
				t.Errorf("ActorFrom() = %v, want nil", got)
			case tt.want != nil && (got == nil || *got != *tt.want):
				t.Errorf("ActorFrom() = %v, want %v", got, *tt.want)
			}
		})
	}
}

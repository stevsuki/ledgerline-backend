package domain

import (
	"context"

	"github.com/google/uuid"
)

// actorKey is private so the actor can only be put into a context here.
type actorKey struct{}

// WithActor stores the signed-in user as the actor for every write ctx reaches.
func WithActor(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, actorKey{}, userID)
}

// ActorFrom returns the signed-in user, nil on an unauthenticated request
// (register, password reset), which is exactly what the *_by columns store.
func ActorFrom(ctx context.Context) *uuid.UUID {
	id, ok := ctx.Value(actorKey{}).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return nil
	}
	return &id
}

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

// ActorFrom: the signed-in user, nil on an unauthenticated request.
func ActorFrom(ctx context.Context) *uuid.UUID {
	id, ok := ctx.Value(actorKey{}).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return nil
	}
	return &id
}

type auditActorKey struct{}

// AuditActor: who an audit entry is written for, as the request named them.
type AuditActor struct {
	UserID    uuid.UUID
	FullName  string
	RoleName  string
	IPAddress string
}

func WithAuditActor(ctx context.Context, actor AuditActor) context.Context {
	return context.WithValue(ctx, auditActorKey{}, actor)
}

func AuditActorFrom(ctx context.Context) (AuditActor, bool) {
	actor, ok := ctx.Value(auditActorKey{}).(AuditActor)
	return actor, ok && actor.UserID != uuid.Nil
}

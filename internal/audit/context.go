package audit

import "context"

type contextKey struct{}

// WithActor returns a child context carrying actorID for audit attribution.
func WithActor(ctx context.Context, actorID int64) context.Context {
	return context.WithValue(ctx, contextKey{}, actorID)
}

// ActorFromCtx extracts the actor ID set by WithActor.
// Returns nil if no actor is present.
func ActorFromCtx(ctx context.Context) *int64 {
	v, ok := ctx.Value(contextKey{}).(int64)
	if !ok {
		return nil
	}
	return &v
}

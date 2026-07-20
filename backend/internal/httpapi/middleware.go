package httpapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// fixedUserID is the single hardcoded user in v1 (no auth yet). It matches
// the seeded row in migrations/00002_seed.sql.
var fixedUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type contextKey int

const userIDContextKey contextKey = iota

// CurrentUserMiddleware injects the fixed v1 user id into the request
// context for every request. This is the single swap point: real
// authentication can later replace just this middleware (deriving the user
// id from a session/token instead of a constant) without touching any
// handler or service code, since they all read the user id exclusively via
// currentUserID.
func CurrentUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), userIDContextKey, fixedUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// currentUserID reads the current request's user id from context. Every
// handler and service call goes through this accessor rather than reading
// the fixed id directly.
func currentUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(userIDContextKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

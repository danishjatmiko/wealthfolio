package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/domain"
)

// sessionCookieName holds the opaque session token (see
// service.AuthService) — never any user data itself, so a leaked cookie is
// a revocable pointer, not readable information.
const sessionCookieName = "wf_session"

type contextKey int

const userContextKey contextKey = iota

// AuthMiddleware reads the session cookie, validates it via svc.Auth
// (checking both the sliding expiry and the absolute lifetime cap, and
// transparently extending the session if it's getting close to expiring),
// and injects the authenticated user into the request context. Missing,
// unknown, or expired sessions get a 401 JSON response — the frontend
// handles redirecting to the login page, not the API. This is the single
// swap point for auth: every handler/service reads the user exclusively
// via currentUserID/currentUser.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		user, newExpiry, err := h.svc.Auth.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			clearSessionCookie(w, h.cookieSecure)
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		if newExpiry != nil {
			setSessionCookie(w, cookie.Value, *newExpiry, h.cookieSecure)
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// currentUser reads the current request's authenticated user from context.
func currentUser(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(userContextKey).(domain.User)
	return u, ok
}

// currentUserID reads the current request's authenticated user id from
// context. Every handler and service call reads the user exclusively
// through this accessor.
func currentUserID(ctx context.Context) uuid.UUID {
	if u, ok := currentUser(ctx); ok {
		return u.ID
	}
	return uuid.Nil
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

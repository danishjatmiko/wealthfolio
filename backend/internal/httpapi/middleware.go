package httpapi

import (
	"context"
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/domain"
)

// maxRequestBodyBytes caps every request body (including the unauthenticated
// login route) well above any legitimate payload this API ever receives, so
// an oversized body fails fast instead of exhausting server memory.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// BodyLimit wraps every request body in http.MaxBytesReader so decodeJSON's
// json.Decoder can't be driven to allocate unbounded memory on a huge body.
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}

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

// ratesSyncTokenHeader carries the rates-sync script's shared secret. Not
// a cookie, since this script has no browser and no interactive session
// to hold one in.
const ratesSyncTokenHeader = "X-Rates-Sync-Token"

// RatesSyncOrAuthMiddleware is AuthMiddleware plus one extra way in: a
// request bearing X-Rates-Sync-Token matching the configured
// RATES_SYNC_TOKEN authenticates as the fixed account named by
// RATES_SYNC_EMAIL, no session cookie needed. This exists solely for the
// rates-sync cron script (backend/cmd/rates-sync), which runs unattended
// on the VPS and has no way to hold a browser-style session — everything
// else (every other route, and this same route via a normal cookie) is
// unaffected. Deliberately mounted on only /rates and /rates/latest (see
// router.go), not the whole authenticated route group, so a leaked token
// can only read/write gold rates, nothing else. Inert whenever
// RATES_SYNC_TOKEN isn't set (e.g. local dev), since the header check
// short-circuits to false and every request falls through to the normal
// AuthMiddleware path.
func (h *Handler) RatesSyncOrAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.ratesSyncToken != "" {
			given := r.Header.Get(ratesSyncTokenHeader)
			if given != "" && subtle.ConstantTimeCompare([]byte(given), []byte(h.ratesSyncToken)) == 1 {
				user, err := h.repos.Users.GetByEmail(r.Context(), h.ratesSyncEmail)
				if err != nil {
					writeError(w, http.StatusUnauthorized, "not authenticated")
					return
				}
				ctx := context.WithValue(r.Context(), userContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		h.AuthMiddleware(next).ServeHTTP(w, r)
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

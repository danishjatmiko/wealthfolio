package httpapi

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"wealthfolio/backend/internal/service"
)

// oauthStateCookieName holds a short-lived anti-CSRF token bound to the
// in-flight login attempt; scoped to the auth/google path prefix only.
const oauthStateCookieName = "wf_oauth_state"

// oauthPlatformCookieName remembers, across the redirect to Google and
// back, that this login was started by the Android app (?platform=android)
// rather than the web frontend — same lifecycle as oauthStateCookieName,
// just a second cookie alongside it, so googleCallback knows which of the
// two very different "hand the session back" mechanisms to use.
const oauthPlatformCookieName = "wf_oauth_platform"

// mobileAuthCallbackScheme is the custom URI scheme the Android app
// registers an intent-filter for for receiving its session token — see
// googleCallback's "android" branch below.
const mobileAuthCallbackScheme = "wealthfolio://auth-callback"

// googleLogin starts the Authorization Code flow: stash a random state
// value in a short-lived cookie, then redirect the browser to Google's
// consent screen with that same state. ?platform=android marks this as a
// mobile-app login (opened in a Chrome Custom Tab, not a full web
// navigation) so googleCallback hands the token back via a deep link
// instead of a cookie.
func (h *Handler) googleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := service.NewState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/api/v1/auth/google",
		MaxAge:   int((10 * time.Minute).Seconds()),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	if r.URL.Query().Get("platform") == "android" {
		http.SetCookie(w, &http.Cookie{
			Name:     oauthPlatformCookieName,
			Value:    "android",
			Path:     "/api/v1/auth/google",
			MaxAge:   int((10 * time.Minute).Seconds()),
			HttpOnly: true,
			Secure:   h.cookieSecure,
			SameSite: http.SameSiteLaxMode,
		})
	}
	http.Redirect(w, r, h.svc.Auth.AuthCodeURL(state), http.StatusFound)
}

// googleCallback completes the flow: verify the state matches what we
// handed out, exchange the code, upsert the user, and send the caller back
// with a session — as a cookie + web redirect normally, or as a
// wealthfolio://auth-callback deep link carrying the token when this login
// was started by the Android app. The registered Google redirect URI is
// unchanged either way; the branch happens entirely on our side.
func (h *Handler) googleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || stateCookie.Value == "" || r.URL.Query().Get("state") != stateCookie.Value {
		writeError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	clearStateCookie(w, h.cookieSecure)

	isMobile := false
	if platformCookie, err := r.Cookie(oauthPlatformCookieName); err == nil && platformCookie.Value == "android" {
		isMobile = true
		clearPlatformCookie(w, h.cookieSecure)
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing code")
		return
	}

	token, expiresAt, err := h.svc.Auth.HandleCallback(r.Context(), code)
	if err != nil {
		log.Printf("google oauth callback failed: %v", err)
		writeError(w, http.StatusUnauthorized, "google sign-in failed")
		return
	}

	if isMobile {
		dest := mobileAuthCallbackScheme + "?token=" + url.QueryEscape(token) + "&expires_at=" + url.QueryEscape(expiresAt.Format(time.RFC3339))
		http.Redirect(w, r, dest, http.StatusFound)
		return
	}

	setSessionCookie(w, token, expiresAt, h.cookieSecure)
	http.Redirect(w, r, h.appBaseURL, http.StatusFound)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// login handles email+password sign-in. Only the one account seeded by
// migrations/00006_password_auth.sql can currently succeed here — everyone
// else gets the same generic invalid-credentials response as a wrong
// password, whether or not their email exists.
//
// Sets the session cookie (what the web frontend relies on) and also
// returns the raw token in the JSON body, so the Android app — which has
// no cookie jar shared with this response — can pull it out and attach it
// as `Cookie: wf_session=<token>` on its own requests. AuthMiddleware
// doesn't care where that header came from, so no other backend change is
// needed for the app to use the exact same session mechanism as the web.
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	token, expiresAt, err := h.svc.Auth.PasswordLogin(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		log.Printf("password login failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	setSessionCookie(w, token, expiresAt, h.cookieSecure)
	writeJSON(w, http.StatusOK, loginResponse{Token: token, ExpiresAt: expiresAt})
}

func clearStateCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/api/v1/auth/google",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearPlatformCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthPlatformCookieName,
		Value:    "",
		Path:     "/api/v1/auth/google",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// logout deletes the session (if any) and clears the cookie. Deliberately
// not behind AuthMiddleware: it must still succeed (as a no-op) against a
// missing, unknown, or already-expired session so the frontend can always
// force-clear client state without first checking whether it's logged in.
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		if err := h.svc.Auth.Logout(r.Context(), cookie.Value); err != nil {
			log.Printf("logout: delete session: %v", err)
		}
	}
	clearSessionCookie(w, h.cookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

type meResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// me returns the authenticated user's profile. Sits behind AuthMiddleware,
// which is what actually enforces there being a valid session.
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, meResponse{
		ID:          user.ID.String(),
		Email:       derefOrEmpty(user.Email),
		DisplayName: user.DisplayName,
		AvatarURL:   derefOrEmpty(user.AvatarURL),
	})
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

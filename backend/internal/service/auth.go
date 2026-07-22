package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

const (
	// sessionSlidingWindow: a session stays valid for this long since it was
	// last used, refreshed automatically as the user keeps using the app.
	sessionSlidingWindow = 7 * 24 * time.Hour
	// sessionAbsoluteCap: no session survives longer than this from
	// creation, regardless of activity — forces a fresh Google login at
	// least this often, bounding how long a stolen cookie stays useful.
	sessionAbsoluteCap = 30 * 24 * time.Hour
	// sessionRefreshThreshold: the session is only extended (DB write +
	// cookie reissue) once less than this much of the sliding window
	// remains, so an active user triggers it roughly every few days
	// instead of on every single request.
	sessionRefreshThreshold = 3 * 24 * time.Hour
)

// ErrSessionExpired is returned by Authenticate when a session has hit its
// absolute lifetime cap. Distinct from db.ErrNotFound (unknown/already-
// lapsed-sliding-expiry token) only for clearer logging; both map to 401.
var ErrSessionExpired = errors.New("session expired")

// ErrInvalidCredentials is returned by PasswordLogin for every failure mode
// (unknown email, no password set on the account, wrong password) — kept
// deliberately generic so a failed attempt can't be used to enumerate which
// emails have an account or have password sign-in enabled.
var ErrInvalidCredentials = errors.New("invalid email or password")

// dummyPasswordHash is verified against when the requested email doesn't
// exist or has no password set, so PasswordLogin always pays the same
// Argon2id cost either way — otherwise a missing account would respond
// faster than a wrong password, a timing side-channel that leaks which
// emails have an account.
var dummyPasswordHash = mustHashPassword("wealthfolio-timing-guard-dummy-password")

func mustHashPassword(password string) string {
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	return hash
}

// AuthService drives the Google OAuth 2.0 Authorization Code flow and the
// resulting session lifecycle (issue, validate + slide, revoke).
type AuthService struct {
	repos    *db.Repos
	oauthCfg *oauth2.Config
}

func NewAuthService(repos *db.Repos, cfg config.Config) *AuthService {
	return &AuthService{
		repos: repos,
		oauthCfg: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

// AuthCodeURL builds the Google consent-screen URL to redirect the browser
// to, binding the given anti-CSRF state value. AccessTypeOnline is enough
// since the backend only ever needs a one-time profile fetch, not ongoing
// Google API access — so no refresh token is requested or stored.
func (s *AuthService) AuthCodeURL(state string) string {
	return s.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// NewState generates a cryptographically random anti-CSRF state value for
// the login handler to stash in a short-lived cookie and echo to Google.
func NewState() (string, error) {
	return randomToken(24)
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// HandleCallback exchanges an OAuth code for tokens, fetches the verified
// Google profile, upserts the corresponding user, and issues a new
// session. Returns the session token to set as a cookie plus its expiry.
func (s *AuthService) HandleCallback(ctx context.Context, code string) (token string, expiresAt time.Time, err error) {
	oauthToken, err := s.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("exchange code: %w", err)
	}

	profile, err := s.fetchProfile(ctx, oauthToken)
	if err != nil {
		return "", time.Time{}, err
	}

	user, err := s.upsertUser(ctx, profile)
	if err != nil {
		return "", time.Time{}, err
	}

	return s.issueSession(ctx, user.ID)
}

// fetchProfile calls Google's userinfo endpoint using the access token
// obtained directly from Google via the server-to-server code exchange
// above. That exchange is already authenticated with our client secret
// over TLS, so this profile is trustworthy without also having to parse
// and verify the ID token's JWT signature locally.
func (s *AuthService) fetchProfile(ctx context.Context, oauthToken *oauth2.Token) (db.GoogleProfile, error) {
	client := s.oauthCfg.Client(ctx, oauthToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return db.GoogleProfile{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return db.GoogleProfile{}, fmt.Errorf("fetch google profile: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return db.GoogleProfile{}, fmt.Errorf("google userinfo returned %d: %s", resp.StatusCode, body)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return db.GoogleProfile{}, fmt.Errorf("decode google profile: %w", err)
	}
	if info.Sub == "" || info.Email == "" {
		return db.GoogleProfile{}, errors.New("google profile missing sub or email")
	}
	return db.GoogleProfile{
		Sub:           info.Sub,
		Email:         info.Email,
		EmailVerified: info.EmailVerified,
		Name:          info.Name,
		Picture:       info.Picture,
	}, nil
}

// upsertUser finds the user already linked to this Google account, or
// creates one: the very first Google login ever claims the pre-auth seed
// user in place (carrying its existing data forward); every login after
// that gets a brand-new, empty user row (open sign-up).
func (s *AuthService) upsertUser(ctx context.Context, profile db.GoogleProfile) (domain.User, error) {
	existing, err := s.repos.Users.GetByGoogleSub(ctx, profile.Sub)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return domain.User{}, err
	}

	claimed, err := s.repos.Users.HasAnyGoogleUser(ctx)
	if err != nil {
		return domain.User{}, err
	}
	if !claimed {
		return s.repos.Users.ClaimSeedUser(ctx, profile)
	}
	return s.repos.Users.CreateUser(ctx, profile)
}

// PasswordLogin verifies an email+password pair and issues a session on
// success. Only accounts with a password_hash set can use this path —
// currently just one, seeded by migrations/00006_password_auth.sql —
// everyone else must sign in with Google.
func (s *AuthService) PasswordLogin(ctx context.Context, email, password string) (token string, expiresAt time.Time, err error) {
	user, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return "", time.Time{}, err
	}

	hash := dummyPasswordHash
	valid := err == nil && user.PasswordHash != nil
	if valid {
		hash = *user.PasswordHash
	}

	// Always run the full comparison, even against the dummy hash when the
	// account doesn't exist or has no password set (see dummyPasswordHash).
	ok, verifyErr := VerifyPassword(password, hash)
	if verifyErr != nil || !ok || !valid {
		return "", time.Time{}, ErrInvalidCredentials
	}

	return s.issueSession(ctx, user.ID)
}

func (s *AuthService) issueSession(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(sessionSlidingWindow)
	if err := s.repos.Sessions.Create(ctx, token, userID, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// Authenticate validates a session token: db.ErrNotFound if it's unknown or
// its sliding expiry already lapsed, ErrSessionExpired if it hit the
// absolute lifetime cap. On success, if the sliding window is close to
// running out, the session is extended (capped at the absolute deadline)
// and the new expiry is returned so the caller can reissue the cookie —
// nil means the existing cookie is still good as-is.
func (s *AuthService) Authenticate(ctx context.Context, token string) (domain.User, *time.Time, error) {
	sess, err := s.repos.Sessions.GetByToken(ctx, token)
	if err != nil {
		return domain.User{}, nil, err
	}

	absoluteDeadline := sess.CreatedAt.Add(sessionAbsoluteCap)
	now := time.Now()
	if now.After(absoluteDeadline) {
		_ = s.repos.Sessions.Delete(ctx, token)
		return domain.User{}, nil, ErrSessionExpired
	}

	if sess.ExpiresAt.Sub(now) > sessionRefreshThreshold {
		return sess.User, nil, nil
	}

	newExpiry := now.Add(sessionSlidingWindow)
	if newExpiry.After(absoluteDeadline) {
		newExpiry = absoluteDeadline
	}
	if err := s.repos.Sessions.Refresh(ctx, token, newExpiry); err != nil {
		return domain.User{}, nil, err
	}
	return sess.User, &newExpiry, nil
}

// Logout deletes the given session token. No error if it's already gone.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.repos.Sessions.Delete(ctx, token)
}

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// seedUserID is the pre-auth fixed user seeded by migrations/00002_seed.sql.
// The first-ever Google login claims it in place (see ClaimSeedUser) so
// every row already FK'd to this id (snapshots, holdings, targets, ...)
// survives the transition to real auth untouched.
var seedUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// UsersRepo manages users rows.
type UsersRepo struct {
	pool *pgxpool.Pool
}

func NewUsersRepo(pool *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{pool: pool}
}

const userSelectCols = `id, email, display_name, avatar_url, google_sub, password_hash, created_at`

func scanUser(row interface{ Scan(dest ...any) error }) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.GoogleSub, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

// GetByGoogleSub returns the user linked to the given Google subject id.
// ErrNotFound if no user has claimed that sub yet.
func (r *UsersRepo) GetByGoogleSub(ctx context.Context, sub string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userSelectCols+` FROM users WHERE google_sub = $1`, sub)
	u, err := scanUser(row)
	if err != nil {
		return domain.User{}, wrapNotFound(err)
	}
	return u, nil
}

// GetByEmail returns the user with the given email. ErrNotFound if none.
// Used only by the email/password login path — Google sign-in looks users
// up by google_sub instead.
func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userSelectCols+` FROM users WHERE email = $1`, email)
	u, err := scanUser(row)
	if err != nil {
		return domain.User{}, wrapNotFound(err)
	}
	return u, nil
}

// HasAnyGoogleUser reports whether any user has ever linked a Google
// account. AuthService uses this to decide whether a brand-new sign-in
// should claim the pre-auth seed user or create a fresh, empty one.
func (r *UsersRepo) HasAnyGoogleUser(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE google_sub IS NOT NULL)`).Scan(&exists)
	return exists, err
}

// GoogleProfile is the subset of a verified Google ID token needed to
// create or claim a user.
type GoogleProfile struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// ClaimSeedUser links the pre-auth seed user to a Google account by
// stamping its profile fields in place, preserving the row's id. Should
// only ever be called once, guarded by HasAnyGoogleUser.
func (r *UsersRepo) ClaimSeedUser(ctx context.Context, p GoogleProfile) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET google_sub = $1, email = $2, email_verified = $3, display_name = $4, avatar_url = $5
		WHERE id = $6
		RETURNING `+userSelectCols,
		p.Sub, p.Email, p.EmailVerified, displayNameOrEmail(p), nullableStr(p.Picture), seedUserID)
	u, err := scanUser(row)
	if err != nil {
		return domain.User{}, wrapNotFound(err)
	}
	return u, nil
}

// CreateUser inserts a brand-new user for a Google account signing in for
// the first time (after the seed user has already been claimed by someone
// else) — it starts with a completely empty workspace.
func (r *UsersRepo) CreateUser(ctx context.Context, p GoogleProfile) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (google_sub, email, email_verified, display_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+userSelectCols,
		p.Sub, p.Email, p.EmailVerified, displayNameOrEmail(p), nullableStr(p.Picture))
	return scanUser(row)
}

func displayNameOrEmail(p GoogleProfile) string {
	if p.Name != "" {
		return p.Name
	}
	return p.Email
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// SessionsRepo manages sessions rows.
type SessionsRepo struct {
	pool *pgxpool.Pool
}

func NewSessionsRepo(pool *pgxpool.Pool) *SessionsRepo {
	return &SessionsRepo{pool: pool}
}

// Session is a sessions row joined with its owning user.
type Session struct {
	Token     string
	User      domain.User
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Create inserts a new session row. token is the caller-generated opaque,
// cryptographically random cookie value (see service.generateSessionToken).
func (r *SessionsRepo) Create(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)`, token, userID, expiresAt)
	return err
}

// GetByToken returns the session and its owning user. ErrNotFound if the
// token doesn't exist or its sliding expiry has already lapsed; callers
// enforce the absolute lifetime cap themselves using CreatedAt (see
// AuthMiddleware), since that's a policy decision, not a storage one.
func (r *SessionsRepo) GetByToken(ctx context.Context, token string) (Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT s.token, s.created_at, s.expires_at,
			u.id, u.email, u.display_name, u.avatar_url, u.google_sub, u.created_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > now()`, token)

	var sess Session
	err := row.Scan(
		&sess.Token, &sess.CreatedAt, &sess.ExpiresAt,
		&sess.User.ID, &sess.User.Email, &sess.User.DisplayName, &sess.User.AvatarURL, &sess.User.GoogleSub, &sess.User.CreatedAt,
	)
	if err != nil {
		return Session{}, wrapNotFound(err)
	}
	return sess, nil
}

// Refresh extends a session's sliding expiry and bumps last_seen_at.
func (r *SessionsRepo) Refresh(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sessions SET expires_at = $1, last_seen_at = now() WHERE token = $2`, expiresAt, token)
	return err
}

// Delete removes a session (logout). No error if it doesn't exist, so
// logging out twice is a harmless no-op.
func (r *SessionsRepo) Delete(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

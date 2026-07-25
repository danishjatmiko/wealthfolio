package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// NotificationPatternsRepo manages notification_patterns rows.
type NotificationPatternsRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationPatternsRepo(pool *pgxpool.Pool) *NotificationPatternsRepo {
	return &NotificationPatternsRepo{pool: pool}
}

const notificationPatternSelectCols = `id, app_id, priority, field, regex, description, enabled, updated_at`

func scanNotificationPattern(row interface{ Scan(dest ...any) error }) (domain.NotificationPattern, error) {
	var p domain.NotificationPattern
	var description *string
	err := row.Scan(&p.ID, &p.AppID, &p.Priority, &p.Field, &p.Regex, &description, &p.Enabled, &p.UpdatedAt)
	if err != nil {
		return domain.NotificationPattern{}, err
	}
	if description != nil {
		p.Description = *description
	}
	return p, nil
}

// ListByAppSource returns every pattern (enabled or not) for one app,
// priority-ascending — the order Match tries them in.
func (r *NotificationPatternsRepo) ListByAppSource(ctx context.Context, source string) ([]domain.NotificationPattern, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+notificationPatternSelectCols+`
		FROM notification_patterns
		WHERE app_id = (SELECT id FROM notification_apps WHERE source = $1)
		ORDER BY priority`, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.NotificationPattern{}
	for rows.Next() {
		p, err := scanNotificationPattern(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListAllEnabled returns every enabled pattern for every enabled app, in
// one round trip — used to warm the whole in-memory cache at once. Rows
// come back source-then-priority ordered so callers can group by source
// as they scan.
func (r *NotificationPatternsRepo) ListAllEnabled(ctx context.Context) ([]domain.NotificationPattern, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.app_id, p.priority, p.field, p.regex, p.description, p.enabled, p.updated_at
		FROM notification_patterns p
		JOIN notification_apps a ON a.id = p.app_id
		WHERE p.enabled AND a.enabled
		ORDER BY a.source, p.priority`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.NotificationPattern{}
	for rows.Next() {
		p, err := scanNotificationPattern(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetByID returns one pattern. ErrNotFound if id doesn't exist.
func (r *NotificationPatternsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.NotificationPattern, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+notificationPatternSelectCols+` FROM notification_patterns WHERE id = $1`, id)
	p, err := scanNotificationPattern(row)
	if err != nil {
		return domain.NotificationPattern{}, wrapNotFound(err)
	}
	return p, nil
}

// MaxPriorityForSource returns the highest priority currently used by
// source's patterns, or 0 if it has none yet — callers use this to
// auto-assign a new pattern's priority when the caller doesn't specify one.
func (r *NotificationPatternsRepo) MaxPriorityForSource(ctx context.Context, source string) (int, error) {
	var max *int
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(priority) FROM notification_patterns
		WHERE app_id = (SELECT id FROM notification_apps WHERE source = $1)`, source).Scan(&max)
	if err != nil {
		return 0, err
	}
	if max == nil {
		return 0, nil
	}
	return *max, nil
}

// NotificationPatternWrite is the set of fields needed to create or update
// a pattern.
type NotificationPatternWrite struct {
	Priority    int
	Field       string
	Regex       string
	Description string
	Enabled     bool
}

// Create inserts a new pattern under source's app. ErrNotFound if source
// isn't a known catalog entry.
func (r *NotificationPatternsRepo) Create(ctx context.Context, source string, w NotificationPatternWrite) (domain.NotificationPattern, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO notification_patterns (app_id, priority, field, regex, description, enabled)
		VALUES ((SELECT id FROM notification_apps WHERE source = $1), $2, $3, $4, $5, $6)
		RETURNING `+notificationPatternSelectCols,
		source, w.Priority, w.Field, w.Regex, nullIfEmpty(w.Description), w.Enabled)
	p, err := scanNotificationPattern(row)
	if err != nil {
		return domain.NotificationPattern{}, wrapNotFound(err)
	}
	return p, nil
}

// Update overwrites an existing pattern's mutable fields. ErrNotFound if id
// doesn't exist.
func (r *NotificationPatternsRepo) Update(ctx context.Context, id uuid.UUID, w NotificationPatternWrite) (domain.NotificationPattern, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE notification_patterns
		SET priority = $1, field = $2, regex = $3, description = $4, enabled = $5, updated_at = now()
		WHERE id = $6
		RETURNING `+notificationPatternSelectCols,
		w.Priority, w.Field, w.Regex, nullIfEmpty(w.Description), w.Enabled, id)
	p, err := scanNotificationPattern(row)
	if err != nil {
		return domain.NotificationPattern{}, wrapNotFound(err)
	}
	return p, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

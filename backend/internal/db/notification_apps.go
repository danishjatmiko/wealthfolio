package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// NotificationAppsRepo manages notification_apps rows, joined with their
// notification_app_packages into domain.NotificationApp.PackageNames.
type NotificationAppsRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationAppsRepo(pool *pgxpool.Pool) *NotificationAppsRepo {
	return &NotificationAppsRepo{pool: pool}
}

// notificationAppSelect joins in package names via array_agg so every read
// path gets the full app in one round trip; COALESCE avoids a NULL array
// (as opposed to an empty one) when an app has zero packages.
const notificationAppSelect = `
	SELECT a.id, a.source, a.display_name, a.amount_thousand_sep, a.amount_decimal_sep, a.enabled, a.updated_at,
		COALESCE(array_agg(p.package_name) FILTER (WHERE p.package_name IS NOT NULL), '{}')
	FROM notification_apps a
	LEFT JOIN notification_app_packages p ON p.app_id = a.id
`

const notificationAppGroupBy = `GROUP BY a.id`

func scanNotificationApp(row interface{ Scan(dest ...any) error }) (domain.NotificationApp, error) {
	var a domain.NotificationApp
	err := row.Scan(&a.ID, &a.Source, &a.DisplayName, &a.AmountThousandSep, &a.AmountDecimalSep, &a.Enabled, &a.UpdatedAt, &a.PackageNames)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	return a, nil
}

// ListAll returns every catalog entry (enabled or not), source-ordered.
func (r *NotificationAppsRepo) ListAll(ctx context.Context) ([]domain.NotificationApp, error) {
	rows, err := r.pool.Query(ctx, notificationAppSelect+notificationAppGroupBy+` ORDER BY a.source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.NotificationApp{}
	for rows.Next() {
		a, err := scanNotificationApp(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetBySource returns one catalog entry. ErrNotFound if source isn't
// known.
func (r *NotificationAppsRepo) GetBySource(ctx context.Context, source string) (domain.NotificationApp, error) {
	row := r.pool.QueryRow(ctx, notificationAppSelect+`WHERE a.source = $1 `+notificationAppGroupBy, source)
	a, err := scanNotificationApp(row)
	if err != nil {
		return domain.NotificationApp{}, wrapNotFound(err)
	}
	return a, nil
}

// ExistsSource reports whether source is a known catalog entry, regardless
// of its enabled flag — a disabled app can still have a legitimate stale
// expense_source_mappings row pointing at it.
func (r *NotificationAppsRepo) ExistsSource(ctx context.Context, source string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM notification_apps WHERE source = $1)`, source).Scan(&exists)
	return exists, err
}

// NotificationAppWrite is the set of fields needed to create or fully
// replace a catalog entry, including its package names.
type NotificationAppWrite struct {
	Source            string
	DisplayName       string
	PackageNames      []string
	AmountThousandSep string
	AmountDecimalSep  string
	Enabled           bool
}

// Create inserts a new catalog entry plus its package rows in one
// transaction.
func (r *NotificationAppsRepo) Create(ctx context.Context, w NotificationAppWrite) (domain.NotificationApp, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO notification_apps (source, display_name, amount_thousand_sep, amount_decimal_sep, enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		w.Source, w.DisplayName, w.AmountThousandSep, w.AmountDecimalSep, w.Enabled).Scan(&id)
	if err != nil {
		return domain.NotificationApp{}, err
	}

	if err := replaceNotificationAppPackages(ctx, tx, id, w.PackageNames); err != nil {
		return domain.NotificationApp{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.NotificationApp{}, err
	}
	return r.GetBySource(ctx, w.Source)
}

// UpdateBySource fully replaces a catalog entry's mutable fields and
// package list. ErrNotFound if source doesn't exist.
func (r *NotificationAppsRepo) UpdateBySource(ctx context.Context, source string, w NotificationAppWrite) (domain.NotificationApp, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		UPDATE notification_apps
		SET display_name = $1, amount_thousand_sep = $2, amount_decimal_sep = $3, enabled = $4, updated_at = now()
		WHERE source = $5
		RETURNING id`,
		w.DisplayName, w.AmountThousandSep, w.AmountDecimalSep, w.Enabled, source).Scan(&id)
	if err != nil {
		return domain.NotificationApp{}, wrapNotFound(err)
	}

	if err := replaceNotificationAppPackages(ctx, tx, id, w.PackageNames); err != nil {
		return domain.NotificationApp{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.NotificationApp{}, err
	}
	return r.GetBySource(ctx, source)
}

func replaceNotificationAppPackages(ctx context.Context, tx pgx.Tx, appID uuid.UUID, packageNames []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM notification_app_packages WHERE app_id = $1`, appID); err != nil {
		return err
	}
	for _, pkg := range packageNames {
		if _, err := tx.Exec(ctx, `INSERT INTO notification_app_packages (app_id, package_name) VALUES ($1, $2)`, appID, pkg); err != nil {
			return err
		}
	}
	return nil
}

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"wealthfolio/backend/internal/domain"
)

// CategoriesRepo reads the (effectively static, seeded) categories table.
type CategoriesRepo struct {
	pool *pgxpool.Pool
}

func NewCategoriesRepo(pool *pgxpool.Pool) *CategoriesRepo {
	return &CategoriesRepo{pool: pool}
}

// List returns every category ordered by sort_order.
func (r *CategoriesRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, key, label, color_oklch, kind, price_linked, sort_order
		FROM categories
		ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Category{}
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Key, &c.Label, &c.ColorOKLCH, &c.Kind, &c.PriceLinked, &c.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetByID looks up a single category. Returns ErrNotFound if id doesn't exist.
func (r *CategoriesRepo) GetByID(ctx context.Context, id int16) (domain.Category, error) {
	var c domain.Category
	err := r.pool.QueryRow(ctx, `
		SELECT id, key, label, color_oklch, kind, price_linked, sort_order
		FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Key, &c.Label, &c.ColorOKLCH, &c.Kind, &c.PriceLinked, &c.SortOrder)
	if err != nil {
		return domain.Category{}, wrapNotFound(err)
	}
	return c, nil
}

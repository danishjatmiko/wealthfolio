package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
)

// BigExpensesService owns the Big Expense ledger: plain CRUD over
// big_expenses, plus a per-year summary (monthly buckets + category
// breakdown) for the dashboard section of the page.
type BigExpensesService struct {
	repos *db.Repos
}

func NewBigExpensesService(repos *db.Repos) *BigExpensesService {
	return &BigExpensesService{repos: repos}
}

// defaultBigExpenseCategory is filled in whenever an entry arrives with no
// category named, so every row always has one to bucket by in the donut.
const defaultBigExpenseCategory = "Other"

// bigExpenseCategoryPalette gives each category a stable chart color,
// cycled in alphabetical order (see Summary below). A dedicated 14-color
// palette rather than reusing expenseEnvelopePalette (8 colors) — the
// curated category list alone (BIG_EXPENSE_CATEGORIES on the frontend) has
// 13 entries plus "Other", so an 8-color palette would silently repeat
// colors across unrelated categories. Same OKLCH style as every other
// backend-owned palette (couponMonthPalette, expenseEnvelopePalette): hues
// step evenly around the wheel with alternating lightness so neighbouring
// ring slices stay distinguishable.
var bigExpenseCategoryPalette = []string{
	"oklch(0.66 0.10 15)",
	"oklch(0.58 0.10 41)",
	"oklch(0.66 0.10 67)",
	"oklch(0.58 0.10 93)",
	"oklch(0.66 0.10 119)",
	"oklch(0.58 0.10 145)",
	"oklch(0.66 0.10 171)",
	"oklch(0.58 0.10 197)",
	"oklch(0.66 0.10 223)",
	"oklch(0.58 0.10 249)",
	"oklch(0.66 0.10 275)",
	"oklch(0.58 0.10 301)",
	"oklch(0.66 0.10 327)",
	"oklch(0.58 0.10 353)",
}

// BigExpenseDTO is one ledger row — a flat re-declaration of the stored
// columns, house style.
type BigExpenseDTO struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	AmountIdr   int64       `json:"amount_idr"`
	ExpenseDate domain.Date `json:"expense_date"`
	Category    string      `json:"category"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func bigExpenseDTO(e domain.BigExpense) BigExpenseDTO {
	return BigExpenseDTO{
		ID: e.ID, Name: e.Name, AmountIdr: e.AmountIdr, ExpenseDate: e.ExpenseDate,
		Category: e.Category, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

// BigExpenseMonthDTO is one of the twelve calendar buckets for a reference
// year, with the entries landing in it nested so the UI can drill down
// without a second request.
type BigExpenseMonthDTO struct {
	Month     int             `json:"month"`
	Label     string          `json:"label"`
	AmountIdr int64           `json:"amount_idr"`
	Percent   float64         `json:"percent"`
	Entries   []BigExpenseDTO `json:"entries"`
}

// BigExpenseCategoryDTO is one category's slice of a reference year.
type BigExpenseCategoryDTO struct {
	Category   string  `json:"category"`
	ColorOKLCH string  `json:"color_oklch"`
	AmountIdr  int64   `json:"amount_idr"`
	Percent    float64 `json:"percent"`
	Count      int     `json:"count"`
}

// BigExpenseSummaryDTO is the ledger page's dashboard payload: the
// requested year bucketed by month and by category.
type BigExpenseSummaryDTO struct {
	ReferenceYear int                     `json:"reference_year"`
	Months        []BigExpenseMonthDTO    `json:"months"`
	Categories    []BigExpenseCategoryDTO `json:"categories"`
	TotalIdr      int64                   `json:"total_idr"`
	EntriesCount  int                     `json:"entries_count"`
}

// List returns every big expense for the user, newest first.
func (s *BigExpensesService) List(ctx context.Context, userID uuid.UUID) ([]BigExpenseDTO, error) {
	expenses, err := s.repos.BigExpenses.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]BigExpenseDTO, 0, len(expenses))
	for _, e := range expenses {
		out = append(out, bigExpenseDTO(e))
	}
	return out, nil
}

// BigExpenseRequest is the parsed POST/PUT body for a ledger write.
type BigExpenseRequest struct {
	Name        string
	AmountIdr   int64
	ExpenseDate domain.Date
	Category    string
}

func (r BigExpenseRequest) validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if r.AmountIdr <= 0 {
		return fmt.Errorf("%w: amount_idr must be greater than 0", ErrInvalidInput)
	}
	if r.ExpenseDate.Time.IsZero() {
		return fmt.Errorf("%w: expense_date is required", ErrInvalidInput)
	}
	return nil
}

func (r BigExpenseRequest) toWrite() db.BigExpenseWrite {
	category := r.Category
	if category == "" {
		category = defaultBigExpenseCategory
	}
	return db.BigExpenseWrite{
		Name: r.Name, AmountIdr: r.AmountIdr, ExpenseDate: r.ExpenseDate, Category: category,
	}
}

func (s *BigExpensesService) Create(ctx context.Context, userID uuid.UUID, req BigExpenseRequest) (BigExpenseDTO, error) {
	if err := req.validate(); err != nil {
		return BigExpenseDTO{}, err
	}
	e, err := s.repos.BigExpenses.Create(ctx, userID, req.toWrite())
	if err != nil {
		return BigExpenseDTO{}, err
	}
	return bigExpenseDTO(e), nil
}

func (s *BigExpensesService) Update(ctx context.Context, userID, id uuid.UUID, req BigExpenseRequest) (BigExpenseDTO, error) {
	if err := req.validate(); err != nil {
		return BigExpenseDTO{}, err
	}
	e, err := s.repos.BigExpenses.Update(ctx, userID, id, req.toWrite())
	if err != nil {
		return BigExpenseDTO{}, err
	}
	return bigExpenseDTO(e), nil
}

func (s *BigExpensesService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repos.BigExpenses.Delete(ctx, userID, id)
}

// Summary buckets the requested year's entries into twelve months and by
// category. Always returns exactly twelve months — a month with no entries
// still needs a row in the UI.
func (s *BigExpensesService) Summary(ctx context.Context, userID uuid.UUID, year int) (BigExpenseSummaryDTO, error) {
	expenses, err := s.repos.BigExpenses.List(ctx, userID)
	if err != nil {
		return BigExpenseSummaryDTO{}, err
	}

	out := BigExpenseSummaryDTO{
		ReferenceYear: year,
		Months:        make([]BigExpenseMonthDTO, 12),
	}
	for i := range out.Months {
		out.Months[i] = BigExpenseMonthDTO{
			Month:   i + 1,
			Label:   time.Month(i + 1).String(),
			Entries: []BigExpenseDTO{},
		}
	}

	type catAgg struct {
		amountIdr int64
		count     int
	}
	catTotals := map[string]*catAgg{}
	catOrder := []string{}

	for _, e := range expenses {
		if e.ExpenseDate.Time.Year() != year {
			continue
		}

		m := int(e.ExpenseDate.Time.Month()) - 1
		out.Months[m].AmountIdr += e.AmountIdr
		out.Months[m].Entries = append(out.Months[m].Entries, bigExpenseDTO(e))
		out.TotalIdr += e.AmountIdr
		out.EntriesCount++

		cat := e.Category
		if cat == "" {
			cat = defaultBigExpenseCategory
		}
		agg, ok := catTotals[cat]
		if !ok {
			agg = &catAgg{}
			catTotals[cat] = agg
			catOrder = append(catOrder, cat)
		}
		agg.amountIdr += e.AmountIdr
		agg.count++
	}

	for i := range out.Months {
		out.Months[i].Percent = percentOf(float64(out.Months[i].AmountIdr), float64(out.TotalIdr))
	}

	// Alphabetical before assigning palette indices, so a category's color
	// stays stable across requests regardless of what order its entries
	// happen to appear in.
	sort.Strings(catOrder)
	out.Categories = make([]BigExpenseCategoryDTO, 0, len(catOrder))
	for i, name := range catOrder {
		agg := catTotals[name]
		out.Categories = append(out.Categories, BigExpenseCategoryDTO{
			Category:   name,
			ColorOKLCH: bigExpenseCategoryPalette[i%len(bigExpenseCategoryPalette)],
			AmountIdr:  agg.amountIdr,
			Percent:    percentOf(float64(agg.amountIdr), float64(out.TotalIdr)),
			Count:      agg.count,
		})
	}
	sort.Slice(out.Categories, func(i, j int) bool {
		return out.Categories[i].AmountIdr > out.Categories[j].AmountIdr
	})

	return out, nil
}

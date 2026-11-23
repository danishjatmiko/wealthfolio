package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
	"wealthfolio/backend/internal/service"
	"wealthfolio/backend/migrations"
)

// Run against a disposable database; migrations include the app's seed data.
// CASHFLOW_TEST_DATABASE_URL=postgres:///wealthfolio_cashflow_test go test ./internal/httpapi -run TestCashflowIntegration
func TestCashflowIntegration(t *testing.T) {
	dsn := os.Getenv("CASHFLOW_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CASHFLOW_TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	if err := db.RunMigrations(dsn, migrations.FS); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repos := db.NewRepos(pool)
	cfg := config.Config{AppBaseURL: "http://localhost", CORSOrigin: "http://localhost"}
	router := NewRouter(cfg, repos, service.NewServices(repos, cfg))
	makeUser := func() (uuid.UUID, string) {
		id := uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO users (id, display_name) VALUES ($1, 'Cashflow test')`, id); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id) })
		token := uuid.NewString()
		if err := repos.Sessions.Create(ctx, token, id, time.Now().Add(7*24*time.Hour)); err != nil {
			t.Fatal(err)
		}
		return id, token
	}
	owner, token := makeUser()
	other, otherToken := makeUser()
	request := func(method, path, body, cookie string, want int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: cookie})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != want {
			t.Fatalf("%s %s: status %d, want %d: %s", method, path, w.Code, want, w.Body)
		}
		return w
	}
	summary := func(cookie string) service.CashflowDTO {
		w := request("GET", "/cashflow?year=2026", "", cookie, 200)
		var result service.CashflowDTO
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}

	request("GET", "/cashflow?year=2026", "", "", 401)
	request("PUT", "/cashflow/active-income/2026-08", `{"amount_idr":1}`, "", 401)
	request("DELETE", "/cashflow/active-income/2026-08", "", "", 401)
	for _, path := range []string{"/cashflow?year=abc", "/cashflow?year=2025", "/cashflow?year=9999"} {
		request("GET", path, "", token, 400)
	}
	for _, body := range []string{`{}`, `null`, `{"amount_idr":null}`, `{"amount_idr":-1}`, `{"amount_idr":1.5}`, `{"amount_idr":9007199254740992}`} {
		request("PUT", "/cashflow/active-income/2026-08", body, token, 400)
	}
	request("PUT", "/cashflow/active-income/2026-07", `{"amount_idr":1}`, token, 400)
	request("PUT", "/cashflow/active-income/2026-08-01", `{"amount_idr":1}`, token, 400)
	if got := summary(token); len(got.Months) != 5 || got.Months[0].ActiveIncome != nil {
		t.Fatalf("new user should be empty: %+v", got)
	}

	request("PUT", "/cashflow/active-income/2026-08", `{"amount_idr":20000001,"note":"  Salary  "}`, token, 200)
	request("PUT", "/cashflow/active-income/2026-08", `{"amount_idr":25000001,"note":"  Updated salary  "}`, token, 200)
	request("PUT", "/cashflow/active-income/2026-09", `{"amount_idr":0}`, token, 200)
	request("PUT", "/cashflow/active-income/2026-08", `{"amount_idr":999000000}`, otherToken, 200)
	request("DELETE", "/cashflow/active-income/2026-09", "", otherToken, 404)

	date := func(raw string) domain.Date {
		d, err := domain.ParseDate(raw)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	period, err := repos.ExpensePeriods.Create(ctx, owner, date("2026-07-25"), date("2026-08-24"))
	if err != nil {
		t.Fatal(err)
	}
	env, err := repos.BudgetEnvelopes.Create(ctx, db.BudgetEnvelopeWrite{PeriodID: period.ID, Name: "Budget", CommittedAmountIdr: 30_000_000})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repos.FixedExpenses.Create(ctx, db.FixedExpenseWrite{PeriodID: period.ID, EnvelopeID: env.ID, Name: "Actual spend", AmountIdr: 10_000_000, ExpenseDate: date("2026-08-02")}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []uuid.UUID{owner, other} {
		amount := int64(2_000_000)
		if user == other {
			amount = 999_000_000
		}
		if _, err := repos.BigExpenses.Create(ctx, user, db.BigExpenseWrite{Name: "Purchase", ExpenseDate: date("2026-08-31"), AmountIdr: amount}); err != nil {
			t.Fatal(err)
		}
		var catID int16
		if err := pool.QueryRow(ctx, `SELECT id FROM categories WHERE kind = 'asset' ORDER BY id LIMIT 1`).Scan(&catID); err != nil {
			t.Fatal(err)
		}
		if _, err := repos.PassiveIncome.Create(ctx, user, db.PassiveIncomeWrite{CategoryID: catID, Name: "Dividend", ReceivedDate: date("2026-08-01"), AmountIdr: amount, IncomeType: "Dividen"}); err != nil {
			t.Fatal(err)
		}
	}

	got := summary(token)
	aug := got.Months[0]
	if aug.ActiveIncome == nil || aug.ActiveIncome.AmountIdr != 25_000_001 || aug.ActiveIncome.Note != "Updated salary" || aug.PassiveIncomeIdr != 2_000_000 || aug.ExpensesIdr != 10_000_000 || aug.BigExpensesIdr != 2_000_000 || aug.NetIdr != 15_000_001 {
		t.Fatalf("wrong totals or leaked other user's records: %+v", aug)
	}
	if got.Months[1].ActiveIncome == nil || got.Months[1].ActiveIncome.AmountIdr != 0 {
		t.Fatal("explicit zero must round-trip")
	}
	if err := repos.ExpensePeriods.Delete(ctx, owner, period.ID); err != nil {
		t.Fatal(err)
	}
	if aug = summary(token).Months[0]; aug.HasExpensePeriod || aug.ExpensesIdr != 0 || aug.NetIdr != 25_000_001 {
		t.Fatalf("source deletion not reflected: %+v", aug)
	}
	request("DELETE", "/cashflow/active-income/2026-08", "", token, 204)
	if aug = summary(token).Months[0]; aug.ActiveIncome != nil || aug.NetIdr != 0 {
		t.Fatalf("income deletion not reflected: %+v", aug)
	}
	if aug = summary(otherToken).Months[0]; aug.ActiveIncome == nil || aug.ActiveIncome.AmountIdr != 999_000_000 {
		t.Fatal("another user's income was modified")
	}
}

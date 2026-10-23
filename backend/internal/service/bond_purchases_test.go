package service

import (
	"errors"
	"testing"
	"time"

	"wealthfolio/backend/internal/domain"
)

func TestSummarizeByNameCarriesCategory(t *testing.T) {
	buy := mustDate(t, "2026-01-10")
	maturity := mustDate(t, "2048-05-15")
	lot := func(name, category string) domain.BondPurchase {
		return domain.BondPurchase{
			BondName: name, CategoryKey: category, InterestRate: 5, BuyDate: buy, MaturityDate: maturity,
			Quantity: 1, FaceValueUsd: 1000, PriceUsd: 1000, UsdIdrAtPurchase: 16000,
		}
	}
	purchases := []domain.BondPurchase{
		lot("INDON48", bondCategoryGov),
		lot("PLN 48", bondCategoryCorp),
		lot("PLN 48", bondCategoryCorp),
	}

	got := summarizeByName(purchases, time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC), 0)
	want := map[string]string{"INDON48": bondCategoryGov, "PLN 48": bondCategoryCorp}
	if len(got) != len(want) {
		t.Fatalf("got %d bonds, want %d", len(got), len(want))
	}
	for _, b := range got {
		if b.CategoryKey != want[b.BondName] {
			t.Errorf("%s: category %q, want %q", b.BondName, b.CategoryKey, want[b.BondName])
		}
	}
}

func TestBondPurchaseRequestCategory(t *testing.T) {
	base := BondPurchaseRequest{
		BondName: "PLN 48", InterestRate: 5, BuyDate: mustDate(t, "2026-01-10"),
		MaturityDate: mustDate(t, "2048-05-15"), Quantity: 1, FaceValueUsd: 1000,
		PriceUsd: 1000, UsdIdrAtPurchase: 16000,
	}

	tests := []struct {
		name         string
		categoryKey  string
		wantErr      bool
		wantWriteKey string
	}{
		{name: "omitted defaults to government", categoryKey: "", wantWriteKey: bondCategoryGov},
		{name: "government", categoryKey: bondCategoryGov, wantWriteKey: bondCategoryGov},
		{name: "corporate", categoryKey: bondCategoryCorp, wantWriteKey: bondCategoryCorp},
		{name: "any other category is rejected", categoryKey: "saham", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := base
			req.CategoryKey = tt.categoryKey
			err := req.validate()
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("validate() = %v, want ErrInvalidInput", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validate() = %v", err)
			}
			if got := req.toWrite().CategoryKey; got != tt.wantWriteKey {
				t.Errorf("toWrite().CategoryKey = %q, want %q", got, tt.wantWriteKey)
			}
		})
	}
}

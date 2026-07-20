package service

import (
	"errors"
	"testing"

	"wealthfolio/backend/internal/domain"
)

func TestComputeHoldingValue(t *testing.T) {
	rate := &domain.RateEntry{
		Antam:     1200, // Rp 1,200,000/g (thousands of IDR)
		Kinghalim: 1150,
		Ubs:       1180,
		UsdIdr:    15800, // full IDR per 1 USD
	}

	tests := []struct {
		name        string
		categoryKey string
		input       HoldingInput
		rate        *domain.RateEntry
		wantValue   int64
		wantDetail  string
		wantErr     error
	}{
		{
			name:        "gold priced with Antam (default brand)",
			categoryKey: "logam_mulia",
			input:       HoldingInput{Gram: 10, Qty: 1, Brand: "Antam"},
			rate:        rate,
			wantValue:   10 * 1200,
			wantDetail:  "10 g",
		},
		{
			name:        "gold priced with unrecognized brand falls back to Antam",
			categoryKey: "logam_mulia",
			input:       HoldingInput{Gram: 5, Qty: 1, Brand: ""},
			rate:        rate,
			wantValue:   5 * 1200,
			wantDetail:  "5 g",
		},
		{
			name:        "gold priced with King Halim",
			categoryKey: "logam_mulia",
			input:       HoldingInput{Gram: 10, Qty: 2, Brand: "King Halim"},
			rate:        rate,
			wantValue:   10 * 2 * 1150,
			wantDetail:  "2 × 10 g",
		},
		{
			name:        "gold priced with UBS",
			categoryKey: "logam_mulia",
			input:       HoldingInput{Gram: 3, Qty: 1, Brand: "UBS"},
			rate:        rate,
			wantValue:   3 * 1180,
			wantDetail:  "3 g",
		},
		{
			name:        "gold with qty<=0 defaults to qty=1",
			categoryKey: "logam_mulia",
			input:       HoldingInput{Gram: 4, Qty: 0, Brand: "Antam"},
			rate:        rate,
			wantValue:   4 * 1200,
			wantDetail:  "4 g",
		},
		{
			name:        "gold manual value fallback when gram<=0",
			categoryKey: "logam_mulia",
			input:       HoldingInput{Gram: 0, ValueIdr: 50000, Detail: "manual override"},
			rate:        rate,
			wantValue:   50000,
			wantDetail:  "manual override",
		},
		{
			name:        "gold with no rate entry and gram>0 errors",
			categoryKey: "logam_mulia",
			input:       HoldingInput{Gram: 10, Qty: 1, Brand: "Antam"},
			rate:        nil,
			wantErr:     ErrNoRateEntry,
		},
		{
			name:        "bonds_usd priced from usd_idr",
			categoryKey: "bonds_usd",
			input:       HoldingInput{UsdValue: 1000},
			rate:        rate,
			wantValue:   round64(1000 * (15800.0 / 1000)),
			wantDetail:  "1000 USD",
		},
		{
			name:        "bonds_usd manual value fallback when usd_value<=0",
			categoryKey: "bonds_usd",
			input:       HoldingInput{UsdValue: 0, ValueIdr: 20000, Detail: "manual bond"},
			rate:        rate,
			wantValue:   20000,
			wantDetail:  "manual bond",
		},
		{
			name:        "bonds_usd with no rate entry and usd_value>0 errors",
			categoryKey: "bonds_usd",
			input:       HoldingInput{UsdValue: 500},
			rate:        nil,
			wantErr:     ErrNoRateEntry,
		},
		{
			name:        "uang_tunai USD priced from usd_idr",
			categoryKey: "uang_tunai",
			input:       HoldingInput{Currency: "USD", UsdValue: 200},
			rate:        rate,
			wantValue:   round64(200 * (15800.0 / 1000)),
			wantDetail:  "200 USD",
		},
		{
			name:        "uang_tunai IDR uses value_idr directly",
			categoryKey: "uang_tunai",
			input:       HoldingInput{Currency: "IDR", ValueIdr: 75000, Detail: "cash IDR"},
			rate:        rate,
			wantValue:   75000,
			wantDetail:  "cash IDR",
		},
		{
			name:        "uang_tunai with unset currency treated as IDR",
			categoryKey: "uang_tunai",
			input:       HoldingInput{ValueIdr: 12345, Detail: "cash"},
			rate:        rate,
			wantValue:   12345,
			wantDetail:  "cash",
		},
		{
			name:        "plain category uses value_idr directly",
			categoryKey: "saham",
			input:       HoldingInput{ValueIdr: 300000, Detail: "BBCA shares"},
			rate:        rate,
			wantValue:   300000,
			wantDetail:  "BBCA shares",
		},
		{
			name:        "plain category with no rate entry is fine (rate not needed)",
			categoryKey: "properti",
			input:       HoldingInput{ValueIdr: 1000000, Detail: "house"},
			rate:        nil,
			wantValue:   1000000,
			wantDetail:  "house",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotDetail, err := ComputeHoldingValue(tt.categoryKey, tt.input, tt.rate)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotValue != tt.wantValue {
				t.Errorf("value_idr = %d, want %d", gotValue, tt.wantValue)
			}
			if gotDetail != tt.wantDetail {
				t.Errorf("detail = %q, want %q", gotDetail, tt.wantDetail)
			}
		})
	}
}

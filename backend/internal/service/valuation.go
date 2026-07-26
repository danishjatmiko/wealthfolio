// Package service contains Etherna's business logic: value derivation
// for holdings, snapshot mutability rules, dashboard/progress aggregation,
// and target progress computation.
package service

import (
	"fmt"
	"math"

	"wealthfolio/backend/internal/domain"
)

// HoldingInput is the raw, not-yet-priced shape of a holding write request
// (POST/PUT body), after basic presence validation but before value
// derivation. Zero values mean "not provided" for every field here.
type HoldingInput struct {
	Gram     float64
	Qty      float64
	Brand    string
	Currency string
	UsdValue float64
	ValueIdr int64
	Detail   string
}

// GoldPricePerGram resolves the Rp-per-gram price (in full/raw IDR,
// matching rate_entries.antam/kinghalim/ubs) for a given gold brand.
// Unrecognized/empty brands (including "Antam") fall back to the Antam
// rate.
func GoldPricePerGram(brand string, rate domain.RateEntry) float64 {
	switch brand {
	case "King Halim":
		return rate.Kinghalim
	case "UBS":
		return rate.Ubs
	default:
		return rate.Antam
	}
}

// ComputeHoldingValue derives value_idr (in full/raw IDR) and the display
// detail string for a holding, given its category key and raw input. rate
// is the user's latest rate_entries row, or nil if the user has none yet.
//
// Returns ErrNoRateEntry if the category/input combination requires a rate
// entry (gold priced by gram, or USD-denominated bonds/cash) and none is
// available, and no manual value_idr fallback was supplied.
func ComputeHoldingValue(categoryKey string, input HoldingInput, rate *domain.RateEntry) (int64, string, error) {
	switch {
	case categoryKey == "logam_mulia":
		if input.Gram > 0 {
			if rate == nil {
				return 0, "", ErrNoRateEntry
			}
			qty := input.Qty
			if qty <= 0 {
				qty = 1
			}
			value := input.Gram * qty * GoldPricePerGram(input.Brand, *rate)
			var detail string
			if qty > 1 {
				detail = fmt.Sprintf("%v × %v g", qty, input.Gram)
			} else {
				detail = fmt.Sprintf("%v g", input.Gram)
			}
			return round64(value), detail, nil
		}
		return input.ValueIdr, input.Detail, nil

	case categoryKey == "bonds_usd" || categoryKey == "us_etf" || (categoryKey == "uang_tunai" && input.Currency == "USD"):
		if input.UsdValue > 0 {
			if rate == nil {
				return 0, "", ErrNoRateEntry
			}
			value := input.UsdValue * rate.UsdIdr
			detail := fmt.Sprintf("%v USD", input.UsdValue)
			return round64(value), detail, nil
		}
		return input.ValueIdr, input.Detail, nil

	default:
		return input.ValueIdr, input.Detail, nil
	}
}

// round64 rounds a genuinely-computed float (gold gram×price, or USD×rate)
// to the nearest whole Rupiah for bigint storage — not a unit-scaling
// operation, just float-to-int64 rounding.
func round64(v float64) int64 {
	return int64(math.Round(v))
}

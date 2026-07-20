package service

import "wealthfolio/backend/internal/domain"

// netEquity sums asset holdings and liability holdings together (liability
// values are stored as positive magnitudes and added, not subtracted).
func netEquity(holdings []domain.Holding) int64 {
	var total int64
	for _, h := range holdings {
		total += h.ValueIdr
	}
	return total
}

// investedTotal sums only asset-kind (non-liability) holdings.
func investedTotal(holdings []domain.Holding) int64 {
	var total int64
	for _, h := range holdings {
		if !h.IsLiability {
			total += h.ValueIdr
		}
	}
	return total
}

// liabilityTotal sums only liability-kind holdings.
func liabilityTotal(holdings []domain.Holding) int64 {
	var total int64
	for _, h := range holdings {
		if h.IsLiability {
			total += h.ValueIdr
		}
	}
	return total
}

// percentOf returns part/whole*100, or 0 if whole is 0 (avoids division by
// zero throughout the dashboard/target computations).
func percentOf(part, whole float64) float64 {
	if whole == 0 {
		return 0
	}
	return part / whole * 100
}

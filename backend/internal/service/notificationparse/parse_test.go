package notificationparse

import (
	"regexp"
	"testing"
)

func TestNormalizeAmount(t *testing.T) {
	idFormat := AmountFormat{ThousandSep: '.', DecimalSep: ','}

	tests := []struct {
		name    string
		raw     string
		format  AmountFormat
		want    int64
		wantErr bool
	}{
		{"plain thousands", "50.000", idFormat, 50000, false},
		{"currency prefix and label", "Rp50.000", idFormat, 50000, false},
		{"decimal part discarded", "50.000,50", idFormat, 50000, false},
		{"no separators", "5000", idFormat, 5000, false},
		{"comma thousands, dot decimal", "50,000.00", AmountFormat{ThousandSep: ',', DecimalSep: '.'}, 50000, false},
		{"no digits at all", "Rp", idFormat, 0, true},
		{"sub-thousand amount preserved exactly", "1.903", idFormat, 1903, false},
		{"another sub-thousand amount preserved exactly", "1.234", idFormat, 1234, false},
		{"tiny amount preserved exactly", "1", idFormat, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeAmount(tt.raw, tt.format)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeAmount(%q) = %d, want error", tt.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeAmount(%q) unexpected error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Errorf("NormalizeAmount(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	format := AmountFormat{ThousandSep: '.', DecimalSep: ','}

	paymentPattern := Pattern{
		Regex: regexp.MustCompile(`(?P<merchant>.+) paid Rp(?P<amount>[\d.]+)`),
		Field: FieldText,
	}
	topUpPattern := Pattern{
		// No "merchant" group at all — a top-up notification only ever
		// carries an amount.
		Regex: regexp.MustCompile(`Top up of Rp(?P<amount>[\d.]+) successful`),
		Field: FieldText,
	}

	t.Run("matches the first pattern whose field has a hit", func(t *testing.T) {
		got, ok := Match([]Pattern{paymentPattern, topUpPattern}, format, "", "You paid Warung Kopi paid Rp50.000", "")
		if !ok {
			t.Fatal("Match() ok = false, want true")
		}
		if got.AmountIdr != 50000 || got.Merchant != "You paid Warung Kopi" {
			t.Errorf("Match() = %+v, want amount=50000 merchant=%q", got, "You paid Warung Kopi")
		}
	})

	t.Run("falls through to next pattern when merchant group matches but is empty", func(t *testing.T) {
		// Matches the same text as topUpPattern, but its "merchant" group
		// has nothing to capture before "Top up of Rp" — Match must treat
		// that as no-match rather than returning an empty merchant name.
		emptyMerchant := Pattern{
			Regex: regexp.MustCompile(`(?P<merchant>.*)Top up of Rp(?P<amount>[\d.]+) successful`),
			Field: FieldText,
		}
		got, ok := Match([]Pattern{emptyMerchant, topUpPattern}, format, "", "Top up of Rp20.000 successful", "")
		if !ok {
			t.Fatal("Match() ok = false, want true")
		}
		if got.AmountIdr != 20000 || got.Merchant != "" {
			t.Errorf("Match() = %+v, want amount=20000 merchant=\"\"", got)
		}
	})

	t.Run("no pattern matches", func(t *testing.T) {
		_, ok := Match([]Pattern{paymentPattern}, format, "", "some unrelated promo notification", "")
		if ok {
			t.Fatal("Match() ok = true, want false")
		}
	})
}

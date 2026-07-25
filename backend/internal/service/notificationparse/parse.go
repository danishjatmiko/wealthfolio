// Package notificationparse turns the raw title/text/bigText of a payment
// notification (forwarded as-is by the Android app — it does no parsing
// itself) into a structured amount + merchant, using a per-app catalog of
// regex patterns. This package itself is pure and DB-free: it only knows
// how to apply an already-loaded list of patterns to already-loaded text.
// Loading the catalog from Postgres and caching it lives in
// service.NotificationCatalogService — parsing stays server-side (rather
// than on-device) because notification formats drift without notice, and
// editing a DB row is much cheaper to react with than shipping a new APK.
package notificationparse

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// Field names which raw notification field a Pattern's regex is tested
// against.
type Field string

const (
	FieldTitle   Field = "title"
	FieldText    Field = "text"
	FieldBigText Field = "big_text"
)

// Pattern is one compiled matching rule. Regex must contain a named
// "amount" capture group; a named "merchant" group is optional but, when
// present in the regex, must also be non-empty for the pattern to count as
// a match (see Match).
type Pattern struct {
	Regex *regexp.Regexp
	Field Field
}

// AmountFormat describes how one app renders IDR amounts as text — e.g.
// "50.000" (ThousandSep='.') vs "50,000.00" (ThousandSep=',', DecimalSep
// = '.'). This is a property of the app, not of one notification
// template, so it travels alongside the pattern list rather than living on
// individual patterns.
type AmountFormat struct {
	ThousandSep byte
	DecimalSep  byte
}

// ParsedTransaction is what Match extracts once a pattern's format is
// known. AmountIdr is in THOUSANDS of rupiah, like every other IDR field
// in this codebase (fixed_expenses.amount_idr, etc. — see
// frontend/src/lib/format.ts's convention note), not full rupiah.
type ParsedTransaction struct {
	AmountIdr int64
	Merchant  string
}

// errNoDigits is returned by NormalizeAmount when raw contains nothing
// that could be a digit once separators are accounted for.
var errNoDigits = errors.New("notificationparse: no digits in amount")

// Match tries patterns in the order given (callers pre-sort by ascending
// priority) against whichever raw field each one names, until one
// produces a usable result. A pattern whose regex doesn't match its
// field, whose "amount" capture is empty or fails to normalize under
// format, or whose regex declares a "merchant" group that came back empty,
// is treated as no-match — Match falls through to the next pattern rather
// than than returning a partial result, so a notification only ever
// becomes a Fixed Expense with both an amount and (when the pattern
// promised one) a merchant.
func Match(patterns []Pattern, format AmountFormat, title, text, bigText string) (ParsedTransaction, bool) {
	for _, p := range patterns {
		raw := fieldValue(p.Field, title, text, bigText)
		if raw == "" {
			continue
		}
		submatch := p.Regex.FindStringSubmatch(raw)
		if submatch == nil {
			continue
		}

		names := p.Regex.SubexpNames()
		var amountRaw, merchant string
		merchantGroupPresent := false
		for i, name := range names {
			switch name {
			case "amount":
				amountRaw = submatch[i]
			case "merchant":
				merchantGroupPresent = true
				merchant = strings.TrimSpace(submatch[i])
			}
		}
		if amountRaw == "" {
			continue
		}
		amount, err := NormalizeAmount(amountRaw, format)
		if err != nil {
			continue
		}
		if merchantGroupPresent && merchant == "" {
			continue
		}

		return ParsedTransaction{AmountIdr: amount, Merchant: merchant}, true
	}
	return ParsedTransaction{}, false
}

func fieldValue(field Field, title, text, bigText string) string {
	switch field {
	case FieldTitle:
		return title
	case FieldText:
		return text
	case FieldBigText:
		return bigText
	default:
		return ""
	}
}

// NormalizeAmount turns a captured amount substring (e.g. "Rp50.000") into
// an int64 amount in THOUSANDS of rupiah — the unit every monetary field
// in this codebase uses (see frontend/src/lib/format.ts's convention
// note), not full rupiah. Any fractional part (after format.DecimalSep) is
// discarded first, then every character that isn't a digit or
// format.ThousandSep (currency symbols, spaces, etc.) is dropped from
// what remains, then the resulting full-rupiah figure is divided by 1000
// and rounded to the nearest thousand — the same precision loss a user
// manually typing an amount into the "Amount (rb Rp)" field already
// accepts, since that field can't express sub-thousand amounts either.
func NormalizeAmount(raw string, format AmountFormat) (int64, error) {
	whole := raw
	if idx := strings.IndexByte(raw, format.DecimalSep); idx >= 0 {
		whole = raw[:idx]
	}

	var digits strings.Builder
	for i := 0; i < len(whole); i++ {
		if c := whole[i]; c >= '0' && c <= '9' {
			digits.WriteByte(c)
		}
	}
	if digits.Len() == 0 {
		return 0, errNoDigits
	}
	fullRupiah, err := strconv.ParseInt(digits.String(), 10, 64)
	if err != nil {
		return 0, err
	}
	return (fullRupiah + 500) / 1000, nil
}

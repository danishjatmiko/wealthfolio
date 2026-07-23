// Package notificationparse turns the raw title/text/bigText of a
// GoPay/DANA/BCA payment notification (forwarded as-is by the Android app —
// it does no parsing itself) into a structured amount + merchant. Parsing
// lives here, in the backend, rather than on-device: notification formats
// drift without notice, and a Go redeploy is much cheaper to react with
// than shipping a new APK to the phone.
package notificationparse

const (
	SourceGoPay = "gopay"
	SourceDANA  = "dana"
	SourceBCA   = "bca"
)

// ParsedTransaction is what a source parser extracts from a notification,
// once its format is known.
type ParsedTransaction struct {
	AmountIdr int64
	Merchant  string
}

// Parse dispatches to the parser for source. ok is false for an unknown
// source, or when the notification doesn't match any known transaction
// pattern for that source (promo/marketing notifications, unrecognized
// format changes, etc.) — callers treat that as a terminal "ignored"
// outcome, not a retryable failure.
func Parse(source, title, text, bigText string) (ParsedTransaction, bool) {
	switch source {
	case SourceGoPay:
		return parseGopay(title, text, bigText)
	case SourceDANA:
		return parseDana(title, text, bigText)
	case SourceBCA:
		return parseBCA(title, text, bigText)
	default:
		return ParsedTransaction{}, false
	}
}

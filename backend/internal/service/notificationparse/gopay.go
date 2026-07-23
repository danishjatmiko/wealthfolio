package notificationparse

// parseGopay extracts amount/merchant from a GoPay payment notification.
// Not yet implemented — GoPay's real notification format hasn't been
// captured yet. Everything is reported unparseable until this is written
// against real samples collected once the Android app is forwarding raw
// notifications (see the Android plan's build order).
func parseGopay(title, text, bigText string) (ParsedTransaction, bool) {
	return ParsedTransaction{}, false
}

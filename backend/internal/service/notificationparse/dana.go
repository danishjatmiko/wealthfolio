package notificationparse

// parseDana extracts amount/merchant from a DANA payment notification. Not
// yet implemented — see parseGopay's comment; same situation, different
// app.
func parseDana(title, text, bigText string) (ParsedTransaction, bool) {
	return ParsedTransaction{}, false
}

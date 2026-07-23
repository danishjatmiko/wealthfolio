package notificationparse

// parseBCA extracts amount/merchant from a BCA mobile banking (m-BCA)
// payment notification. Not yet implemented — see parseGopay's comment;
// same situation, different app.
func parseBCA(title, text, bigText string) (ParsedTransaction, bool) {
	return ParsedTransaction{}, false
}

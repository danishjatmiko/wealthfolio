// Package httpapi wires the chi router, request/response JSON handling, and
// per-resource HTTP handlers for the Etherna API.
package httpapi

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes v as a JSON response body with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

type errorBody struct {
	Error string `json:"error"`
}

// writeError writes {"error": msg} with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// decodeJSON decodes the request body into v.
func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return errBodyRequired
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

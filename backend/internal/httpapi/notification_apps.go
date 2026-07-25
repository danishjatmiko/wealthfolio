package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"wealthfolio/backend/internal/service"
)

// listNotificationApps serves both the Android app's catalog sync (fetched
// on app-open and on Settings-open) and the admin listing — same data
// either way, so one handler covers both.
func (h *Handler) listNotificationApps(w http.ResponseWriter, r *http.Request) {
	apps, err := h.svc.NotificationCatalog.ListApps(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, apps)
}

type notificationAppRequest struct {
	Source            string   `json:"source"`
	DisplayName       string   `json:"display_name"`
	PackageNames      []string `json:"package_names"`
	AmountThousandSep string   `json:"amount_thousand_sep"`
	AmountDecimalSep  string   `json:"amount_decimal_sep"`
	Enabled           *bool    `json:"enabled"`
}

func (req notificationAppRequest) toServiceRequest() service.NotificationAppRequest {
	return service.NotificationAppRequest{
		Source:            req.Source,
		DisplayName:       req.DisplayName,
		PackageNames:      req.PackageNames,
		AmountThousandSep: req.AmountThousandSep,
		AmountDecimalSep:  req.AmountDecimalSep,
		Enabled:           req.Enabled,
	}
}

// createNotificationApp is an admin-only call (same session-cookie auth as
// every other route — no separate admin auth surface) to add a new
// supported app to the catalog without a backend deploy.
func (h *Handler) createNotificationApp(w http.ResponseWriter, r *http.Request) {
	var req notificationAppRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	app, err := h.svc.NotificationCatalog.CreateApp(r.Context(), req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

// updateNotificationApp is a full replace of an existing catalog entry
// (display name, package names, amount format, enabled flag) — admin-only.
func (h *Handler) updateNotificationApp(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")

	var req notificationAppRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	app, err := h.svc.NotificationCatalog.UpdateApp(r.Context(), source, req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, app)
}

// listNotificationPatterns is admin-only — needed to see an app's existing
// priorities/regexes before curling in a new one or editing one in place.
func (h *Handler) listNotificationPatterns(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	patterns, err := h.svc.NotificationCatalog.ListPatterns(r.Context(), source)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, patterns)
}

type notificationPatternRequest struct {
	Priority    *int   `json:"priority"`
	Field       string `json:"field"`
	Regex       string `json:"regex"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

func (req notificationPatternRequest) toServiceRequest() service.NotificationPatternRequest {
	return service.NotificationPatternRequest{
		Priority:    req.Priority,
		Field:       req.Field,
		Regex:       req.Regex,
		Description: req.Description,
		Enabled:     req.Enabled,
	}
}

// createNotificationPattern is admin-only — adds one more regex template
// to try for source, priority auto-assigned when omitted.
func (h *Handler) createNotificationPattern(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")

	var req notificationPatternRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	pattern, err := h.svc.NotificationCatalog.CreatePattern(r.Context(), source, req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pattern)
}

// updateNotificationPattern is admin-only — full replace of an existing
// pattern by id.
func (h *Handler) updateNotificationPattern(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pattern id")
		return
	}

	var req notificationPatternRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	pattern, err := h.svc.NotificationCatalog.UpdatePattern(r.Context(), id, req.toServiceRequest())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pattern)
}

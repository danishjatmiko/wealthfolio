package service

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/google/uuid"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/domain"
	"wealthfolio/backend/internal/service/notificationparse"
)

// catalogEntry is one source's cached, ready-to-match state.
type catalogEntry struct {
	format   notificationparse.AmountFormat
	patterns []notificationparse.Pattern
}

// NotificationCatalogService bridges the DB-backed notification app/pattern
// catalog to the pure notificationparse matching engine, and owns the admin
// CRUD surface (the /notification-apps* routes) used to grow the catalog
// without a backend redeploy. The parse-time cache is explicit-invalidation
// only, no TTL: every catalog write flows through this service's own
// methods, so there's no path for the DB and the cache to drift apart
// between invalidations, and no staleness window to reason about.
type NotificationCatalogService struct {
	repos *db.Repos

	mu    sync.RWMutex
	cache map[string]catalogEntry // by source; nil until first warm
}

func NewNotificationCatalogService(repos *db.Repos) *NotificationCatalogService {
	return &NotificationCatalogService{repos: repos}
}

// Parse loads (and lazily warms) the pattern cache, then matches source's
// notification text against it. ok is false both for an unknown/disabled
// source and for a known source whose patterns didn't match — callers
// treat both as "ignored", not an error.
func (s *NotificationCatalogService) Parse(ctx context.Context, source, title, text, bigText string) (notificationparse.ParsedTransaction, bool, error) {
	cache, err := s.warmedCache(ctx)
	if err != nil {
		return notificationparse.ParsedTransaction{}, false, err
	}
	entry, ok := cache[source]
	if !ok {
		return notificationparse.ParsedTransaction{}, false, nil
	}
	parsed, ok := notificationparse.Match(entry.patterns, entry.format, title, text, bigText)
	return parsed, ok, nil
}

func (s *NotificationCatalogService) warmedCache(ctx context.Context) (map[string]catalogEntry, error) {
	s.mu.RLock()
	cache := s.cache
	s.mu.RUnlock()
	if cache != nil {
		return cache, nil
	}
	return s.rebuildCache(ctx)
}

// Invalidate reloads the cache from the DB. Called after every admin write
// so a curl'd pattern edit is live on the very next ingestion — no restart
// required.
func (s *NotificationCatalogService) Invalidate(ctx context.Context) error {
	_, err := s.rebuildCache(ctx)
	return err
}

func (s *NotificationCatalogService) rebuildCache(ctx context.Context) (map[string]catalogEntry, error) {
	apps, err := s.repos.NotificationApps.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	sourceByAppID := make(map[uuid.UUID]string, len(apps))
	cache := make(map[string]catalogEntry, len(apps))
	for _, a := range apps {
		sourceByAppID[a.ID] = a.Source
		if a.Enabled {
			cache[a.Source] = catalogEntry{
				format: notificationparse.AmountFormat{
					ThousandSep: separatorByte(a.AmountThousandSep),
					DecimalSep:  separatorByte(a.AmountDecimalSep),
				},
			}
		}
	}

	patterns, err := s.repos.NotificationPatterns.ListAllEnabled(ctx)
	if err != nil {
		return nil, err
	}
	// ListAllEnabled is source-then-priority ordered, so appending in
	// order keeps each entry's patterns in the priority order Match
	// requires.
	for _, p := range patterns {
		source, ok := sourceByAppID[p.AppID]
		if !ok {
			continue
		}
		entry, ok := cache[source]
		if !ok {
			continue // app disabled
		}
		re, err := regexp.Compile(p.Regex)
		if err != nil {
			// The admin API validates regexes at write time, so this
			// only fires if one somehow slipped through — skip it
			// rather than failing cache warmup for every other
			// pattern.
			continue
		}
		entry.patterns = append(entry.patterns, notificationparse.Pattern{
			Regex: re,
			Field: notificationparse.Field(p.Field),
		})
		cache[source] = entry
	}

	s.mu.Lock()
	s.cache = cache
	s.mu.Unlock()
	return cache, nil
}

func separatorByte(s string) byte {
	if len(s) == 0 {
		return 0
	}
	return s[0]
}

// ListApps returns the full catalog (enabled or not) for the Android sync
// endpoint and the admin listing.
func (s *NotificationCatalogService) ListApps(ctx context.Context) ([]domain.NotificationApp, error) {
	return s.repos.NotificationApps.ListAll(ctx)
}

// NotificationAppRequest is the request body shape for both creating and
// updating a catalog entry. Source is only read on create — update targets
// the source named in the URL path.
type NotificationAppRequest struct {
	Source            string
	DisplayName       string
	PackageNames      []string
	AmountThousandSep string // defaults to "." when empty
	AmountDecimalSep  string // defaults to "," when empty
	Enabled           *bool  // defaults to true when nil
}

func (req NotificationAppRequest) toWrite(source string) (db.NotificationAppWrite, error) {
	if req.DisplayName == "" {
		return db.NotificationAppWrite{}, fmt.Errorf("%w: display_name is required", ErrInvalidInput)
	}
	thousandSep := req.AmountThousandSep
	if thousandSep == "" {
		thousandSep = "."
	}
	decimalSep := req.AmountDecimalSep
	if decimalSep == "" {
		decimalSep = ","
	}
	if len(thousandSep) != 1 || len(decimalSep) != 1 {
		return db.NotificationAppWrite{}, fmt.Errorf("%w: amount_thousand_sep/amount_decimal_sep must be exactly one character", ErrInvalidInput)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	return db.NotificationAppWrite{
		Source:            source,
		DisplayName:       req.DisplayName,
		PackageNames:      req.PackageNames,
		AmountThousandSep: thousandSep,
		AmountDecimalSep:  decimalSep,
		Enabled:           enabled,
	}, nil
}

// CreateApp adds a new catalog entry. Invalidates the parse cache on
// success so the new app is immediately eligible to match, even before it
// has any patterns.
func (s *NotificationCatalogService) CreateApp(ctx context.Context, req NotificationAppRequest) (domain.NotificationApp, error) {
	if req.Source == "" {
		return domain.NotificationApp{}, fmt.Errorf("%w: source is required", ErrInvalidInput)
	}
	exists, err := s.repos.NotificationApps.ExistsSource(ctx, req.Source)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	if exists {
		return domain.NotificationApp{}, fmt.Errorf("%w: source %q already exists", ErrInvalidInput, req.Source)
	}

	w, err := req.toWrite(req.Source)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	app, err := s.repos.NotificationApps.Create(ctx, w)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	if err := s.Invalidate(ctx); err != nil {
		return domain.NotificationApp{}, err
	}
	return app, nil
}

// UpdateApp fully replaces an existing catalog entry's fields and package
// list.
func (s *NotificationCatalogService) UpdateApp(ctx context.Context, source string, req NotificationAppRequest) (domain.NotificationApp, error) {
	w, err := req.toWrite(source)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	app, err := s.repos.NotificationApps.UpdateBySource(ctx, source, w)
	if err != nil {
		return domain.NotificationApp{}, err
	}
	if err := s.Invalidate(ctx); err != nil {
		return domain.NotificationApp{}, err
	}
	return app, nil
}

// ListPatterns returns every pattern (enabled or not) configured for
// source, priority-ascending.
func (s *NotificationCatalogService) ListPatterns(ctx context.Context, source string) ([]domain.NotificationPattern, error) {
	exists, err := s.repos.NotificationApps.ExistsSource(ctx, source)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, db.ErrNotFound
	}
	return s.repos.NotificationPatterns.ListByAppSource(ctx, source)
}

// NotificationPatternRequest is the request body shape for both creating
// and updating a pattern.
type NotificationPatternRequest struct {
	Priority    *int // auto-assigned (max existing + 10) when nil
	Field       string
	Regex       string
	Description string
	Enabled     *bool // defaults to true when nil
}

func (req NotificationPatternRequest) toWrite() (db.NotificationPatternWrite, error) {
	switch notificationparse.Field(req.Field) {
	case notificationparse.FieldTitle, notificationparse.FieldText, notificationparse.FieldBigText:
	default:
		return db.NotificationPatternWrite{}, fmt.Errorf("%w: field must be one of title, text, big_text", ErrInvalidInput)
	}
	if req.Regex == "" {
		return db.NotificationPatternWrite{}, fmt.Errorf("%w: regex is required", ErrInvalidInput)
	}
	re, err := regexp.Compile(req.Regex)
	if err != nil {
		return db.NotificationPatternWrite{}, fmt.Errorf("%w: invalid regex: %v", ErrInvalidInput, err)
	}
	hasAmount := false
	for _, name := range re.SubexpNames() {
		if name == "amount" {
			hasAmount = true
		}
	}
	if !hasAmount {
		return db.NotificationPatternWrite{}, fmt.Errorf(`%w: regex must contain a named "amount" capture group`, ErrInvalidInput)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	return db.NotificationPatternWrite{
		Field:       req.Field,
		Regex:       req.Regex,
		Description: req.Description,
		Enabled:     enabled,
	}, nil
}

// CreatePattern adds a new pattern under source's app.
func (s *NotificationCatalogService) CreatePattern(ctx context.Context, source string, req NotificationPatternRequest) (domain.NotificationPattern, error) {
	w, err := req.toWrite()
	if err != nil {
		return domain.NotificationPattern{}, err
	}
	if req.Priority != nil {
		w.Priority = *req.Priority
	} else {
		max, err := s.repos.NotificationPatterns.MaxPriorityForSource(ctx, source)
		if err != nil {
			return domain.NotificationPattern{}, err
		}
		w.Priority = max + 10
	}

	p, err := s.repos.NotificationPatterns.Create(ctx, source, w)
	if err != nil {
		return domain.NotificationPattern{}, err
	}
	if err := s.Invalidate(ctx); err != nil {
		return domain.NotificationPattern{}, err
	}
	return p, nil
}

// UpdatePattern overwrites an existing pattern's mutable fields. Priority
// is left unchanged when the request omits it.
func (s *NotificationCatalogService) UpdatePattern(ctx context.Context, id uuid.UUID, req NotificationPatternRequest) (domain.NotificationPattern, error) {
	w, err := req.toWrite()
	if err != nil {
		return domain.NotificationPattern{}, err
	}
	if req.Priority != nil {
		w.Priority = *req.Priority
	} else {
		existing, err := s.repos.NotificationPatterns.GetByID(ctx, id)
		if err != nil {
			return domain.NotificationPattern{}, err
		}
		w.Priority = existing.Priority
	}

	p, err := s.repos.NotificationPatterns.Update(ctx, id, w)
	if err != nil {
		return domain.NotificationPattern{}, err
	}
	if err := s.Invalidate(ctx); err != nil {
		return domain.NotificationPattern{}, err
	}
	return p, nil
}

// Command rates-sync fetches the day's Antam/UBS/King Halim gold prices
// (per gram, 1g denomination) and the USD/IDR rate, then logs into
// Etherna's own account and POSTs them to /api/v1/rates — the same
// endpoint the Rates page's manual entry form uses. Meant to run once a
// day via cron; see the env vars and crontab line below.
//
// Data sources — each brand/rate's own site, not a third-party aggregator:
//
//   - Antam & UBS: indogold.id's "harga-emas-hari-ini" page. Its price
//     table loads via an internal AJAX endpoint
//     (POST /home/get_data_pricelist) gated by a per-session CSRF-style
//     token embedded in the page's inline JS — so this fetches the page
//     first to pick up a session cookie and the current token, then POSTs
//     with both. product=comparison_antamxubs returns Antam and UBS
//     side by side in one response.
//   - King Halim: kinghalim.com/goldbarwithamala. This is a page-builder
//     site (no real API) — the 1 Gram bar's price is scraped directly out
//     of the rendered HTML.
//   - USD/IDR: Google Finance's USD-IDR quote page. Plain HTML, no JS
//     execution needed — the rate sits in a stable, labeled span.
//
// logammulia.com (Antam's own site) was tried first but returns a flat
// 403 from what looks like an Akamai WAF that blocks datacenter/VPS IP
// ranges — that's why Antam comes from indogold.id instead. All three
// scrapes here depend on markup/endpoints that could change without
// notice; every fetch function returns the raw response body in its
// error rather than silently producing a wrong number.
//
// If any single value's scrape fails, that value falls back to
// yesterday's rate entry (fetched via GET /rates/latest) instead of
// aborting the whole run — so e.g. indogold.id being briefly down still
// lets USD/IDR and King Halim update normally, with Antam/UBS just
// repeating their last known value (logged as a warning). This only
// fails outright if a value's scrape fails AND there's no previous rate
// entry to fall back to (the very first run).
//
// Required env vars:
//
//	ETHERNA_EMAIL       login email for your Etherna account
//	ETHERNA_PASSWORD    login password
//
// Optional env vars (default shown):
//
//	ETHERNA_API_BASE_URL      https://etherna.id/api/v1
//
// Build once on the VPS:
//
//	cd /opt/wealthfolio/backend && go build -o /usr/local/bin/rates-sync ./cmd/rates-sync
//
// Crontab (9AM WIB daily; adjust if the VPS isn't already on Asia/Jakarta):
//
//	0 9 * * * ETHERNA_EMAIL=you@example.com ETHERNA_PASSWORD=... /usr/local/bin/rates-sync >> /var/log/rates-sync.log 2>&1
//
// Prefer not to put secrets directly in crontab? Put the exports in
// /opt/wealthfolio/rates-sync.env and use:
//
//	0 9 * * * . /opt/wealthfolio/rates-sync.env && /usr/local/bin/rates-sync >> /var/log/rates-sync.log 2>&1
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type config struct {
	EthernaBaseURL  string
	EthernaEmail    string
	EthernaPassword string
}

func loadConfig() (config, error) {
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	required := func(key string) (string, error) {
		v := os.Getenv(key)
		if v == "" {
			return "", fmt.Errorf("missing required env var %s", key)
		}
		return v, nil
	}

	cfg := config{
		EthernaBaseURL: get("ETHERNA_API_BASE_URL", "https://etherna.id/api/v1"),
	}
	var err error
	if cfg.EthernaEmail, err = required("ETHERNA_EMAIL"); err != nil {
		return cfg, err
	}
	if cfg.EthernaPassword, err = required("ETHERNA_PASSWORD"); err != nil {
		return cfg, err
	}
	return cfg, nil
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

func main() {
	log.SetFlags(0)

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	client := &http.Client{Timeout: 20 * time.Second}

	// Logged in first (rather than last, as before) — the fallback path
	// below needs a token to read yesterday's rate before we can decide
	// what to post today.
	token, err := loginEtherna(client, cfg)
	if err != nil {
		log.Fatalf("etherna login: %v", err)
	}

	latest, hasLatest, err := fetchLatestRate(client, cfg, token)
	if err != nil {
		log.Fatalf("fetch latest rate (needed as a fallback if today's scrape fails): %v", err)
	}

	usdIdrFresh, usdIdrErr := fetchUsdIdrFromGoogle(client)
	antamFresh, ubsFresh, auErr := fetchAntamAndUbsFromIndogold()
	kinghalimFresh, khErr := fetchKingHalim(client)

	antam, antamFallback, err := resolveValue("Antam", antamFresh, auErr, latest.Antam, hasLatest)
	if err != nil {
		log.Fatalf("%v", err)
	}
	ubs, ubsFallback, err := resolveValue("UBS", ubsFresh, auErr, latest.Ubs, hasLatest)
	if err != nil {
		log.Fatalf("%v", err)
	}
	kinghalim, khFallback, err := resolveValue("King Halim", kinghalimFresh, khErr, latest.Kinghalim, hasLatest)
	if err != nil {
		log.Fatalf("%v", err)
	}
	usdIdr, usdIdrFallback, err := resolveValue("USD/IDR", usdIdrFresh, usdIdrErr, latest.UsdIdr, hasLatest)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if err := postRate(client, cfg, token, antam, kinghalim, ubs, usdIdr); err != nil {
		log.Fatalf("post rate: %v", err)
	}

	log.Printf(
		"rates updated for %s: antam=%.0f%s kinghalim=%.0f%s ubs=%.0f%s usd_idr=%.2f%s",
		jakartaToday(),
		antam, fallbackTag(antamFallback),
		kinghalim, fallbackTag(khFallback),
		ubs, fallbackTag(ubsFallback),
		usdIdr, fallbackTag(usdIdrFallback),
	)
}

// resolveValue returns the freshly-scraped value, unless fetchErr is set —
// in which case it falls back to the last known rate (with a logged
// warning) if one exists, or returns an error if there's truly nothing to
// fall back to (e.g. the very first run, before any rate entry exists).
func resolveValue(label string, fresh float64, fetchErr error, fallback float64, hasFallback bool) (value float64, usedFallback bool, err error) {
	if fetchErr == nil {
		return fresh, false, nil
	}
	if !hasFallback {
		return 0, false, fmt.Errorf("%s: %v (and no previous rate entry to fall back to)", label, fetchErr)
	}
	log.Printf("warning: %s fetch failed, using yesterday's value instead: %v", label, fetchErr)
	return fallback, true, nil
}

func fallbackTag(used bool) string {
	if used {
		return " (fallback)"
	}
	return ""
}

func jakartaToday() string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC // fallback if the VPS has no tzdata installed
	}
	return time.Now().In(loc).Format("2006-01-02")
}

// parseRupiah turns "Rp. 2,634,000" / "Rp 2,485,560.00" into 2634000 /
// 2485560.00 full Rupiah.
func parseRupiah(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "Rp.")
	s = strings.TrimPrefix(s, "Rp")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing rupiah amount %q: %w", s, err)
	}
	return v, nil
}

// --- USD/IDR (Google Finance) ---

var googleFinanceRatePattern = regexp.MustCompile(
	`(?s)United States Dollar / Indonesian Rupiah.{0,300}?<span jsname="Pdsbrc"[^>]*><span>([0-9,]+\.[0-9]+)</span>`,
)

func fetchUsdIdrFromGoogle(client *http.Client) (float64, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.google.com/finance/quote/USD-IDR", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("google finance returned %d (body len %d)", resp.StatusCode, len(body))
	}

	m := googleFinanceRatePattern.FindSubmatch(body)
	if m == nil {
		return 0, fmt.Errorf("couldn't find USD/IDR rate in google finance page (page structure may have changed; got %d bytes)", len(body))
	}
	rate, err := strconv.ParseFloat(strings.ReplaceAll(string(m[1]), ",", ""), 64)
	if err != nil {
		return 0, fmt.Errorf("parsing google finance rate %q: %w", m[1], err)
	}
	return rate, nil
}

// --- Antam & UBS (indogold.id) ---

var indogoldTokenPattern = regexp.MustCompile(`simulasi-token","([a-f0-9]+)"`)

type indogoldPriceResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		DataDenom map[string]map[string]struct {
			Harga string `json:"harga"`
		} `json:"data_denom"`
	} `json:"data"`
}

// fetchAntamAndUbsFromIndogold scrapes indogold.id's price-comparison
// page. Uses its own http.Client (rather than the shared one) because it
// needs a cookie jar: the page's AJAX price endpoint is gated by a
// session cookie plus a CSRF-style token embedded in that same page's
// inline JS, both obtained by loading the page first.
func fetchAntamAndUbsFromIndogold() (antam float64, ubs float64, err error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return 0, 0, err
	}
	client := &http.Client{Timeout: 20 * time.Second, Jar: jar}

	const pageURL = "https://www.indogold.id/harga-emas-hari-ini"
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("loading indogold page: %w", err)
	}
	pageBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("indogold page returned %d", resp.StatusCode)
	}

	tokenMatch := indogoldTokenPattern.FindSubmatch(pageBody)
	if tokenMatch == nil {
		return 0, 0, fmt.Errorf("couldn't find simulasi-token on indogold page (page structure may have changed)")
	}
	token := string(tokenMatch[1])

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("form", `{"product":"comparison_antamxubs"}`); err != nil {
		return 0, 0, err
	}
	if err := mw.WriteField("simulasi-token", token); err != nil {
		return 0, 0, err
	}
	if err := mw.Close(); err != nil {
		return 0, 0, err
	}

	const priceURL = "https://www.indogold.id/home/get_data_pricelist"
	priceReq, err := http.NewRequest(http.MethodPost, priceURL, &buf)
	if err != nil {
		return 0, 0, err
	}
	priceReq.Header.Set("Content-Type", mw.FormDataContentType())
	priceReq.Header.Set("User-Agent", userAgent)
	priceReq.Header.Set("Referer", pageURL)
	priceReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	priceResp, err := client.Do(priceReq)
	if err != nil {
		return 0, 0, fmt.Errorf("posting to indogold pricelist endpoint: %w", err)
	}
	defer priceResp.Body.Close()

	priceBody, err := io.ReadAll(priceResp.Body)
	if err != nil {
		return 0, 0, err
	}
	if priceResp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("indogold pricelist endpoint returned %d: %s", priceResp.StatusCode, priceBody)
	}

	var out indogoldPriceResponse
	if err := json.Unmarshal(priceBody, &out); err != nil {
		return 0, 0, fmt.Errorf("parsing indogold pricelist response: %w (body: %s)", err, priceBody)
	}
	if !out.Status {
		return 0, 0, fmt.Errorf("indogold pricelist endpoint reported failure: %s (body: %s)", out.Error, priceBody)
	}

	// Weight keys observed as "0.5", "1.0", "2.0", ... — fall back to a
	// bare "1" in case that ever changes.
	row, ok := out.Data.DataDenom["1.0"]
	if !ok {
		row, ok = out.Data.DataDenom["1"]
	}
	if !ok {
		return 0, 0, fmt.Errorf("no 1-gram row in indogold response (body: %s)", priceBody)
	}

	antamCell, ok := row["Antam"]
	if !ok {
		return 0, 0, fmt.Errorf("no Antam column in indogold 1-gram row (body: %s)", priceBody)
	}
	ubsCell, ok := row["UBS"]
	if !ok {
		return 0, 0, fmt.Errorf("no UBS column in indogold 1-gram row (body: %s)", priceBody)
	}

	antamRp, err := parseRupiah(antamCell.Harga)
	if err != nil {
		return 0, 0, fmt.Errorf("antam price: %w", err)
	}
	ubsRp, err := parseRupiah(ubsCell.Harga)
	if err != nil {
		return 0, 0, fmt.Errorf("ubs price: %w", err)
	}

	// rate_entries.antam/ubs store raw whole Rupiah (see migration
	// 00014_exact_rupiah_unit.sql) — same unit indogold returns, no
	// rescaling needed.
	return antamRp, ubsRp, nil
}

// --- King Halim ---

var kingHalimPricePattern = regexp.MustCompile(`(?s)1\s*Gr</b>.{0,1000}?Rp\s*([0-9,]+\.[0-9]{2})`)

func fetchKingHalim(client *http.Client) (float64, error) {
	req, err := http.NewRequest(http.MethodGet, "https://kinghalim.com/goldbarwithamala", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("kinghalim.com returned %d", resp.StatusCode)
	}

	m := kingHalimPricePattern.FindSubmatch(body)
	if m == nil {
		return 0, fmt.Errorf("couldn't find 1 Gr price on kinghalim.com/goldbarwithamala (page structure may have changed; got %d bytes)", len(body))
	}
	rp, err := parseRupiah("Rp " + string(m[1]))
	if err != nil {
		return 0, err
	}
	// Same raw-whole-Rupiah convention as Antam/UBS above — no rescaling.
	return rp, nil
}

// --- Etherna ---

type latestRate struct {
	Antam     float64 `json:"antam"`
	Kinghalim float64 `json:"kinghalim"`
	Ubs       float64 `json:"ubs"`
	UsdIdr    float64 `json:"usd_idr"`
}

// fetchLatestRate reads the most recent rate entry, used as the fallback
// source when today's scrape of a given value fails. hasLatest is false
// (not an error) when the account has no rate entries yet — the very
// first run has nothing to fall back to, which the caller must treat as
// a hard failure for whichever value is missing.
func fetchLatestRate(client *http.Client, cfg config, token string) (latestRate, bool, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.EthernaBaseURL+"/rates/latest", nil)
	if err != nil {
		return latestRate{}, false, err
	}
	req.Header.Set("Cookie", "wf_session="+token)

	resp, err := client.Do(req)
	if err != nil {
		return latestRate{}, false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return latestRate{}, false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return latestRate{}, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return latestRate{}, false, fmt.Errorf("GET /rates/latest returned %d: %s", resp.StatusCode, body)
	}

	var out latestRate
	if err := json.Unmarshal(body, &out); err != nil {
		return latestRate{}, false, fmt.Errorf("parsing /rates/latest response: %w (body: %s)", err, body)
	}
	return out, true, nil
}

func loginEtherna(client *http.Client, cfg config) (string, error) {
	body, err := json.Marshal(map[string]string{
		"email":    cfg.EthernaEmail,
		"password": cfg.EthernaPassword,
	})
	if err != nil {
		return "", err
	}

	resp, err := client.Post(cfg.EthernaBaseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned %d: %s", resp.StatusCode, respBody)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("parsing login response: %w (body: %s)", err, respBody)
	}
	return out.Token, nil
}

func postRate(client *http.Client, cfg config, token string, antam, kinghalim, ubs, usdIdr float64) error {
	body, err := json.Marshal(map[string]any{
		"entry_date": jakartaToday(),
		"antam":      antam,
		"kinghalim":  kinghalim,
		"ubs":        ubs,
		"usd_idr":    usdIdr,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.EthernaBaseURL+"/rates", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// Matches the Android app's AuthInterceptor: the backend's
	// AuthMiddleware just reads this cookie, whether a real cookie jar
	// or a manually-set header put it there.
	req.Header.Set("Cookie", "wf_session="+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST /rates returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

// Command rates-sync fetches the day's Antam/UBS/King Halim gold prices
// (per gram, 1g denomination) and the USD/IDR rate, then logs into
// Etherna's own account and POSTs them to /api/v1/rates — the same
// endpoint the Rates page's manual entry form uses. Meant to run once a
// day via cron; see the env vars and crontab line below.
//
// Data sources:
//   - Gold (Antam/UBS/King Halim): emas.maulanar.my.id ("Emas API ID") —
//     the only free source found that covers all three brands, including
//     King Halim, which no other public API/scraper lists. Needs a free
//     API key: sign up at https://emas.maulanar.my.id, then set
//     EMAS_API_KEY. Unauthenticated requests get a flat 401.
//   - USD/IDR: api.frankfurter.dev (ECB reference rates) — free, no key,
//     no meaningful rate limit for one request/day.
//
// Required env vars:
//
//	ETHERNA_EMAIL       login email for your Etherna account
//	ETHERNA_PASSWORD    login password
//	EMAS_API_KEY        API key from emas.maulanar.my.id
//
// Optional env vars (defaults shown):
//
//	ETHERNA_API_BASE_URL      https://etherna.id/api/v1
//	EMAS_API_BASE_URL         https://emas.maulanar.my.id
//	EMAS_ANTAM_RESOURCE       antam
//	EMAS_UBS_RESOURCE         galeri24    (UBS bars are sold through Galeri24)
//	EMAS_KINGHALIM_RESOURCE   kinghalim
//
// The three EMAS_*_RESOURCE values are my best read of the docs at
// emas.maulanar.my.id/docs, not something I could verify against a real
// key — after you sign up, hit
// https://emas.maulanar.my.id/api/prices?brand[eq]=antam with your key
// and confirm the "resource" field it actually returns; adjust the env
// vars if it differs. If a brand/resource pair matches nothing, this
// program prints the raw API response and exits non-zero rather than
// silently posting a wrong number.
//
// Build once on the VPS:
//
//	cd /opt/wealthfolio/backend && go build -o /usr/local/bin/rates-sync ./cmd/rates-sync
//
// Crontab (9AM WIB daily; adjust if the VPS isn't already on Asia/Jakarta):
//
//	0 9 * * * ETHERNA_EMAIL=you@example.com ETHERNA_PASSWORD=... EMAS_API_KEY=... /usr/local/bin/rates-sync >> /var/log/rates-sync.log 2>&1
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
	"net/http"
	"net/url"
	"os"
	"time"
)

type config struct {
	EthernaBaseURL    string
	EthernaEmail      string
	EthernaPassword   string
	EmasBaseURL       string
	EmasAPIKey        string
	AntamResource     string
	UbsResource       string
	KingHalimResource string
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
		EthernaBaseURL:    get("ETHERNA_API_BASE_URL", "https://etherna.id/api/v1"),
		EmasBaseURL:       get("EMAS_API_BASE_URL", "https://emas.maulanar.my.id"),
		AntamResource:     get("EMAS_ANTAM_RESOURCE", "antam"),
		UbsResource:       get("EMAS_UBS_RESOURCE", "galeri24"),
		KingHalimResource: get("EMAS_KINGHALIM_RESOURCE", "kinghalim"),
	}

	var err error
	if cfg.EthernaEmail, err = required("ETHERNA_EMAIL"); err != nil {
		return cfg, err
	}
	if cfg.EthernaPassword, err = required("ETHERNA_PASSWORD"); err != nil {
		return cfg, err
	}
	if cfg.EmasAPIKey, err = required("EMAS_API_KEY"); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func main() {
	log.SetFlags(0)

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	client := &http.Client{Timeout: 15 * time.Second}

	usdIdr, err := fetchUsdIdr(client)
	if err != nil {
		log.Fatalf("fetch USD/IDR: %v", err)
	}

	antam, err := fetchGoldPricePerGram(client, cfg, "antam", cfg.AntamResource)
	if err != nil {
		log.Fatalf("fetch Antam price: %v", err)
	}
	ubs, err := fetchGoldPricePerGram(client, cfg, "ubs", cfg.UbsResource)
	if err != nil {
		log.Fatalf("fetch UBS price: %v", err)
	}
	kinghalim, err := fetchGoldPricePerGram(client, cfg, "king halim", cfg.KingHalimResource)
	if err != nil {
		log.Fatalf("fetch King Halim price: %v", err)
	}

	token, err := loginEtherna(client, cfg)
	if err != nil {
		log.Fatalf("etherna login: %v", err)
	}

	if err := postRate(client, cfg, token, antam, kinghalim, ubs, usdIdr); err != nil {
		log.Fatalf("post rate: %v", err)
	}

	log.Printf(
		"rates updated for %s: antam=%.2f kinghalim=%.2f ubs=%.2f usd_idr=%.2f",
		jakartaToday(), antam, kinghalim, ubs, usdIdr,
	)
}

func jakartaToday() string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC // fallback if the VPS has no tzdata installed
	}
	return time.Now().In(loc).Format("2006-01-02")
}

// --- USD/IDR ---

func fetchUsdIdr(client *http.Client) (float64, error) {
	resp, err := client.Get("https://api.frankfurter.dev/v1/latest?base=USD&symbols=IDR")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("frankfurter returned %d: %s", resp.StatusCode, body)
	}

	var out struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, fmt.Errorf("parsing frankfurter response: %w (body: %s)", err, body)
	}
	rate, ok := out.Rates["IDR"]
	if !ok {
		return 0, fmt.Errorf("no IDR rate in frankfurter response: %s", body)
	}
	return rate, nil
}

// --- Gold prices (Emas API ID) ---

type priceRow struct {
	Brand        string  `json:"brand"`
	Resource     string  `json:"resource"`
	Weight       float64 `json:"weight"`
	SellPrice    float64 `json:"sell_price"`
	BuybackPrice float64 `json:"buyback_price"`
	UpdatedAt    string  `json:"updated_at"`
}

// parsePriceRows accepts either the documented {"data": [...]} envelope
// or a bare JSON array, since the exact response shape wasn't something
// I could confirm without a live API key — whichever this API actually
// returns, one of the two will parse.
func parsePriceRows(body []byte) ([]priceRow, error) {
	var wrapped struct {
		Data []priceRow `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Data) > 0 {
		return wrapped.Data, nil
	}
	var bare []priceRow
	if err := json.Unmarshal(body, &bare); err == nil && len(bare) > 0 {
		return bare, nil
	}
	return nil, fmt.Errorf("couldn't parse price rows from response body: %s", body)
}

func fetchGoldPricePerGram(client *http.Client, cfg config, brand, resource string) (float64, error) {
	q := url.Values{}
	q.Set("brand[eq]", brand)
	q.Set("resource[eq]", resource)
	q.Set("weight[eq]", "1")
	q.Set("sort_by", "updated_at")
	q.Set("order", "desc")
	q.Set("limit", "1")

	req, err := http.NewRequest(http.MethodGet, cfg.EmasBaseURL+"/api/prices?"+q.Encode(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("X-API-Key", cfg.EmasAPIKey)

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
		return 0, fmt.Errorf("emas API returned %d for brand=%q resource=%q: %s", resp.StatusCode, brand, resource, body)
	}

	rows, err := parsePriceRows(body)
	if err != nil {
		return 0, fmt.Errorf("brand=%q resource=%q: %w", brand, resource, err)
	}

	// rate_entries stores antam/kinghalim/ubs in full/raw IDR (see
	// backend/internal/domain/domain.go's RateEntry comment), matching what
	// this API already returns directly — no scaling needed.
	return rows[0].SellPrice, nil
}

// --- Etherna ---

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

package POToken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Default bgutil HTTP server URL (from bgutil-ytdlp-pot-provider)
// Docker: docker run -d -p 4416:4416 brainicism/bgutil-ytdlp-pot-provider

const DefaultBgUtilServerURL = "http://127.0.0.1:4416"

type BgUtilGenerator struct {

	// ServerURL is the base URL of the bgutil HTTP server

	ServerURL string

	// HTTPClient for making requests to the bgutil server

	HTTPClient *http.Client

	// Cache for generated tokens (keyed by content binding)

	cache     map[string]*CachedToken
	cacheLock sync.RWMutex

	// CacheTTL is how long to cache tokens before regenerating
	// Default: 5 hours (tokens are valid for ~6 hours)

	CacheTTL time.Duration

}

// CachedToken represents a cached PO token with expiration

type CachedToken struct {

	Token     string
	ExpiresAt time.Time

}

// BgUtilRequest represents a request to the bgutil /get_pot endpoint

type BgUtilRequest struct {

	ContentBinding         string `json:"content_binding,omitempty"`
	Proxy                  string `json:"proxy,omitempty"`
	BypassCache            bool   `json:"bypass_cache,omitempty"`
	SourceAddress          string `json:"source_address,omitempty"`
	DisableTLSVerification bool   `json:"disable_tls_verification,omitempty"`
	DisableInnertube       bool   `json:"disable_innertube,omitempty"`

}

// BgUtilResponse represents the response from bgutil /get_pot endpoint

type BgUtilResponse struct {

	PoToken        string    `json:"poToken"`
	ContentBinding string    `json:"contentBinding"`
	ExpiresAt      time.Time `json:"expiresAt"`
	Error          string    `json:"error,omitempty"`

}

// BgUtilPingResponse represents the response from bgutil /ping endpoint

type BgUtilPingResponse struct {

	ServerUptime float64 `json:"server_uptime"`
	Version      string  `json:"version"`

}

// GeneratorOptions configures the BgUtilGenerator

type GeneratorOptions struct {

	// ServerURL overrides the default bgutil server URL

	ServerURL string

	// Timeout for HTTP requests to the bgutil server
	// Default: 30 seconds

	Timeout time.Duration

	// CacheTTL is how long to cache tokens
	// Default: 5 hours

	CacheTTL time.Duration

	// Proxy URL to use for bgutil requests (passed to the server)

	Proxy string

}

// NewGenerator creates a new BgUtilGenerator with the given options
// If options is nil, defaults are used

func NewGenerator(Options *GeneratorOptions) *BgUtilGenerator {

	Generator := &BgUtilGenerator{

		ServerURL: DefaultBgUtilServerURL,

		cache:     make(map[string]*CachedToken),
		CacheTTL:  5 * time.Hour,

		HTTPClient: &http.Client{

			Timeout: 30 * time.Second,

		},

	}

	if Options != nil {

		if Options.ServerURL != "" {

			Generator.ServerURL = Options.ServerURL

		}

		if Options.Timeout > 0 {

			Generator.HTTPClient.Timeout = Options.Timeout

		}

		if Options.CacheTTL > 0 {

			Generator.CacheTTL = Options.CacheTTL

		}

	}

	return Generator

}

// Name returns the provider name

func (g *BgUtilGenerator) Name() string {

	return "bgutil-http"

}

// IsAvailable checks if the bgutil server is reachable

func (g *BgUtilGenerator) IsAvailable() bool {

	_, Err := g.Ping()

	return Err == nil

}

// Ping checks if the bgutil server is running and returns version info

func (g *BgUtilGenerator) Ping() (*BgUtilPingResponse, error) {

	Resp, Err := g.HTTPClient.Get(g.ServerURL + "/ping")

	if Err != nil {

		return nil, fmt.Errorf("bgutil server unreachable: %v", Err)

	}

	defer Resp.Body.Close()

	if Resp.StatusCode != http.StatusOK {

		return nil, fmt.Errorf("bgutil server returned status %d", Resp.StatusCode)

	}

	var PingResp BgUtilPingResponse

	if Err := json.NewDecoder(Resp.Body).Decode(&PingResp); Err != nil {

		return nil, fmt.Errorf("failed to decode ping response: %v", Err)

	}

	return &PingResp, nil

}

// RequestPoToken implements PoTokenProvider interface

func (g *BgUtilGenerator) RequestPoToken(Request *PoTokenRequest) (*PoTokenResponse, error) {

	ContentBinding, _ := GetContentBinding(Request)

	if ContentBinding == "" {

		ContentBinding = Request.VisitorData

	}

	Token, Err := g.GetPoToken(ContentBinding, nil)

	if Err != nil {

		return nil, Err

	}

	return &PoTokenResponse{

		PoToken: Token,

	}, nil

}

// GetPoToken fetches a PO token for the given content binding (typically visitor_data)
// Uses caching to avoid regenerating tokens unnecessarily

func (g *BgUtilGenerator) GetPoToken(ContentBinding string, Options *BgUtilRequest) (string, error) {

	// Check cache first

	g.cacheLock.RLock()

	if Cached, Ok := g.cache[ContentBinding]; Ok {

		if time.Now().Before(Cached.ExpiresAt) {

			g.cacheLock.RUnlock()

			return Cached.Token, nil

		}

	}

	g.cacheLock.RUnlock()

	// Generate new token

	Token, ExpiresAt, Err := g.generateToken(ContentBinding, Options)

	if Err != nil {

		return "", Err

	}

	// Cache the token

	g.cacheLock.Lock()

	g.cache[ContentBinding] = &CachedToken{

		Token:     Token,
		ExpiresAt: ExpiresAt,

	}

	g.cacheLock.Unlock()

	return Token, nil

}

// GetPoTokenWithVisitorData is a convenience method that generates a token using visitor_data

func (g *BgUtilGenerator) GetPoTokenWithVisitorData(VisitorData string) (string, error) {

	return g.GetPoToken(VisitorData, nil)

}

// extractSessionID extracts the account Session ID from a DataSyncID
// DataSyncID format is "SESSION_ID||..." - we need only the first part for GVS tokens

func extractSessionID(DataSyncID string) string {

	if DataSyncID == "" {

		return ""

	}

	// Split by || and take the first part

	Parts := strings.Split(DataSyncID, "||")

	if len(Parts) > 0 && Parts[0] != "" {

		return Parts[0]

	}

	return DataSyncID

}

// GetPoTokenForGVS generates a GVS context PO token (for video streaming)
// Uses visitor_data for logged-out users or session ID (first part of data_sync_id) for logged-in users

func (g *BgUtilGenerator) GetPoTokenForGVS(VisitorData string, DataSyncID string, IsAuthenticated bool) (string, error) {

	ContentBinding := VisitorData

	if IsAuthenticated && DataSyncID != "" {

		// Extract just the session ID from DataSyncID (format: "SESSION_ID||...")

		ContentBinding = extractSessionID(DataSyncID)

	}

	return g.GetPoToken(ContentBinding, nil)

}

// generateToken makes the actual HTTP request to the bgutil server

func (g *BgUtilGenerator) generateToken(ContentBinding string, Options *BgUtilRequest) (string, time.Time, error) {

	if Options == nil {

		Options = &BgUtilRequest{}

	}

	Options.ContentBinding = ContentBinding

	RequestBody, Err := json.Marshal(Options)

	if Err != nil {

		return "", time.Time{}, fmt.Errorf("failed to marshal request: %v", Err)

	}

	Req, Err := http.NewRequest("POST", g.ServerURL+"/get_pot", bytes.NewReader(RequestBody))

	if Err != nil {

		return "", time.Time{}, fmt.Errorf("failed to create request: %v", Err)

	}

	Req.Header.Set("Content-Type", "application/json")

	Resp, Err := g.HTTPClient.Do(Req)

	if Err != nil {

		return "", time.Time{}, fmt.Errorf("request to bgutil server failed: %v", Err)

	}

	defer Resp.Body.Close()

	Body, Err := io.ReadAll(Resp.Body)

	if Err != nil {

		return "", time.Time{}, fmt.Errorf("failed to read response: %v", Err)

	}

	var BgResp BgUtilResponse

	if Err := json.Unmarshal(Body, &BgResp); Err != nil {

		return "", time.Time{}, fmt.Errorf("failed to decode response: %v (body: %s)", Err, string(Body))

	}

	if BgResp.Error != "" {

		return "", time.Time{}, fmt.Errorf("bgutil error: %s", BgResp.Error)

	}

	if BgResp.PoToken == "" {

		return "", time.Time{}, fmt.Errorf("bgutil returned empty token")

	}

	// Use server's expiration or default to cache TTL

	ExpiresAt := BgResp.ExpiresAt

	if ExpiresAt.IsZero() {

		ExpiresAt = time.Now().Add(g.CacheTTL)

	}

	return BgResp.PoToken, ExpiresAt, nil

}

// InvalidateCache clears all cached tokens

func (g *BgUtilGenerator) InvalidateCache() {

	g.cacheLock.Lock()

	g.cache = make(map[string]*CachedToken)

	g.cacheLock.Unlock()

}

// InvalidateCacheFor removes a specific content binding from the cache

func (g *BgUtilGenerator) InvalidateCacheFor(ContentBinding string) {

	g.cacheLock.Lock()

	delete(g.cache, ContentBinding)

	g.cacheLock.Unlock()

}

// GetCachedToken returns a cached token if available (without generating)

func (g *BgUtilGenerator) GetCachedToken(ContentBinding string) (string, bool) {

	g.cacheLock.RLock()

	defer g.cacheLock.RUnlock()

	if Cached, Ok := g.cache[ContentBinding]; Ok {

		if time.Now().Before(Cached.ExpiresAt) {

			return Cached.Token, true

		}

	}

	return "", false

}

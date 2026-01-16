// Package pot provides PO (Proof of Origin) token generation for YouTube.
// PO tokens are required by YouTube to prevent 403 errors on video streams.
package pot

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

// DefaultServerURL is the default bgutil HTTP server URL
// Run: docker run -d -p 4416:4416 brainicism/bgutil-ytdlp-pot-provider
const DefaultServerURL = "http://127.0.0.1:4416"

// Provider generates PO tokens using a bgutil HTTP server
type Provider struct {
	serverURL  string
	httpClient *http.Client

	cache      map[string]*cachedToken
	cacheLock  sync.RWMutex
	cacheTTL   time.Duration
}

// cachedToken stores a token with its expiration
type cachedToken struct {
	Token     string
	ExpiresAt time.Time
}

// Request represents a request to the bgutil server
type Request struct {
	ContentBinding         string `json:"content_binding,omitempty"`
	Proxy                  string `json:"proxy,omitempty"`
	BypassCache            bool   `json:"bypass_cache,omitempty"`
	SourceAddress          string `json:"source_address,omitempty"`
	DisableTLSVerification bool   `json:"disable_tls_verification,omitempty"`
	DisableInnertube       bool   `json:"disable_innertube,omitempty"`
}

// Response represents the response from bgutil
type Response struct {
	PoToken        string    `json:"poToken"`
	ContentBinding string    `json:"contentBinding"`
	ExpiresAt      time.Time `json:"expiresAt"`
	Error          string    `json:"error,omitempty"`
}

// PingResponse represents the response from bgutil /ping endpoint
type PingResponse struct {
	ServerUptime float64 `json:"server_uptime"`
	Version      string  `json:"version"`
}

// NewProvider creates a new PO token provider
func NewProvider(serverURL string, httpClient *http.Client) *Provider {
	if serverURL == "" {
		serverURL = DefaultServerURL
	}

	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &Provider{
		serverURL:  serverURL,
		httpClient: httpClient,
		cache:      make(map[string]*cachedToken),
		cacheTTL:   5 * time.Hour,
	}
}

// IsAvailable checks if the bgutil server is reachable
func (p *Provider) IsAvailable() bool {
	_, err := p.Ping()
	return err == nil
}

// Ping checks if the bgutil server is running
func (p *Provider) Ping() (*PingResponse, error) {
	resp, err := p.httpClient.Get(p.serverURL + "/ping")
	if err != nil {
		return nil, fmt.Errorf("bgutil server unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bgutil server returned status %d", resp.StatusCode)
	}

	var pingResp PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pingResp); err != nil {
		return nil, fmt.Errorf("failed to decode ping response: %w", err)
	}

	return &pingResp, nil
}

// GetToken fetches a PO token for the given content binding
// Content binding is typically visitor_data for logged-out users
// or the session ID (first part of data_sync_id) for logged-in users
func (p *Provider) GetToken(contentBinding string) (string, error) {
	return p.GetTokenWithOptions(contentBinding, nil)
}

// GetTokenWithOptions fetches a PO token with custom options
func (p *Provider) GetTokenWithOptions(contentBinding string, opts *Request) (string, error) {
	// Check cache first
	p.cacheLock.RLock()
	if cached, ok := p.cache[contentBinding]; ok {
		if time.Now().Before(cached.ExpiresAt) {
			p.cacheLock.RUnlock()
			return cached.Token, nil
		}
	}
	p.cacheLock.RUnlock()

	// Generate new token
	token, expiresAt, err := p.generateToken(contentBinding, opts)
	if err != nil {
		return "", err
	}

	// Cache the token
	p.cacheLock.Lock()
	p.cache[contentBinding] = &cachedToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}
	p.cacheLock.Unlock()

	return token, nil
}

// GetGVSToken generates a GVS context PO token for video streaming
// Use visitor_data for logged-out users or data_sync_id for logged-in users
func (p *Provider) GetGVSToken(visitorData, dataSyncID string) (string, error) {
	contentBinding := visitorData

	// If logged in, use session ID from DataSyncID
	if dataSyncID != "" {
		contentBinding = extractSessionID(dataSyncID)
	}

	return p.GetToken(contentBinding)
}

// generateToken makes the actual HTTP request to the bgutil server
func (p *Provider) generateToken(contentBinding string, opts *Request) (string, time.Time, error) {
	if opts == nil {
		opts = &Request{}
	}
	opts.ContentBinding = contentBinding

	reqBody, err := json.Marshal(opts)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.serverURL+"/get_pot", bytes.NewReader(reqBody))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("request to bgutil server failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to read response: %w", err)
	}

	var bgResp Response
	if err := json.Unmarshal(body, &bgResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode response: %w (body: %s)", err, string(body))
	}

	if bgResp.Error != "" {
		return "", time.Time{}, fmt.Errorf("bgutil error: %s", bgResp.Error)
	}

	if bgResp.PoToken == "" {
		return "", time.Time{}, fmt.Errorf("bgutil returned empty token")
	}

	// Use expiry from response if available, otherwise use cache TTL
	expiresAt := bgResp.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(p.cacheTTL)
	}

	return bgResp.PoToken, expiresAt, nil
}

// ClearCache clears the token cache
func (p *Provider) ClearCache() {
	p.cacheLock.Lock()
	p.cache = make(map[string]*cachedToken)
	p.cacheLock.Unlock()
}

// extractSessionID extracts the session ID from a DataSyncID
// DataSyncID format is "SESSION_ID||..." - we need only the first part
func extractSessionID(dataSyncID string) string {
	if dataSyncID == "" {
		return ""
	}

	parts := strings.Split(dataSyncID, "||")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return dataSyncID
}

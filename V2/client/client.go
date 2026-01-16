package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/elucid503/overture-play/v2/auth"
	"github.com/elucid503/overture-play/v2/decipher"
	"github.com/elucid503/overture-play/v2/innertube"
	"github.com/elucid503/overture-play/v2/pot"
	"github.com/elucid503/overture-play/v2/types"
)

// Client is the main YouTube client for fetching video information
type Client struct {
	HTTPClient  *http.Client
	POTProvider *pot.Provider
	Decipherer  *decipher.Decipherer
	Auth        *auth.Auth

	Clients      []innertube.ClientConfig
	PlayerURL    string
	PlayerID     string
	PlayerCode   string
	VisitorData  string

	UserAgent   string
	AcceptLang  string
	Debug       bool
}

// NewClient creates a new YouTube client with default configuration
func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		POTProvider: pot.NewProvider("http://127.0.0.1:4416", nil),
		Clients:     innertube.DefaultClients(),

		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		AcceptLang: "en-US,en;q=0.9",
	}
}

// NewClientWithOptions creates a new YouTube client with custom options
func NewClientWithOptions(opts ClientOptions) *Client {
	c := NewClient()

	if opts.HTTPClient != nil {
		c.HTTPClient = opts.HTTPClient
	}
	if opts.POTServerURL != "" {
		c.POTProvider = pot.NewProvider(opts.POTServerURL, opts.HTTPClient)
	}
	if len(opts.Clients) > 0 {
		c.Clients = opts.Clients
	}
	if opts.UserAgent != "" {
		c.UserAgent = opts.UserAgent
	}
	if opts.AcceptLang != "" {
		c.AcceptLang = opts.AcceptLang
	}
	if opts.Auth != nil {
		c.Auth = opts.Auth
		// Update HTTP client with cookie jar if auth has cookies
		if jar, err := opts.Auth.GetCookieJar(); err == nil {
			c.HTTPClient.Jar = jar
		}
		// Use authenticated clients if auth is logged in
		if opts.Auth.IsLoggedIn() && len(opts.Clients) == 0 {
			c.Clients = innertube.DefaultAuthenticatedClients()
		}
	}
	if opts.CookieFile != "" {
		a, err := auth.NewAuthFromFile(opts.CookieFile)
		if err == nil {
			c.Auth = a
			if jar, err := a.GetCookieJar(); err == nil {
				c.HTTPClient.Jar = jar
			}
			if a.IsLoggedIn() && len(opts.Clients) == 0 {
				c.Clients = innertube.DefaultAuthenticatedClients()
			}
		}
	}
	if opts.CookieString != "" {
		c.Auth = auth.NewAuthFromString(opts.CookieString)
		if jar, err := c.Auth.GetCookieJar(); err == nil {
			c.HTTPClient.Jar = jar
		}
		// Use authenticated clients if auth is logged in
		if c.Auth.IsLoggedIn() && len(opts.Clients) == 0 {
			c.Clients = innertube.DefaultAuthenticatedClients()
		}
	}

	c.Debug = opts.Debug

	return c
}

// ClientOptions configures the YouTube client
type ClientOptions struct {
	HTTPClient   *http.Client
	POTServerURL string
	Clients      []innertube.ClientConfig
	UserAgent    string
	AcceptLang   string
	Debug        bool

	// Authentication options
	Auth         *auth.Auth   // Pre-configured auth
	CookieFile   string       // Path to Netscape cookie file
	CookieString string       // Cookie header string
}

// GetVideo fetches video information and formats
func (c *Client) GetVideo(videoID string) (*types.Video, error) {
	videoID = c.extractVideoID(videoID)
	if videoID == "" {
		return nil, fmt.Errorf("invalid video ID or URL")
	}

	// Fetch player info first
	if err := c.ensurePlayer(); err != nil {
		return nil, fmt.Errorf("failed to fetch player: %w", err)
	}

	// Try each client until one works
	var lastErr error
	for _, clientConfig := range c.Clients {
		video, err := c.fetchWithClient(videoID, clientConfig)
		if err == nil {
			return video, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("all clients failed, last error: %w", lastErr)
}

// extractVideoID extracts the video ID from a URL or returns as-is if already an ID
func (c *Client) extractVideoID(input string) string {
	input = strings.TrimSpace(input)

	// Already an ID (11 characters)
	if len(input) == 11 && regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`).MatchString(input) {
		return input
	}

	// Parse as URL
	patterns := []string{
		`(?:youtube\.com/watch\?v=|youtu\.be/)([a-zA-Z0-9_-]{11})`,
		`youtube\.com/embed/([a-zA-Z0-9_-]{11})`,
		`youtube\.com/v/([a-zA-Z0-9_-]{11})`,
		`youtube\.com/shorts/([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(input)
		if len(match) >= 2 {
			return match[1]
		}
	}

	return ""
}

// ensurePlayer fetches and caches the player script
func (c *Client) ensurePlayer() error {
	if c.Decipherer != nil {
		return nil
	}

	// Fetch player URL from YouTube page
	playerURL, err := c.fetchPlayerURL()
	if err != nil {
		return err
	}

	c.PlayerURL = playerURL
	c.PlayerID = decipher.ExtractPlayerID(playerURL)

	// Fetch player code
	playerCode, err := c.fetchPlayerCode(playerURL)
	if err != nil {
		return err
	}

	c.PlayerCode = playerCode

	// Create decipherer
	c.Decipherer, err = decipher.NewDecipherer(playerCode)
	if err != nil {
		return err
	}

	return nil
}

// fetchPlayerURL gets the current player URL from YouTube
// Uses a clean HTTP client without cookies to avoid auth-related redirects
func (c *Client) fetchPlayerURL() (string, error) {
	req, err := http.NewRequest("GET", "https://www.youtube.com/iframe_api", nil)
	if err != nil {
		return "", err
	}

	c.setBasicRequestHeaders(req)

	// Use a clean HTTP client without cookie jar to avoid auth redirects
	cleanClient := &http.Client{Timeout: c.HTTPClient.Timeout}
	resp, err := cleanClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	content := string(body)

	if c.Debug {
		fmt.Printf("[DEBUG] iframe_api status: %d, body length: %d\n", resp.StatusCode, len(content))
		if len(content) < 2000 {
			fmt.Printf("[DEBUG] iframe_api body: %s\n", content)
		} else {
			// Print first and last 500 chars to help debug
			fmt.Printf("[DEBUG] iframe_api body (first 500): %s\n", content[:500])
		}
	}

	// Extract player ID from iframe_api - format: \/s\/player\/XXXXXXXX\/ (escaped) or /s/player/XXXXXXXX/
	re := regexp.MustCompile(`\\?/s\\?/player\\?/([0-9a-fA-F]{8})\\?/`)
	match := re.FindStringSubmatch(content)
	if len(match) >= 2 {
		playerID := match[1]
		return fmt.Sprintf("https://www.youtube.com/s/player/%s/player_ias.vflset/en_US/base.js", playerID), nil
	}

	if c.Debug {
		fmt.Println("[DEBUG] iframe_api: player ID not found, falling back to page")
	}

	// Fallback: fetch from YouTube page
	return c.fetchPlayerURLFromPage()
}

// fetchPlayerURLFromPage extracts player URL from main YouTube page
// Uses a clean HTTP client without cookies to avoid auth-related redirects
func (c *Client) fetchPlayerURLFromPage() (string, error) {
	// Try a video watch page first - more reliable for extracting player
	req, err := http.NewRequest("GET", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", nil)
	if err != nil {
		return "", err
	}

	c.setBasicRequestHeaders(req)

	// Use a clean HTTP client without cookie jar to avoid auth redirects
	cleanClient := &http.Client{Timeout: c.HTTPClient.Timeout}
	resp, err := cleanClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	pageContent := string(body)

	if c.Debug {
		fmt.Printf("[DEBUG] watch page status: %d, body length: %d\n", resp.StatusCode, len(pageContent))
	}

	// Extract visitor_data from the page if we don't have it yet
	if c.VisitorData == "" {
		c.VisitorData = auth.ExtractVisitorDataFromHTML(pageContent)
	}

	// Look for player script URL with multiple patterns
	patterns := []string{
		// JSON format in ytInitialPlayerResponse - handles both player_ias and player_es6 formats
		`"jsUrl"\s*:\s*"(/s/player/[^"]+/player_(?:ias|es6)\.vflset/[^"]+/base\.js)"`,
		// PLAYER_JS_URL format
		`"PLAYER_JS_URL"\s*:\s*"(/s/player/[^"]+base\.js)"`,
		// Script tag src
		`<script[^>]+src="(/s/player/[^"]+/base\.js)"`,
		// Raw URL pattern - capture full path for both player_ias and player_es6 formats
		`(/s/player/[a-zA-Z0-9_-]+/player_(?:ias|es6)\.vflset/[a-zA-Z_]+/base\.js)`,
		// Alternative format with hash
		`/s/player/([a-fA-F0-9]{8})/`,
	}

	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(pageContent)
		if len(match) >= 2 {
			playerPath := match[1]
			// Last pattern returns just the hash, need to construct URL
			if i == len(patterns)-1 && !strings.Contains(playerPath, "/") {
				playerPath = fmt.Sprintf("/s/player/%s/player_ias.vflset/en_US/base.js", playerPath)
			}
			if !strings.HasPrefix(playerPath, "http") {
				playerPath = "https://www.youtube.com" + playerPath
			}
			if c.Debug {
				fmt.Printf("[DEBUG] watch page: found player URL with pattern %d: %s\n", i, playerPath)
			}
			return playerPath, nil
		}
	}

	if c.Debug {
		fmt.Println("[DEBUG] watch page: no patterns matched, falling back to embed")
	}

	// Final fallback: try embed page which is simpler
	return c.fetchPlayerURLFromEmbed()
}

// fetchPlayerURLFromEmbed extracts player URL from embed page
func (c *Client) fetchPlayerURLFromEmbed() (string, error) {
	req, err := http.NewRequest("GET", "https://www.youtube.com/embed/dQw4w9WgXcQ", nil)
	if err != nil {
		return "", err
	}

	c.setBasicRequestHeaders(req)

	// Use a clean HTTP client without cookie jar to avoid auth redirects
	cleanClient := &http.Client{Timeout: c.HTTPClient.Timeout}
	resp, err := cleanClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	embedContent := string(body)

	if c.Debug {
		fmt.Printf("[DEBUG] embed page status: %d, body length: %d\n", resp.StatusCode, len(embedContent))
	}

	// Embed page uses simpler patterns
	patterns := []string{
		`"jsUrl"\s*:\s*"([^"]+base\.js)"`,
		`"PLAYER_JS_URL"\s*:\s*"([^"]+)"`,
		`/s/player/([a-fA-F0-9]{8})/`,
	}

	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(embedContent)
		if len(match) >= 2 {
			playerPath := match[1]
			if i == len(patterns)-1 && !strings.Contains(playerPath, "/") {
				playerPath = fmt.Sprintf("/s/player/%s/player_ias.vflset/en_US/base.js", playerPath)
			}
			if !strings.HasPrefix(playerPath, "http") {
				playerPath = "https://www.youtube.com" + playerPath
			}
			if c.Debug {
				fmt.Printf("[DEBUG] embed page: found player URL with pattern %d: %s\n", i, playerPath)
			}
			return playerPath, nil
		}
	}

	if c.Debug {
		fmt.Println("[DEBUG] embed page: no patterns matched")
	}

	return "", fmt.Errorf("player URL not found")
}

// fetchPlayerCode downloads the player JavaScript code
func (c *Client) fetchPlayerCode(playerURL string) (string, error) {
	req, err := http.NewRequest("GET", playerURL, nil)
	if err != nil {
		return "", err
	}

	c.setBasicRequestHeaders(req)

	// Use a clean HTTP client without cookie jar to avoid auth redirects
	cleanClient := &http.Client{Timeout: c.HTTPClient.Timeout}
	resp, err := cleanClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// fetchWithClient fetches video info using a specific innertube client
func (c *Client) fetchWithClient(videoID string, clientConfig innertube.ClientConfig) (*types.Video, error) {
	// Get context with visitor data if available
	var ctx innertube.InnertubeContext
	visitorData := c.getVisitorData()
	if visitorData != "" {
		ctx = clientConfig.GetContextWithVisitor(visitorData)
	} else {
		ctx = clientConfig.GetContext()
	}

	sts := 0
	if c.Decipherer != nil {
		sts = c.Decipherer.GetSignatureTimestamp()
	}

	payload := map[string]interface{}{
		"context": ctx,
		"videoId": videoID,
		"playbackContext": map[string]interface{}{
			"contentPlaybackContext": map[string]interface{}{
				"signatureTimestamp": sts,
				"html5Preference":    "HTML5_PREF_WANTS",
			},
		},
		"racyCheckOk":    true,
		"contentCheckOk": true,
	}

	// Add Player PO token if required for this client
	playerPOToken, err := c.getPlayerPOToken(videoID, clientConfig)
	if err == nil && playerPOToken != "" {
		payload["serviceIntegrityDimensions"] = map[string]string{
			"poToken": playerPOToken,
		}
	}

	// Make player API request - no API key needed for modern clients
	apiURL := "https://www.youtube.com/youtubei/v1/player?prettyPrint=false"

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, err
	}

	c.setAPIRequestHeaders(req, clientConfig)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response
	return c.parsePlayerResponse(body, clientConfig, videoID)
}

// getPlayerPOToken gets a PO token for the player API request (bound to video ID)
func (c *Client) getPlayerPOToken(videoID string, clientConfig innertube.ClientConfig) (string, error) {
	// Check if client requires PO token for player
	if !clientConfig.RequiresPoToken() {
		return "", nil
	}

	// Try to get PO token from provider
	if c.POTProvider == nil {
		return "", nil
	}

	if !c.POTProvider.IsAvailable() {
		return "", nil
	}

	// Player PO token is bound to video ID
	return c.POTProvider.GetToken(videoID)
}

// getGVSPOToken gets a GVS PO token for stream URLs (bound to visitor_data or data_sync_id)
func (c *Client) getGVSPOToken(videoID string, clientConfig innertube.ClientConfig) (string, error) {
	// Check if client requires PO token for GVS
	if len(clientConfig.GVSPoTokenPolicies) == 0 {
		return "", nil
	}

	// Try to get PO token from provider
	if c.POTProvider == nil {
		return "", nil
	}

	if !c.POTProvider.IsAvailable() {
		return "", nil
	}

	// GVS PO token is bound to visitor_data (unauthenticated) or data_sync_id (authenticated)
	visitorData := c.getVisitorData()
	dataSyncID := ""
	if c.Auth != nil && c.Auth.IsLoggedIn() {
		dataSyncID = c.Auth.GetDataSyncID()
	}

	return c.POTProvider.GetGVSToken(visitorData, dataSyncID)
}

// parsePlayerResponse parses the player API response
func (c *Client) parsePlayerResponse(data []byte, clientConfig innertube.ClientConfig, videoID string) (*types.Video, error) {
	var resp PlayerResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for playability errors
	if resp.PlayabilityStatus.Status != "OK" {
		return nil, fmt.Errorf("video not playable: %s - %s",
			resp.PlayabilityStatus.Status,
			resp.PlayabilityStatus.Reason)
	}

	// Get GVS PO token for stream URLs (bound to visitor_data or data_sync_id)
	gvsPOToken, _ := c.getGVSPOToken(videoID, clientConfig)

	video := &types.Video{
		ID: resp.VideoDetails.VideoID,

		Title:       resp.VideoDetails.Title,
		Description: resp.VideoDetails.ShortDescription,
		Author:      resp.VideoDetails.Author,
		ChannelID:   resp.VideoDetails.ChannelID,

		Duration: c.parseDuration(resp.VideoDetails.LengthSeconds),

		ViewCount: c.parseInt(resp.VideoDetails.ViewCount),
		IsLive:    resp.VideoDetails.IsLiveContent,
		IsPrivate: resp.VideoDetails.IsPrivate,

		Formats:    make([]types.Format, 0),
		Thumbnails: c.parseThumbnails(resp.VideoDetails.Thumbnail),
	}

	// Parse formats
	allFormats := append(resp.StreamingData.Formats, resp.StreamingData.AdaptiveFormats...)
	for _, sf := range allFormats {
		format, err := c.parseFormat(sf, gvsPOToken)
		if err != nil {
			continue
		}
		video.Formats = append(video.Formats, format)
	}

	return video, nil
}

// parseFormat parses a streaming format and deciphers URLs
func (c *Client) parseFormat(sf StreamingFormat, gvsPOToken string) (types.Format, error) {
	format := types.Format{
		ITag:     sf.ITag,
		MimeType: sf.MimeType,

		Bitrate:        sf.Bitrate,
		AverageBitrate: sf.AverageBitrate,
		ContentLength:  c.parseInt(sf.ContentLength),

		Width:        sf.Width,
		Height:       sf.Height,
		FPS:          sf.FPS,
		Quality:      sf.Quality,
		QualityLabel: sf.QualityLabel,

		AudioQuality:    sf.AudioQuality,
		AudioChannels:   sf.AudioChannels,
		AudioSampleRate: c.parseInt(sf.AudioSampleRate),

		IndexRange: c.parseRange(sf.IndexRange),
		InitRange:  c.parseRange(sf.InitRange),
	}

	// Get URL
	var streamURL string
	if sf.URL != "" {
		streamURL = sf.URL
	} else if sf.SignatureCipher != "" {
		// Decipher the URL
		deciphered, err := c.decipherURL(sf.SignatureCipher)
		if err != nil {
			return format, err
		}
		streamURL = deciphered
	} else {
		return format, fmt.Errorf("no URL available for format %d", sf.ITag)
	}

	// Process n-parameter
	streamURL = c.processNParameter(streamURL)

	// Add GVS PO token if available
	if gvsPOToken != "" {
		streamURL = c.addPOTokenToURL(streamURL, gvsPOToken)
	}

	format.URL = streamURL
	return format, nil
}

// addPOTokenToURL adds a PO token to the stream URL
func (c *Client) addPOTokenToURL(streamURL string, poToken string) string {
	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		return streamURL
	}

	q := parsedURL.Query()
	q.Set("pot", poToken)
	parsedURL.RawQuery = q.Encode()
	return parsedURL.String()
}

// decipherURL deciphers a signature cipher
func (c *Client) decipherURL(signatureCipher string) (string, error) {
	params, err := url.ParseQuery(signatureCipher)
	if err != nil {
		return "", err
	}

	streamURL := params.Get("url")
	signature := params.Get("s")
	signatureParam := params.Get("sp")
	if signatureParam == "" {
		signatureParam = "sig"
	}

	if signature != "" && c.Decipherer != nil {
		deciphered := c.Decipherer.DecipherSignature(signature)

		parsedURL, err := url.Parse(streamURL)
		if err != nil {
			return "", err
		}

		q := parsedURL.Query()
		q.Set(signatureParam, deciphered)
		parsedURL.RawQuery = q.Encode()
		streamURL = parsedURL.String()
	}

	return streamURL, nil
}

// processNParameter processes and solves the n-parameter challenge
func (c *Client) processNParameter(streamURL string) string {
	if c.Decipherer == nil {
		return streamURL
	}

	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		return streamURL
	}

	q := parsedURL.Query()
	n := q.Get("n")
	if n == "" {
		return streamURL
	}

	solved, err := c.Decipherer.SolveNChallenge(n)
	if err != nil || solved == n {
		return streamURL
	}

	q.Set("n", solved)
	parsedURL.RawQuery = q.Encode()
	return parsedURL.String()
}

// parseThumbnails parses thumbnail data
func (c *Client) parseThumbnails(data ThumbnailContainer) []types.Thumbnail {
	thumbnails := make([]types.Thumbnail, 0, len(data.Thumbnails))
	for _, t := range data.Thumbnails {
		thumbnails = append(thumbnails, types.Thumbnail{
			URL:    t.URL,
			Width:  t.Width,
			Height: t.Height,
		})
	}
	return thumbnails
}

// parseRange parses a range object
func (c *Client) parseRange(r *RangeData) *types.Range {
	if r == nil {
		return nil
	}
	return &types.Range{
		Start: c.parseInt(r.Start),
		End:   c.parseInt(r.End),
	}
}

// parseDuration parses duration string to seconds
func (c *Client) parseDuration(s string) int {
	return c.parseInt(s)
}

// parseInt parses an integer string
func (c *Client) parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

// setRequestHeaders sets standard request headers
func (c *Client) setRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept-Language", c.AcceptLang)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Add cookie header if auth is configured
	if c.Auth != nil {
		req.Header.Set("Cookie", c.Auth.GetCookieHeader())
	}
}

// setBasicRequestHeaders sets headers without authentication cookies
// Used for fetching player URLs where auth cookies can cause redirects to login pages
func (c *Client) setBasicRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept-Language", c.AcceptLang)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
}

// setAPIRequestHeaders sets headers for API requests
func (c *Client) setAPIRequestHeaders(req *http.Request, clientConfig innertube.ClientConfig) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-YouTube-Client-Name", fmt.Sprintf("%d", clientConfig.ContextName))
	req.Header.Set("X-YouTube-Client-Version", clientConfig.Version)
	req.Header.Set("Origin", "https://www.youtube.com")
	req.Header.Set("Referer", "https://www.youtube.com/")

	if clientConfig.UserAgent != "" {
		req.Header.Set("User-Agent", clientConfig.UserAgent)
	} else {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	// Add visitor ID header
	visitorData := c.getVisitorData()
	if visitorData != "" {
		req.Header.Set("X-Goog-Visitor-Id", visitorData)
	}

	// Add authentication headers
	if c.Auth != nil {
		// Add cookie header
		req.Header.Set("Cookie", c.Auth.GetCookieHeader())

		// Add SAPISIDHASH for authenticated requests
		if c.Auth.IsLoggedIn() {
			sapisidhash := c.Auth.GetSAPISIDHash("https://www.youtube.com")
			if sapisidhash != "" {
				req.Header.Set("Authorization", sapisidhash)
				req.Header.Set("X-Origin", "https://www.youtube.com")
			}
		}
	}
}

// getVisitorData returns the visitor data for API requests
func (c *Client) getVisitorData() string {
	if c.VisitorData != "" {
		return c.VisitorData
	}
	if c.Auth != nil {
		return c.Auth.GetVisitorData()
	}
	return ""
}

// SetVisitorData sets the visitor data for API requests
func (c *Client) SetVisitorData(visitorData string) {
	c.VisitorData = visitorData
}

// IsAuthenticated returns true if the client has valid authentication
func (c *Client) IsAuthenticated() bool {
	return c.Auth != nil && c.Auth.IsLoggedIn()
}

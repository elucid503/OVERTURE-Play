// Package auth provides authentication support for YouTube via cookies
package auth

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// YouTubeURL is the base YouTube URL for cookies
const YouTubeURL = "https://www.youtube.com"

// Auth represents authenticated YouTube session data
type Auth struct {
	Cookies     []*http.Cookie
	VisitorData string
	DataSyncID  string
	SessionID   string
	SAPISID     string
}

// CookieFile represents a Netscape cookie file entry
type CookieFile struct {
	Domain   string
	HostOnly bool
	Path     string
	Secure   bool
	Expires  time.Time
	Name     string
	Value    string
}

// NewAuth creates a new Auth from cookies
func NewAuth(cookies []*http.Cookie) *Auth {
	auth := &Auth{
		Cookies: cookies,
	}
	auth.extractAuthData()
	return auth
}

// NewAuthFromFile loads cookies from a Netscape cookie file
func NewAuthFromFile(path string) (*Auth, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open cookie file: %w", err)
	}
	defer file.Close()

	var cookies []*http.Cookie
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			continue
		}

		domain := parts[0]
		// hostOnly := parts[1] == "FALSE"
		path := parts[2]
		secure := parts[3] == "TRUE"

		// Parse expiry
		var expiryTime time.Time
		var expiryInt int64
		fmt.Sscanf(parts[4], "%d", &expiryInt)
		if expiryInt > 0 {
			expiryTime = time.Unix(expiryInt, 0)
		}

		name := parts[5]
		value := parts[6]

		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			Secure:   secure,
			Expires:  expiryTime,
			HttpOnly: true,
		}

		cookies = append(cookies, cookie)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read cookie file: %w", err)
	}

	return NewAuth(cookies), nil
}

// NewAuthFromJSON loads cookies from a JSON file (exported from browser extension)
func NewAuthFromJSON(path string) (*Auth, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON cookie file: %w", err)
	}

	// Browser extension format
	type JSONCookie struct {
		Name     string  `json:"name"`
		Value    string  `json:"value"`
		Domain   string  `json:"domain"`
		Path     string  `json:"path"`
		Secure   bool    `json:"secure"`
		HttpOnly bool    `json:"httpOnly"`
		Expires  float64 `json:"expirationDate"`
	}

	var jsonCookies []JSONCookie
	if err := json.Unmarshal(data, &jsonCookies); err != nil {
		return nil, fmt.Errorf("failed to parse JSON cookies: %w", err)
	}

	var cookies []*http.Cookie
	for _, jc := range jsonCookies {
		var expiryTime time.Time
		if jc.Expires > 0 {
			expiryTime = time.Unix(int64(jc.Expires), 0)
		}

		cookie := &http.Cookie{
			Name:     jc.Name,
			Value:    jc.Value,
			Domain:   jc.Domain,
			Path:     jc.Path,
			Secure:   jc.Secure,
			HttpOnly: jc.HttpOnly,
			Expires:  expiryTime,
		}
		cookies = append(cookies, cookie)
	}

	return NewAuth(cookies), nil
}

// NewAuthFromString parses cookies from a Cookie header string
func NewAuthFromString(cookieHeader string) *Auth {
	var cookies []*http.Cookie

	// Parse "name=value; name2=value2" format
	pairs := strings.Split(cookieHeader, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		idx := strings.Index(pair, "=")
		if idx < 0 {
			continue
		}

		name := strings.TrimSpace(pair[:idx])
		value := strings.TrimSpace(pair[idx+1:])

		cookie := &http.Cookie{
			Name:   name,
			Value:  value,
			Domain: ".youtube.com",
			Path:   "/",
		}

		// Set Secure flag for __Secure- prefixed cookies (required by browsers)
		if strings.HasPrefix(name, "__Secure-") {
			cookie.Secure = true
		}

		cookies = append(cookies, cookie)
	}

	return NewAuth(cookies)
}

// extractAuthData extracts visitor data and other auth info from cookies
func (a *Auth) extractAuthData() {
	for _, cookie := range a.Cookies {
		switch cookie.Name {
		case "VISITOR_INFO1_LIVE":
			a.VisitorData = cookie.Value
		case "__Secure-3PAPISID", "SAPISID":
			a.SAPISID = cookie.Value
		case "__Secure-3PSID", "SID":
			// SID cookie indicates logged in
		}
	}
}

// IsLoggedIn returns true if the auth has valid login cookies
func (a *Auth) IsLoggedIn() bool {
	hasLoginCookie := false
	for _, cookie := range a.Cookies {
		if cookie.Name == "__Secure-3PSID" || cookie.Name == "SID" {
			hasLoginCookie = true
			break
		}
	}
	return hasLoginCookie
}

// GetCookieJar creates an http.CookieJar with the auth cookies
func (a *Auth) GetCookieJar() (http.CookieJar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	ytURL, _ := url.Parse(YouTubeURL)
	jar.SetCookies(ytURL, a.Cookies)

	return jar, nil
}

// GetVisitorData returns the visitor data from cookies or fetches it
func (a *Auth) GetVisitorData() string {
	return a.VisitorData
}

// GetDataSyncID returns the data sync ID for PO token generation
func (a *Auth) GetDataSyncID() string {
	return a.DataSyncID
}

// SetDataSyncID sets the data sync ID extracted from API responses
func (a *Auth) SetDataSyncID(dataSyncID string) {
	a.DataSyncID = dataSyncID
	// Extract session ID from data sync ID
	if dataSyncID != "" {
		parts := strings.Split(dataSyncID, "||")
		if len(parts) > 0 && parts[0] != "" {
			a.SessionID = parts[0]
		}
	}
}

// GetSessionID returns the session ID for PO token generation
func (a *Auth) GetSessionID() string {
	return a.SessionID
}

// GetSAPISIDHash generates the SAPISIDHASH authorization header
func (a *Auth) GetSAPISIDHash(origin string) string {
	if a.SAPISID == "" {
		return ""
	}

	timestamp := time.Now().Unix()
	input := fmt.Sprintf("%d %s %s", timestamp, a.SAPISID, origin)

	// SHA1 hash
	// Note: In production, use crypto/sha1
	hash := sha1Hash(input)

	return fmt.Sprintf("SAPISIDHASH %d_%s", timestamp, hash)
}

// GetCookieHeader returns the cookies as a Cookie header string
func (a *Auth) GetCookieHeader() string {
	var pairs []string
	for _, cookie := range a.Cookies {
		pairs = append(pairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	return strings.Join(pairs, "; ")
}

// sha1Hash computes SHA1 hash of a string
func sha1Hash(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

// ExtractVisitorDataFromHTML extracts visitor_data from YouTube HTML page
func ExtractVisitorDataFromHTML(html string) string {
	patterns := []string{
		`"VISITOR_DATA"\s*:\s*"([^"]+)"`,
		`ytcfg\.set\s*\(\s*\{[^}]*"VISITOR_DATA"\s*:\s*"([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(html)
		if len(match) >= 2 {
			return match[1]
		}
	}

	return ""
}

// ExtractDataSyncIDFromResponse extracts data_sync_id from API response
func ExtractDataSyncIDFromResponse(data []byte) string {
	// Look for dataSyncId in response
	pattern := regexp.MustCompile(`"dataSyncId"\s*:\s*"([^"]+)"`)
	match := pattern.FindSubmatch(data)
	if len(match) >= 2 {
		return string(match[1])
	}
	return ""
}

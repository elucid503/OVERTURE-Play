package POToken

import (
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
)

// PoTokenContext represents the context in which a PO Token is used
// Following yt-dlp's implementation: GVS (Google Video Server), PLAYER, SUBS

type PoTokenContext string

const (

	ContextGVS    PoTokenContext = "gvs"
	ContextPlayer PoTokenContext = "player"
	ContextSubs   PoTokenContext = "subs"

)

// ContentBindingType represents what the PO token is bound to

type ContentBindingType string

const (

	BindingVisitorData ContentBindingType = "visitor_data"
	BindingDataSyncID  ContentBindingType = "datasync_id"
	BindingVideoID     ContentBindingType = "video_id"
	BindingVisitorID   ContentBindingType = "visitor_id"

)

// PoTokenRequest contains parameters needed to generate or fetch a PO Token

type PoTokenRequest struct {

	Context      PoTokenContext
	VisitorData  string
	DataSyncID   string
	VideoID      string
	ClientName   string
	PlayerURL    string
	Authenticated bool

}

// PoTokenResponse contains the PO Token and optional expiration

type PoTokenResponse struct {

	PoToken   string
	ExpiresAt int64

}

// GetContentBinding returns the appropriate content binding value for a PO Token request
// Based on yt-dlp's get_webpo_content_binding implementation

func GetContentBinding(Request *PoTokenRequest) (string, ContentBindingType) {

	if Request == nil {

		return "", ""

	}

	// For GVS context, bind to visitor_data (logged out) or data_sync_id (logged in)

	if Request.Context == ContextGVS {

		if Request.Authenticated && Request.DataSyncID != "" {

			return Request.DataSyncID, BindingDataSyncID

		}

		if Request.VisitorData != "" {

			return Request.VisitorData, BindingVisitorData

		}

	}

	// For Player and Subs context, bind to video_id

	if Request.Context == ContextPlayer || Request.Context == ContextSubs {

		return Request.VideoID, BindingVideoID

	}

	return "", ""

}

// ExtractVisitorID attempts to extract the visitor ID from visitor_data protobuf
// Based on yt-dlp's _extract_visitor_id implementation

func ExtractVisitorID(VisitorData string) string {

	if VisitorData == "" {

		return ""

	}

	// URL decode the visitor data first

	Decoded, Err := url.QueryUnescape(VisitorData)

	if Err != nil {

		return ""

	}

	// Base64 URL decode

	DecodedBytes, Err := base64.URLEncoding.DecodeString(Decoded)

	if Err != nil {

		// Try without padding

		DecodedBytes, Err = base64.RawURLEncoding.DecodeString(Decoded)

		if Err != nil {

			return ""

		}

	}

	// Extract bytes 2-13 which should contain the visitor ID

	if len(DecodedBytes) < 13 {

		return ""

	}

	VisitorID := string(DecodedBytes[2:13])

	// Validate the visitor ID format (11 alphanumeric characters)

	ValidIDRegex := regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

	if ValidIDRegex.MatchString(VisitorID) {

		return VisitorID

	}

	return ""

}

// CleanPoToken validates and normalizes a PO Token
// Ensures the token is properly base64url encoded

func CleanPoToken(Token string) string {

	if Token == "" {

		return ""

	}

	// URL decode first

	Decoded, Err := url.QueryUnescape(Token)

	if Err != nil {

		return Token

	}

	// Try to decode and re-encode to normalize

	DecodedBytes, Err := base64.URLEncoding.DecodeString(Decoded)

	if Err != nil {

		DecodedBytes, Err = base64.RawURLEncoding.DecodeString(Decoded)

		if Err != nil {

			return Token

		}

	}

	// Re-encode with standard base64url encoding (no padding for URL safety)

	return base64.RawURLEncoding.EncodeToString(DecodedBytes)

}

// ApplyToHLSManifestURL appends a PO Token to an HLS manifest URL
// Following yt-dlp pattern: manifest_url.rstrip('/') + f'/pot/{po_token}'

func ApplyToHLSManifestURL(ManifestURL string, PoToken string) string {

	if PoToken == "" {

		return ManifestURL

	}

	// Remove trailing slash

	ManifestURL = strings.TrimRight(ManifestURL, "/")

	// Append /pot/<token>

	return ManifestURL + "/pot/" + PoToken

}

// ApplyToDASHManifestURL appends a PO Token to a DASH manifest URL
// Same pattern as HLS

func ApplyToDASHManifestURL(ManifestURL string, PoToken string) string {

	return ApplyToHLSManifestURL(ManifestURL, PoToken)

}

// ApplyToSegmentURL appends a PO Token to a segment URL as a query parameter
// Following yt-dlp pattern: update_url_query(fmt_url, {'pot': po_token})

func ApplyToSegmentURL(SegmentURL string, PoToken string) string {

	if PoToken == "" {

		return SegmentURL

	}

	ParsedURL, Err := url.Parse(SegmentURL)

	if Err != nil {

		return SegmentURL

	}

	Query := ParsedURL.Query()
	Query.Set("pot", PoToken)
	ParsedURL.RawQuery = Query.Encode()

	return ParsedURL.String()

}

// IsPoTokenRequired determines if a PO Token is required based on the streaming protocol
// Based on yt-dlp's WEB_PO_TOKEN_POLICIES

func IsPoTokenRequired(Protocol string, IsHLS bool, IsPremium bool) bool {

	// Premium users generally don't require PO tokens (except for subtitles)

	if IsPremium {

		return false

	}

	// For HTTPS direct streaming and DASH, PO tokens are required

	if Protocol == "https" || Protocol == "dash" {

		return true

	}

	// For HLS, PO tokens are recommended but not strictly required

	if IsHLS || Protocol == "hls" {

		return false // Recommended but not required

	}

	return true

}

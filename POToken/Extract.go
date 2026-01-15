package POToken

import (
	"regexp"
	"strings"
)

// ExtractVisitorData extracts visitor_data from YouTube API response or ytcfg
// Following yt-dlp's _extract_visitor_data implementation
// Searches multiple possible locations in the response JSON

func ExtractVisitorData(Response map[string]interface{}) string {

	if Response == nil {

		return ""

	}

	// Check VISITOR_DATA key (from ytcfg)

	if VisitorData, Ok := Response["VISITOR_DATA"].(string); Ok && VisitorData != "" {

		return VisitorData

	}

	// Check INNERTUBE_CONTEXT.client.visitorData

	if InnertubeContext, Ok := Response["INNERTUBE_CONTEXT"].(map[string]interface{}); Ok {

		if Client, Ok := InnertubeContext["client"].(map[string]interface{}); Ok {

			if VisitorData, Ok := Client["visitorData"].(string); Ok && VisitorData != "" {

				return VisitorData

			}

		}

	}

	// Check responseContext.visitorData (from API response)

	if ResponseContext, Ok := Response["responseContext"].(map[string]interface{}); Ok {

		if VisitorData, Ok := ResponseContext["visitorData"].(string); Ok && VisitorData != "" {

			return VisitorData

		}

	}

	return ""

}

// ExtractDataSyncID extracts the data sync ID for authenticated users
// Following yt-dlp's _extract_data_sync_id implementation

func ExtractDataSyncID(Response map[string]interface{}) string {

	if Response == nil {

		return ""

	}

	// Check DATASYNC_ID key (from ytcfg)

	if DataSyncID, Ok := Response["DATASYNC_ID"].(string); Ok && DataSyncID != "" {

		return DataSyncID

	}

	// Check responseContext.mainAppWebResponseContext.datasyncId

	if ResponseContext, Ok := Response["responseContext"].(map[string]interface{}); Ok {

		if MainAppWebResponse, Ok := ResponseContext["mainAppWebResponseContext"].(map[string]interface{}); Ok {

			if DataSyncID, Ok := MainAppWebResponse["datasyncId"].(string); Ok && DataSyncID != "" {

				return DataSyncID

			}

		}

	}

	return ""

}

// ExtractVisitorDataFromHTML extracts visitor_data from YouTube page HTML
// Searches for VISITOR_DATA in ytcfg.set calls

func ExtractVisitorDataFromHTML(HTML string) string {

	if HTML == "" {

		return ""

	}

	// Match ytcfg.set containing VISITOR_DATA

	VisitorDataRegex := regexp.MustCompile(`"VISITOR_DATA"\s*:\s*"([^"]+)"`)

	Matches := VisitorDataRegex.FindStringSubmatch(HTML)

	if len(Matches) >= 2 {

		return Matches[1]

	}

	// Alternative pattern

	AltRegex := regexp.MustCompile(`VISITOR_DATA['"]?\s*[:=]\s*['"]([^'"]+)['"]`)

	Matches = AltRegex.FindStringSubmatch(HTML)

	if len(Matches) >= 2 {

		return Matches[1]

	}

	return ""

}

// ExtractFromCookie extracts visitor data from the VISITOR_INFO1_LIVE cookie

func ExtractFromCookie(CookieString string) string {

	if CookieString == "" {

		return ""

	}

	// Parse cookies

	Cookies := strings.Split(CookieString, ";")

	for _, Cookie := range Cookies {

		Cookie = strings.TrimSpace(Cookie)

		if strings.HasPrefix(Cookie, "VISITOR_INFO1_LIVE=") {

			Parts := strings.SplitN(Cookie, "=", 2)

			if len(Parts) == 2 {

				return Parts[1]

			}

		}

	}

	return ""

}

// PoTokenProvider interface for implementing different PO Token providers
// Matches yt-dlp's PoTokenProvider pattern

type PoTokenProvider interface {

	// Name returns the provider name

	Name() string

	// IsAvailable checks if the provider can be used

	IsAvailable() bool

	// RequestPoToken fetches a PO Token for the given request

	RequestPoToken(Request *PoTokenRequest) (*PoTokenResponse, error)

}

// DefaultProvider is a simple provider that uses pre-configured tokens
// For users who have tokens from external sources (e.g., browser extraction)

type DefaultProvider struct {

	ConfiguredToken string

}

// Name returns the provider name

func (p *DefaultProvider) Name() string {

	return "default"

}

// IsAvailable checks if a token is configured

func (p *DefaultProvider) IsAvailable() bool {

	return p.ConfiguredToken != ""

}

// RequestPoToken returns the configured token

func (p *DefaultProvider) RequestPoToken(Request *PoTokenRequest) (*PoTokenResponse, error) {

	if p.ConfiguredToken == "" {

		return nil, nil

	}

	return &PoTokenResponse{

		PoToken:   CleanPoToken(p.ConfiguredToken),
		ExpiresAt: 0, // No expiration for manually configured tokens

	}, nil

}

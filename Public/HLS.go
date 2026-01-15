package Public

import (
	"fmt"
	"strings"

	"github.com/elucid503/Overture-Play/Config"
	"github.com/elucid503/Overture-Play/Functions"
	"github.com/elucid503/Overture-Play/POToken"
	"github.com/elucid503/Overture-Play/Structs"
)

// HLSOptions configures HLS manifest and playlist fetching

type HLSOptions struct {

	Proxy     *Proxy
	UserAgent string

	// PO Token for authenticated streaming (required for many videos)
	// This should be a base64url-encoded PO Token
	// If empty but Generator is set, will auto-generate using VisitorData

	PoToken string

	// Generator for automatic PO token generation
	// Set this to a *POToken.BgUtilGenerator to enable auto-generation
	// Requires running the bgutil-ytdlp-pot-provider server

	Generator *POToken.BgUtilGenerator

	// VisitorData for automatic PO token generation
	// This is typically obtained from the video info response
	// Required if using Generator and PoToken is empty

	VisitorData string

	// DataSyncID for authenticated users (optional)
	// If set along with IsAuthenticated, will be used instead of VisitorData

	DataSyncID string

	// IsAuthenticated indicates if the user is logged in
	// Affects which content binding is used for token generation

	IsAuthenticated bool

}

// getOrGeneratePoToken returns the PO token from options, generating one if needed

// extractSessionID extracts the account Session ID from a DataSyncID
// DataSyncID format is "SESSION_ID||..." - we need only the first part

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

func getOrGeneratePoToken(Options *HLSOptions) (string, error) {

	// If token already provided, use it

	if Options.PoToken != "" {

		return Options.PoToken, nil

	}

	// If no generator, can't auto-generate

	if Options.Generator == nil {

		return "", nil

	}

	// Need content binding for generation
	// For GVS (video streaming), use:
	// - Session ID (first part of DataSyncID) for logged-in users
	// - VisitorData for logged-out users

	ContentBinding := Options.VisitorData

	if Options.IsAuthenticated && Options.DataSyncID != "" {

		// Extract just the session ID from DataSyncID (format: "SESSION_ID||...")

		ContentBinding = extractSessionID(Options.DataSyncID)

	}

	if ContentBinding == "" {

		return "", fmt.Errorf("cannot generate PO token: no VisitorData or DataSyncID provided")

	}

	// Generate token

	Token, Err := Options.Generator.GetPoToken(ContentBinding, nil)

	if Err != nil {

		return "", fmt.Errorf("failed to generate PO token: %v", Err)

	}

	return Token, nil

}

// GetHLSManifest fetches and decodes an HLS master manifest, returning playlists and audio groups
// If a PoToken is provided in Options, it will be appended to the manifest URL
// If Generator and VisitorData are set but no PoToken, will auto-generate

func GetHLSManifest(ManifestURL string, Options *HLSOptions) (*Structs.HLSManifest, error) {

	if Options == nil {

		Options = &HLSOptions{

			UserAgent: Config.Current.GetInnertubeClient().UserAgent,

		}

	}

	if Options.UserAgent == "" {

		Options.UserAgent = Config.Current.GetInnertubeClient().UserAgent

	}

	// Get or generate PO token

	PoToken, Err := getOrGeneratePoToken(Options)

	if Err != nil {

		return nil, fmt.Errorf("failed to get PO token: %v", Err)

	}

	// Apply PO Token to manifest URL if available

	if PoToken != "" {

		ManifestURL = POToken.ApplyToHLSManifestURL(ManifestURL, PoToken)

	}

	var ProxyStruct *Structs.Proxy

	if Options.Proxy != nil {

		ProxyStruct = &Structs.Proxy{

			Host:     Options.Proxy.Host,
			Port:     Options.Proxy.Port,
			UserPass: Options.Proxy.UserPass,

		}

	}

	Content, Err := Functions.FetchHLSContent(ManifestURL, ProxyStruct, Options.UserAgent)

	if Err != nil {

		return nil, fmt.Errorf("failed to fetch HLS manifest: %v", Err)

	}

	Manifest := Functions.ParseHLSManifest(Content, ManifestURL)

	return Manifest, nil

}

// GetHLSPlaylist fetches and decodes an HLS media playlist (with segments) from a playlist URI
// The playlist URI should already contain the PO Token if it was passed to GetHLSManifest
// since child playlist URLs inherit the token from the manifest
// If auto-generation is configured, will generate token if needed

func GetHLSPlaylist(PlaylistURI string, Options *HLSOptions) (*Structs.HLSMediaPlaylist, error) {

	if Options == nil {

		Options = &HLSOptions{

			UserAgent: Config.Current.GetInnertubeClient().UserAgent,

		}

	}

	if Options.UserAgent == "" {

		Options.UserAgent = Config.Current.GetInnertubeClient().UserAgent

	}

	// Get or generate PO token

	PoToken, Err := getOrGeneratePoToken(Options)

	if Err != nil {

		return nil, fmt.Errorf("failed to get PO token: %v", Err)

	}

	// Apply PO Token to playlist URL if available

	if PoToken != "" {

		PlaylistURI = POToken.ApplyToHLSManifestURL(PlaylistURI, PoToken)

	}

	var ProxyStruct *Structs.Proxy

	if Options.Proxy != nil {

		ProxyStruct = &Structs.Proxy{

			Host:     Options.Proxy.Host,
			Port:     Options.Proxy.Port,
			UserPass: Options.Proxy.UserPass,

		}

	}

	Content, Err := Functions.FetchHLSContent(PlaylistURI, ProxyStruct, Options.UserAgent)

	if Err != nil {

		return nil, fmt.Errorf("failed to fetch HLS playlist: %v", Err)

	}

	Playlist := Functions.ParseMediaPlaylist(Content, PlaylistURI)

	return Playlist, nil

}

// GetHLSSegment fetches raw bytes from an HLS segment URI
// For segment URLs, the PO Token is applied as a query parameter (?pot=<token>)
// rather than a path component
// If auto-generation is configured, will generate token if needed

func GetHLSSegment(SegmentURI string, Options *HLSOptions) ([]byte, error) {

	if Options == nil {

		Options = &HLSOptions{

			UserAgent: Config.Current.GetInnertubeClient().UserAgent,

		}

	}

	if Options.UserAgent == "" {

		Options.UserAgent = Config.Current.GetInnertubeClient().UserAgent

	}

	// Get or generate PO token

	PoToken, Err := getOrGeneratePoToken(Options)

	if Err != nil {

		return nil, fmt.Errorf("failed to get PO token: %v", Err)

	}

	// Apply PO Token to segment URL as query parameter

	if PoToken != "" {

		SegmentURI = POToken.ApplyToSegmentURL(SegmentURI, PoToken)

	}

	var ProxyStruct *Structs.Proxy

	if Options.Proxy != nil {

		ProxyStruct = &Structs.Proxy{

			Host:     Options.Proxy.Host,
			Port:     Options.Proxy.Port,
			UserPass: Options.Proxy.UserPass,

		}

	}

	Bytes, Err := Functions.FetchHLSSegmentBytes(SegmentURI, ProxyStruct, Options.UserAgent)

	if Err != nil {

		return nil, fmt.Errorf("failed to fetch HLS segment: %v", Err)

	}

	return Bytes, nil

}

// ApplyPoTokenToManifestURL applies a PO Token to an HLS or DASH manifest URL
// This is useful when you want to manually construct URLs with PO tokens

func ApplyPoTokenToManifestURL(ManifestURL string, PoToken string) string {

	return POToken.ApplyToHLSManifestURL(ManifestURL, PoToken)

}

// ApplyPoTokenToSegmentURL applies a PO Token to a segment URL as a query parameter
// This is useful when you want to manually construct segment URLs with PO tokens

func ApplyPoTokenToSegmentURL(SegmentURL string, PoToken string) string {

	return POToken.ApplyToSegmentURL(SegmentURL, PoToken)

}

// CleanPoToken validates and normalizes a PO Token
// Ensures the token is properly base64url encoded

func CleanPoToken(Token string) string {

	return POToken.CleanPoToken(Token)

}
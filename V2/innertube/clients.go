package innertube

import "github.com/elucid503/overture-play/v2/types"

// Predefined client configurations based on yt-dlp reference
// These are the clients that work best for avoiding 403 errors

var (
	// Web is the standard web browser client
	Web = ClientConfig{
		Name:        "WEB",
		Version:     "2.20250925.01.00",
		Host:        "www.youtube.com",
		ContextName: 1,

		RequireJSPlayer:           true,
		SupportsCookies:           true,
		SupportsAdPlaybackContext: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolDASH:  {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolHLS:   {Required: false, Recommended: true},
		},
		PlayerPoTokenPolicy: types.PoTokenPolicy{Required: false},
		SubsPoTokenPolicy:   types.PoTokenPolicy{Required: false},
	}

	// WebSafari returns HLS formats with pre-merged video+audio
	WebSafari = ClientConfig{
		Name:        "WEB",
		Version:     "2.20250925.01.00",
		Host:        "www.youtube.com",
		ContextName: 1,
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15,gzip(gfe)",

		RequireJSPlayer:           true,
		SupportsCookies:           true,
		SupportsAdPlaybackContext: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolDASH:  {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolHLS:   {Required: false, Recommended: true},
		},
		PlayerPoTokenPolicy: types.PoTokenPolicy{Required: false},
		SubsPoTokenPolicy:   types.PoTokenPolicy{Required: false},
	}

	// WebEmbedded is for embedded player context
	WebEmbedded = ClientConfig{
		Name:        "WEB_EMBEDDED_PLAYER",
		Version:     "1.20250923.21.00",
		Host:        "www.youtube.com",
		ContextName: 56,

		RequireJSPlayer: true,
		SupportsCookies: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: false},
			types.ProtocolDASH:  {Required: false},
			types.ProtocolHLS:   {Required: false},
		},
	}

	// WebMusic is YouTube Music client
	WebMusic = ClientConfig{
		Name:        "WEB_REMIX",
		Version:     "1.20250922.03.00",
		Host:        "music.youtube.com",
		ContextName: 67,

		RequireJSPlayer:           true,
		SupportsCookies:           true,
		SupportsAdPlaybackContext: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolDASH:  {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolHLS:   {Required: false, Recommended: true},
		},
	}

	// WebCreator requires authentication
	WebCreator = ClientConfig{
		Name:        "WEB_CREATOR",
		Version:     "1.20250922.03.00",
		Host:        "www.youtube.com",
		ContextName: 62,

		RequireJSPlayer: true,
		RequireAuth:     true,
		SupportsCookies: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolDASH:  {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolHLS:   {Required: false, Recommended: true},
		},
	}

	// Android is the Android app client
	Android = ClientConfig{
		Name:        "ANDROID",
		Version:     "20.10.38",
		Host:        "www.youtube.com",
		ContextName: 3,
		UserAgent:   "com.google.android.youtube/20.10.38 (Linux; U; Android 11) gzip",
		OSName:      "Android",
		OSVersion:   "11",

		RequireJSPlayer: false,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true, NotRequiredWithPlayerToken: true},
			types.ProtocolDASH:  {Required: true, Recommended: true, NotRequiredWithPlayerToken: true},
			types.ProtocolHLS:   {Required: false, Recommended: true, NotRequiredWithPlayerToken: true},
		},
		PlayerPoTokenPolicy: types.PoTokenPolicy{Required: false, Recommended: true},
	}

	// AndroidSDKLess doesn't require PO Token (useful fallback)
	AndroidSDKLess = ClientConfig{
		Name:        "ANDROID",
		Version:     "20.10.38",
		Host:        "www.youtube.com",
		ContextName: 3,
		UserAgent:   "com.google.android.youtube/20.10.38 (Linux; U; Android 11) gzip",
		OSName:      "Android",
		OSVersion:   "11",

		RequireJSPlayer: false,

		// No PO token policies - this client doesn't require them
		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: false},
			types.ProtocolDASH:  {Required: false},
			types.ProtocolHLS:   {Required: false},
		},
	}

	// AndroidVR is Oculus Quest client (doesn't return Kids videos)
	AndroidVR = ClientConfig{
		Name:        "ANDROID_VR",
		Version:     "1.65.10",
		Host:        "www.youtube.com",
		ContextName: 28,
		UserAgent:   "com.google.android.apps.youtube.vr.oculus/1.65.10 (Linux; U; Android 12L; eureka-user Build/SQ3A.220605.009.A1) gzip",
		DeviceMake:  "Oculus",
		DeviceModel: "Quest 3",
		OSName:      "Android",
		OSVersion:   "12L",

		RequireJSPlayer: false,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: false},
			types.ProtocolDASH:  {Required: false},
			types.ProtocolHLS:   {Required: false},
		},
	}

	// IOS is the iOS app client (provides HLS live streams)
	IOS = ClientConfig{
		Name:        "IOS",
		Version:     "20.10.4",
		Host:        "www.youtube.com",
		ContextName: 5,
		UserAgent:   "com.google.ios.youtube/20.10.4 (iPhone16,2; U; CPU iOS 18_3_2 like Mac OS X;)",
		DeviceMake:  "Apple",
		DeviceModel: "iPhone16,2",
		OSName:      "iPhone",
		OSVersion:   "18.3.2.22D82",

		RequireJSPlayer: false,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true, NotRequiredWithPlayerToken: true},
			types.ProtocolHLS:   {Required: true, Recommended: true, NotRequiredWithPlayerToken: true},
		},
		PlayerPoTokenPolicy: types.PoTokenPolicy{Required: false, Recommended: true},
	}

	// MWeb is the mobile web client (has ultralow formats)
	MWeb = ClientConfig{
		Name:        "MWEB",
		Version:     "2.20250925.01.00",
		Host:        "www.youtube.com",
		ContextName: 2,
		UserAgent:   "Mozilla/5.0 (iPad; CPU OS 16_7_10 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1,gzip(gfe)",

		RequireJSPlayer:           true,
		SupportsCookies:           true,
		SupportsAdPlaybackContext: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolDASH:  {Required: true, Recommended: true, NotRequiredForPremium: true},
			types.ProtocolHLS:   {Required: false, Recommended: true},
		},
	}

	// TV is the smart TV client
	TV = ClientConfig{
		Name:        "TVHTML5",
		Version:     "7.20250923.13.00",
		Host:        "www.youtube.com",
		ContextName: 7,
		UserAgent:   "Mozilla/5.0 (ChromiumStylePlatform) Cobalt/Version",

		RequireJSPlayer: true,
		SupportsCookies: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: false},
			types.ProtocolDASH:  {Required: false},
			types.ProtocolHLS:   {Required: false},
		},
	}

	// TVDowngraded is the older TV client (works better for some videos)
	TVDowngraded = ClientConfig{
		Name:        "TVHTML5",
		Version:     "5.20251105",
		Host:        "www.youtube.com",
		ContextName: 7,
		UserAgent:   "Mozilla/5.0 (ChromiumStylePlatform) Cobalt/Version",

		RequireJSPlayer: true,
		SupportsCookies: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: false},
			types.ProtocolDASH:  {Required: false},
			types.ProtocolHLS:   {Required: false},
		},
	}

	// TVSimply is a simplified TV client
	TVSimply = ClientConfig{
		Name:        "TVHTML5_SIMPLY",
		Version:     "1.0",
		Host:        "www.youtube.com",
		ContextName: 75,

		RequireJSPlayer: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: true, Recommended: true},
			types.ProtocolDASH:  {Required: true, Recommended: true},
			types.ProtocolHLS:   {Required: false, Recommended: true},
		},
	}

	// TVEmbedded is for TV embedded player (requires auth)
	TVEmbedded = ClientConfig{
		Name:        "TVHTML5_SIMPLY_EMBEDDED_PLAYER",
		Version:     "2.0",
		Host:        "www.youtube.com",
		ContextName: 85,

		RequireJSPlayer: true,
		RequireAuth:     true,
		SupportsCookies: true,

		GVSPoTokenPolicies: map[types.StreamingProtocol]types.PoTokenPolicy{
			types.ProtocolHTTPS: {Required: false},
			types.ProtocolDASH:  {Required: false},
			types.ProtocolHLS:   {Required: false},
		},
	}
)

// DefaultClients returns the recommended client order for unauthenticated users
func DefaultClients() []ClientConfig {
	return []ClientConfig{TV, AndroidSDKLess, Web}
}

// DefaultWebClients returns web-based clients
func DefaultWebClients() []ClientConfig {
	return []ClientConfig{Web, WebSafari, MWeb}
}

// DefaultAndroidClients returns Android clients
func DefaultAndroidClients() []ClientConfig {
	return []ClientConfig{AndroidSDKLess, Android, AndroidVR}
}

// DefaultAuthenticatedClients returns the recommended client order for authenticated users
func DefaultAuthenticatedClients() []ClientConfig {
	return []ClientConfig{TVDowngraded, WebSafari, Web}
}

// DefaultPremiumClients returns the recommended client order for premium subscribers
func DefaultPremiumClients() []ClientConfig {
	return []ClientConfig{TVDowngraded, WebCreator, Web}
}

// GetClientByName returns a client config by name
func GetClientByName(name string) *ClientConfig {
	clients := map[string]*ClientConfig{
		"web":           &Web,
		"web_safari":    &WebSafari,
		"web_embedded":  &WebEmbedded,
		"web_music":     &WebMusic,
		"web_creator":   &WebCreator,
		"android":       &Android,
		"android_sdkless": &AndroidSDKLess,
		"android_vr":    &AndroidVR,
		"ios":           &IOS,
		"mweb":          &MWeb,
		"tv":            &TV,
		"tv_downgraded": &TVDowngraded,
		"tv_simply":     &TVSimply,
		"tv_embedded":   &TVEmbedded,
	}

	if c, ok := clients[name]; ok {
		return c
	}
	return nil
}
